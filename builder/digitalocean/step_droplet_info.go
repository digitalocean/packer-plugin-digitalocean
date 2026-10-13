package digitalocean

import (
	"context"
	"errors"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepDropletInfo struct{}

func (s *stepDropletInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	dropletID := state.Get("droplet_id").(int)

	ui.Say("Waiting for droplet to become active...")

	err := waitForDropletState("active", dropletID, client, c.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for droplet to become active: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for droplet IPv4 address...")
	err = retryWithBackoff(
		ctx,
		defaultRetryTimeout,
		defaultRetryInitial,
		defaultRetryMaxBackoff,
		func(err error) bool {
			return errors.Is(err, errIPv4NotFound) || isNotFoundError(err)
		},
		func() error {
			droplet, _, getErr := client.Droplets.Get(ctx, dropletID)
			if getErr != nil {
				return getErr
			}
			ip, ipErr := dropletIPv4(droplet, c.ConnectWithPrivateIP)
			if ipErr != nil {
				return ipErr
			}
			state.Put("droplet_ip", ip)
			return nil
		},
	)
	if err != nil {
		err := fmt.Errorf("Error retrieving droplet IPv4 address: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDropletInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}
