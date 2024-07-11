package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"golang.org/x/sync/errgroup"
)

type stepSnapshot struct {
	snapshotTimeout         time.Duration
	transferTimeout         time.Duration
	waitForSnapshotTransfer bool
}

func (s *stepSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	dropletId := state.Get("droplet_id").(int)
	var snapshotRegions []string

	ui.Say(fmt.Sprintf("Creating snapshot: %v", c.SnapshotName))
	action, _, err := client.DropletActions.Snapshot(context.TODO(), dropletId, c.SnapshotName)
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// With the pending state over, verify that we're in the active state
	// because action can take a long time and may depend on the size of the final snapshot,
	// the timeout is parameterized
	ui.Say("Waiting for snapshot to complete...")
	if err := waitForActionState(godo.ActionCompleted, dropletId, action.ID,
		client, s.snapshotTimeout); err != nil {
		// If we get an error the first time, actually report it
		err := fmt.Errorf("Error waiting for snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the droplet to become unlocked first. For snapshots
	// this can end up taking quite a long time, so we hardcode this to
	// 20 minutes.
	if err := waitForDropletUnlocked(client, dropletId, 20*time.Minute); err != nil {
		// If we get an error the first time, actually report it
		err := fmt.Errorf("Error shutting down droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Looking up snapshot ID for snapshot: %s", c.SnapshotName)
	images, _, err := client.Droplets.Snapshots(context.TODO(), dropletId, nil)
	if err != nil {
		err := fmt.Errorf("Error looking up snapshot ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var imageId int
	if len(images) == 1 {
		imageId = images[0].ID
		log.Printf("Snapshot image ID: %d", imageId)
	} else {
		err := errors.New("Couldn't find snapshot to get the image ID. Bug?")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(c.SnapshotTags) > 0 {
		for _, tag := range c.SnapshotTags {
			_, err = client.Tags.TagResources(context.TODO(), tag, &godo.TagResourcesRequest{Resources: []godo.Resource{{ID: strconv.Itoa(imageId), Type: "image"}}})
			if err != nil {
				err := fmt.Errorf("Error Tagging Image: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			ui.Say(fmt.Sprintf("Added snapshot tag: %s...", tag))
		}
	}

	if len(c.SnapshotRegions) > 0 {
		regionSet := make(map[string]bool)
		regions := make([]string, 0, len(c.SnapshotRegions))
		regionSet[c.Region] = true
		for _, region := range c.SnapshotRegions {
			// If we already saw the region, then don't look again
			if regionSet[region] {
				continue
			}

			// Mark that we saw the region
			regionSet[region] = true

			regions = append(regions, region)
		}

		eg, gCtx := errgroup.WithContext(ctx)
		for _, r := range regions {
			region := r
			eg.Go(func() error {
				transferRequest := &godo.ActionRequest{
					"type":   "transfer",
					"region": region,
				}

				ui.Say(fmt.Sprintf("Transferring snapshot (ID: %d) to %s...", imageId, region))
				imageTransfer, _, err := client.ImageActions.Transfer(gCtx, imageId, transferRequest)
				if err != nil {
					return fmt.Errorf("Error transferring snapshot: %s", err)
				}

				if s.waitForSnapshotTransfer {
					if err := WaitForImageState(
						godo.ActionCompleted,
						imageId,
						imageTransfer.ID,
						client, s.transferTimeout); err != nil {
						return fmt.Errorf("Error waiting for snapshot transfer: %s", err)
					}
					ui.Say(fmt.Sprintf("Transfer to %s is complete.", region))
				}

				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	snapshotRegions = append(snapshotRegions, c.Region)

	state.Put("snapshot_image_id", imageId)
	state.Put("snapshot_name", c.SnapshotName)
	state.Put("regions", snapshotRegions)

	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state multistep.StateBag) {
	// no cleanup
}
