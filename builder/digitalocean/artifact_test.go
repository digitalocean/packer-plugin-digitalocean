package digitalocean

import (
	"reflect"
	"testing"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/mitchellh/mapstructure"
)

func generatedData() map[string]interface{} {
	return make(map[string]interface{})
}

func TestArtifactId(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, []string{"sfo", "tor1"}, nil, generatedData()}
	expected := "sfo,tor1:42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactIdWithoutMultipleRegions(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, []string{"sfo"}, nil, generatedData()}
	expected := "sfo:42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, []string{"sfo", "tor1"}, nil, generatedData()}
	expected := "A snapshot was created: 'packer-foobar' (ID: 42) in regions 'sfo,tor1'"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}

func TestArtifactStringWithoutMultipleRegions(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, []string{"sfo"}, nil, generatedData()}
	expected := "A snapshot was created: 'packer-foobar' (ID: 42) in regions 'sfo'"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}

func TestArtifactState_StateData(t *testing.T) {
	expectedData := "this is the data"
	artifact := &Artifact{
		StateData: map[string]interface{}{"state_data": expectedData},
	}

	// Valid state
	result := artifact.State("state_data")
	if result != expectedData {
		t.Fatalf("Bad: State data was %s instead of %s", result, expectedData)
	}

	// Invalid state
	result = artifact.State("invalid_key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for invalid state data name")
	}

	// Nil StateData should not fail and should return nil
	artifact = &Artifact{}
	result = artifact.State("key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for nil StateData")
	}
}

func TestArtifactState_hcpPackerRegistryMetadata(t *testing.T) {
	regions := []string{"nyc1", "nyc3"}
	artifact := &Artifact{
		SnapshotName: "snapshot-1",
		SnapshotId:   12345,
		RegionNames:  regions,
		StateData:    map[string]interface{}{"source_image_id": "centos-stream-8-x64"},
	}
	// result should contain "something"
	result := artifact.State(registryimage.ArtifactStateURI)
	if result == nil {
		t.Fatalf("Bad: HCP Packer registry image data was nil")
	}

	// check for proper decoding of result into slice of registryimage.Image
	var images []registryimage.Image
	err := mapstructure.Decode(result, &images)
	if err != nil {
		t.Errorf("Bad: unexpected error when trying to decode state into registryimage.Image %v", err)
	}

	// check and make sure multi-region is working
	if len(images) != 2 {
		t.Errorf("Bad: we should have two images for this test Artifact but we got %d", len(images))
	}

	// check that all properties of the images were set correctly
	expected := []registryimage.Image{
		{
			ImageID:        "12345",
			ProviderName:   "digitalocean",
			ProviderRegion: "nyc1",
			SourceImageID:  "centos-stream-8-x64",
			Labels:         map[string]string{"source_image_id": "centos-stream-8-x64"},
		},
		{
			ImageID:        "12345",
			ProviderName:   "digitalocean",
			ProviderRegion: "nyc3",
			SourceImageID:  "centos-stream-8-x64",
			Labels:         map[string]string{"source_image_id": "centos-stream-8-x64"},
		},
	}
	if !reflect.DeepEqual(images, expected) {
		t.Fatalf("Bad: expected %#v got %#v", expected, images)
	}
}
