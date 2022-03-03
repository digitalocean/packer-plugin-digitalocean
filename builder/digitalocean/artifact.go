package digitalocean

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

type Artifact struct {
	// The name of the snapshot
	SnapshotName string

	// The ID of the image
	SnapshotId int

	// The name of the region
	RegionNames []string

	// The client for making API calls
	Client *godo.Client

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

var _ packersdk.Artifact = new(Artifact)

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with DigitalOcean
	return nil
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s:%s", strings.Join(a.RegionNames[:], ","), strconv.FormatUint(uint64(a.SnapshotId), 10))
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: '%v' (ID: %v) in regions '%v'", a.SnapshotName, a.SnapshotId, strings.Join(a.RegionNames[:], ","))
}

func (a *Artifact) State(name string) interface{} {
	if name == registryimage.ArtifactStateURI {
		return a.stateHCPPackerRegistryMetadata()
	}
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d (%s)", a.SnapshotId, a.SnapshotName)
	_, err := a.Client.Images.Delete(context.TODO(), a.SnapshotId)
	return err
}

func (a *Artifact) stateHCPPackerRegistryMetadata() interface{} {
	// declare slice of images to be filled by the loop
	images := make([]*registryimage.Image, 0, len(a.RegionNames))
	// iterate over the regions names and create image metadata for each
	for _, region := range a.RegionNames {
		labels := make(map[string]string)
		var sourceID string
		var drpSize string
		var drpName string

		// Get and set the source image ID
		sourceID, ok := a.StateData["source_image_id"].(string)
		if ok {
			labels["source_image_id"] = sourceID
		}
		// Get and set the region information
		buildRegion, ok := a.StateData["build_region"].(string)
		if ok {
			labels["build_region"] = buildRegion
		}
		// Get and set droplet size
		drpSize, ok = a.StateData["droplet_size"].(string)
		if ok {
			labels["droplet_size"] = drpSize
		}
		// Get and set droplet name
		drpName, ok = a.StateData["droplet_name"].(string)
		if ok {
			labels["droplet_name"] = drpName
		}
		// instantiate the image
		img, err := registryimage.FromArtifact(a,
			registryimage.WithSourceID(sourceID),
			registryimage.WithID(a.SnapshotName),
			registryimage.WithProvider("DigitalOcean"),
			registryimage.WithRegion(region),
		)
		if err != nil {
			log.Printf("[DEBUG] error encountered when creating registry image %s", err)
			return nil
		}

		// Set labels
		img.Labels = labels
		images = append(images, img)
	}
	return images
}
