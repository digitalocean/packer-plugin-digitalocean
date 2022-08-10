package digitalocean

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"io/ioutil"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateDroplet struct {
	dropletId int
}

func (s *stepCreateDroplet) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	// Store the source image ID and
	// other miscellaneous info for HCP Packer
	state.Put("source_image_id", c.Image)
	state.Put("droplet_size", c.Size)
	state.Put("droplet_name", c.DropletName)
	state.Put("build_region", c.Region)

	// Create the droplet based on configuration
	ui.Say("Creating droplet...")
	dropletCreateReq, err := s.buildDropletCreateRequest(state)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	log.Printf("[DEBUG] Droplet create parameters: %s", godo.Stringify(dropletCreateReq))

	droplet, _, err := client.Droplets.Create(context.TODO(), dropletCreateReq)
	if err != nil {
		err := fmt.Errorf("Error creating droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this in cleanup
	s.dropletId = droplet.ID

	// Store the droplet id for later
	state.Put("droplet_id", droplet.ID)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", droplet.ID)

	return multistep.ActionContinue
}

func (s *stepCreateDroplet) buildDropletCreateRequest(state multistep.StateBag) (*godo.DropletCreateRequest, error) {
	c := state.Get("config").(*Config)

	sshKeys := []godo.DropletCreateSSHKey{}
	sshKeyID, hasSSHkey := state.GetOk("ssh_key_id")
	if hasSSHkey {
		sshKeys = append(sshKeys, godo.DropletCreateSSHKey{
			ID: sshKeyID.(int),
		})
	}
	if c.SSHKeyID != 0 {
		sshKeys = append(sshKeys, godo.DropletCreateSSHKey{
			ID: c.SSHKeyID,
		})
	}

	userData := c.UserData
	if c.UserDataFile != "" {
		contents, err := ioutil.ReadFile(c.UserDataFile)
		if err != nil {
			return nil, fmt.Errorf("Problem reading user data file: %s", err)
		}

		userData = string(contents)
	}

	createImage := getImageType(c.Image)

	return &godo.DropletCreateRequest{
		Name:              c.DropletName,
		Region:            c.Region,
		Size:              c.Size,
		Image:             createImage,
		SSHKeys:           sshKeys,
		PrivateNetworking: c.PrivateNetworking,
		Monitoring:        c.Monitoring,
		WithDropletAgent:  c.DropletAgent,
		IPv6:              c.IPv6,
		UserData:          userData,
		Tags:              c.Tags,
		VPCUUID:           c.VPCUUID,
	}, nil
}

func (s *stepCreateDroplet) Cleanup(state multistep.StateBag) {
	// If the dropletid isn't there, we probably never created it
	if s.dropletId == 0 {
		return
	}

	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packersdk.Ui)

	// Destroy the droplet we just created
	ui.Say("Destroying droplet...")
	_, err := client.Droplets.Delete(context.TODO(), s.dropletId)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying droplet. Please destroy it manually: %s", err))
	}
}

func getImageType(image string) godo.DropletCreateImage {
	createImage := godo.DropletCreateImage{Slug: image}

	imageId, err := strconv.Atoi(image)
	if err == nil {
		createImage = godo.DropletCreateImage{ID: imageId}
	}

	return createImage
}
