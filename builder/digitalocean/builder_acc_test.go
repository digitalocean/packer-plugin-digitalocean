package digitalocean

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/packer-plugin-digitalocean/version"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/useragent"
	"golang.org/x/oauth2"
)

func TestBuilderAcc_basic(t *testing.T) {
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-digitalocean-builder-basic",
		Template: fmt.Sprintf(testBuilderAccBasic, "ubuntu-20-04-x64"),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_imageId(t *testing.T) {
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-digitalocean-builder-imageID",
		Template: makeTemplateWithImageId(t),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_multiRegion(t *testing.T) {
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-digitalocean-builder-multi-region",
		Template: fmt.Sprintf(testBuilderAccMultiRegion, "ubuntu-20-04-x64"),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_multiRegionNoWait(t *testing.T) {
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-digitalocean-builder-multi-region",
		Template: fmt.Sprintf(testBuilderAccMultiRegionNoWait, "ubuntu-20-04-x64"),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			logsBytes, err := io.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			logsString := string(logsBytes)

			notExpected := regexp.MustCompile(`Transfer to .* is complete.`)
			matches := notExpected.FindStringSubmatch(logsString)
			if len(matches) > 0 {
				return fmt.Errorf("logs contains unexpected value: %v", matches)
			}

			return nil
		},
	})
}

func testAccPreCheck(t *testing.T) bool {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", acctest.TestEnvVar)
		return true
	}
	v := os.Getenv("DIGITALOCEAN_TOKEN")
	if v == "" {
		v = os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	}
	if v == "" {
		v = os.Getenv("DIGITALOCEAN_API_TOKEN")
	}
	if v == "" {
		t.Fatal("DIGITALOCEAN_TOKEN or DIGITALOCEAN_ACCESS_TOKEN must be set for acceptance tests")
		return true
	}
	return false
}

func makeTemplateWithImageId(t *testing.T) string {
	if os.Getenv(acctest.TestEnvVar) != "" {
		token := os.Getenv("DIGITALOCEAN_TOKEN")
		if token == "" {
			token = os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
		}

		ua := useragent.String(version.PluginVersion.FormattedVersion())
		opts := []godo.ClientOpt{
			godo.SetUserAgent(ua),
			godo.WithRetryAndBackoffs(godo.RetryConfig{
				RetryMax: 5,
			}),
		}
		client, err := godo.New(oauth2.NewClient(context.TODO(), &APITokenSource{
			AccessToken: token,
		}), opts...)
		if err != nil {
			t.Fatalf("could not create client: %s", err)
		}

		image, _, err := client.Images.GetBySlug(context.TODO(), "ubuntu-20-04-x64")
		if err != nil {
			t.Fatalf("failed to retrieve image ID: %s", err)
		}

		return fmt.Sprintf(testBuilderAccBasic, image.ID)
	}

	return ""
}

const (
	testBuilderAccBasic = `
{
	"builders": [{
		"type": "digitalocean",
		"region": "nyc2",
		"size": "s-1vcpu-1gb",
		"image": "%v",
		"ssh_username": "root"
	}]
}
`

	testBuilderAccMultiRegion = `
{
	"builders": [{
		"type": "digitalocean",
		"region": "nyc2",
		"size": "s-1vcpu-1gb",
		"image": "%v",
		"ssh_username": "root",
		"snapshot_regions": ["nyc1", "nyc2", "nyc3"]
	}]
}
`

	testBuilderAccMultiRegionNoWait = `
{
	"builders": [{
		"type": "digitalocean",
		"region": "nyc2",
		"size": "s-1vcpu-1gb",
		"image": "%v",
		"ssh_username": "root",
		"snapshot_regions": ["nyc2", "nyc3"],
		"wait_snapshot_transfer": false
	}]
}
`
)
