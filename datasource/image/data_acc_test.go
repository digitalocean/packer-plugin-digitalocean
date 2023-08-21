package image

import (
	"context"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/digitalocean/godo"
	builder "github.com/digitalocean/packer-plugin-digitalocean/builder/digitalocean"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"golang.org/x/oauth2"
)

func TestAccDatasource_Validations(t *testing.T) {
	// store to reset in Teardown
	doToken := os.Getenv("DIGITALOCEAN_TOKEN")
	doAccessToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")

	tests := []*acctest.PluginTestCase{
		{
			Name: "test missing required values",
			Setup: func() error {
				// unset to ensure failure of token check
				os.Unsetenv("DIGITALOCEAN_TOKEN")
				os.Unsetenv("DIGITALOCEAN_ACCESS_TOKEN")
				return nil
			},
			Teardown: func() error {
				os.Setenv("DIGITALOCEAN_TOKEN", doToken)
				os.Setenv("DIGITALOCEAN_ACCESS_TOKEN", doAccessToken)
				return nil
			},
			Template: `data "digitalocean-image" "test" {}`,
			Type:     "digitalocean-image",
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				tokenRequired := "api_token is required"
				nameOrRegex := "one of name or name_regex is required"

				err := findInTestLog(t, logfile, tokenRequired)
				if err != nil {
					return err
				}
				err = findInTestLog(t, logfile, nameOrRegex)
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name: "only one of name or name_regex can be set",
			Setup: func() error {
				return nil
			},
			Teardown: func() error {
				return nil
			},
			Template: `data "digitalocean-image" "test" {
				name = "foo"
				name_regex = "foo.*"
			}`,
			Type: "digitalocean-image",
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				nameOrRegex := "only one of name or name_regex can be set"
				err := findInTestLog(t, logfile, nameOrRegex)
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name: "invalid image type",
			Setup: func() error {
				return nil
			},
			Teardown: func() error {
				return nil
			},
			Template: `data "digitalocean-image" "test" {
				name = "foo"
				type = "1-click"
			}`,
			Type: "digitalocean-image",
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 1 {
						return fmt.Errorf("Unexpected exit code. Logfile: %s", logfile)
					}
				}

				invalid := `invalid type; must be one of`
				err := findInTestLog(t, logfile, invalid)
				if err != nil {
					return err
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			acctest.TestPlugin(t, tt)
		})
	}
}

func TestAccDatasource_Basic(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'PACKER_ACC' set")
	}

	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		t.Fatal("DIGITALOCEAN_TOKEN environment variable required")
	}
	oauthClient := oauth2.NewClient(context.TODO(), &builder.APITokenSource{
		AccessToken: token,
	})
	client, err := godo.New(oauthClient, godo.WithRetryAndBackoffs(godo.RetryConfig{
		RetryMax: 5,
	}))
	if err != nil {
		t.Fatalf("could not create client: %s", err)
	}

	file, err := os.CreateTemp("/tmp", "packer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	images, _, err := client.Images.ListApplication(context.TODO(), nil)
	if err != nil {
		t.Error(err)
	}
	expectedImageID := images[0].ID
	datsourceFixture := fmt.Sprintf(`
data "digitalocean-image" "test" {
	name = "%s"
	region = "nyc3"
	type  = "application"
}

locals {
	image_id = data.digitalocean-image.test.image_id
}

source "file" "basic-example" {
	content =  local.image_id
	target =  "%s"
}

build {
	sources = ["sources.file.basic-example"]
}`, images[0].Name, file.Name())

	testCase := &acctest.PluginTestCase{
		Name: "scaffolding_datasource_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: datsourceFixture,
		Type:     "digitalocean-image",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			imageLog := fmt.Sprintf("found image: %d", expectedImageID)
			err := findInTestLog(t, logfile, imageLog)
			if err != nil {
				return err
			}
			return nil
		},
	}

	acctest.TestPlugin(t, testCase)
}

func findInTestLog(t *testing.T, logfile string, expected string) error {
	logs, err := os.Open(logfile)
	if err != nil {
		return fmt.Errorf("Unable find %s", logfile)
	}
	defer logs.Close()

	logsBytes, err := ioutil.ReadAll(logs)
	if err != nil {
		return fmt.Errorf("Unable to read %s", logfile)
	}
	logsString := string(logsBytes)

	if matched, _ := regexp.MatchString(expected+".*", logsString); !matched {
		t.Fatalf("logs doesn't contain expected value %q", logsString)
	}

	return nil
}
