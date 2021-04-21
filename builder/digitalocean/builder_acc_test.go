package digitalocean

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"golang.org/x/oauth2"
)

func TestBuilderAcc_basic(t *testing.T) {
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-digitalocean-builder-basic",
		Template: fmt.Sprintf(testBuilderAccBasic, "ubuntu-20-04-x64"),
	})
}

func TestBuilderAcc_imageId(t *testing.T) {
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-digitalocean-builder-imageID",
		Template: makeTemplateWithImageId(t),
	})
}

func testAccPreCheck(t *testing.T) bool {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			acctest.TestEnvVar))
		return true
	}
	if v := os.Getenv("DIGITALOCEAN_API_TOKEN"); v == "" {
		t.Fatal("DIGITALOCEAN_API_TOKEN must be set for acceptance tests")
		return true
	}
	return false
}

func makeTemplateWithImageId(t *testing.T) string {
	if os.Getenv(acctest.TestEnvVar) != "" {
		token := os.Getenv("DIGITALOCEAN_API_TOKEN")
		client := godo.NewClient(oauth2.NewClient(context.TODO(), &apiTokenSource{
			AccessToken: token,
		}))
		image, _, err := client.Images.GetBySlug(context.TODO(), "ubuntu-20-04-x64")
		if err != nil {
			t.Fatalf("failed to retrieve image ID: %s", err)
		}

		return fmt.Sprintf(testBuilderAccBasic, image.ID)
	}

	return ""
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "digitalocean",
		"region": "nyc2",
		"size": "s-1vcpu-1gb",
		"image": "%v",
		"ssh_username": "root",
		"user_data": "",
		"user_data_file": ""
	}]
}
`
