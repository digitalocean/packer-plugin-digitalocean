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
	var sourceID string
	var region string
	sourceImageID, ok := a.StateData["source_image_id"].(string)
	if ok {
		sourceID = sourceImageID
	}
	reg, ok := a.StateData["region"].(string)
	if ok {
		region = reg
	}
	img, _ := registryimage.FromArtifact(a,
		registryimage.WithSourceID(sourceID),
		registryimage.WithID(a.SnapshotName),
		registryimage.WithProvider("DigitalOcean"),
		registryimage.WithRegion(region),
	)
	return img
}
