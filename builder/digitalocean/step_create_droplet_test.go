package digitalocean

import (
	"testing"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/require"
)

func TestBuilder_GetImageType(t *testing.T) {
	imageTypeTests := []struct {
		in  string
		out godo.DropletCreateImage
	}{
		{"ubuntu-20-04-x64", godo.DropletCreateImage{Slug: "ubuntu-20-04-x64"}},
		{"123456", godo.DropletCreateImage{ID: 123456}},
	}

	for _, tt := range imageTypeTests {
		t.Run(tt.in, func(t *testing.T) {
			i := getImageType(tt.in)
			if i != tt.out {
				t.Errorf("got %q, want %q", godo.Stringify(i), godo.Stringify(tt.out))
			}
		})
	}
}

func TestBuilder_buildDropletCreateRequest(t *testing.T) {
	imageTypeTests := []struct {
		name       string
		in         *Config
		out        *godo.DropletCreateRequest
		addToState map[string]interface{}
	}{
		{
			name: "DropletAgent is false",
			in: &Config{
				DropletName:  "ubuntu-20-04-x64-build",
				Region:       "nyc3",
				Size:         "s-1vcpu-1gb",
				Image:        "ubuntu-20-04-x64",
				DropletAgent: godo.Bool(false),
			},
			out: &godo.DropletCreateRequest{
				Name:              "ubuntu-20-04-x64-build",
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             godo.DropletCreateImage{ID: 0, Slug: "ubuntu-20-04-x64"},
				SSHKeys:           []godo.DropletCreateSSHKey{},
				Backups:           false,
				IPv6:              false,
				PrivateNetworking: false,
				Monitoring:        false,
				UserData:          "",
				VPCUUID:           "",
				WithDropletAgent:  godo.Bool(false),
			},
		},
		{
			name: "DropletAgent is true",
			in: &Config{
				DropletName:  "ubuntu-20-04-x64-build",
				Region:       "nyc3",
				Size:         "s-1vcpu-1gb",
				Image:        "ubuntu-20-04-x64",
				DropletAgent: godo.Bool(true),
			},
			out: &godo.DropletCreateRequest{
				Name:              "ubuntu-20-04-x64-build",
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             godo.DropletCreateImage{ID: 0, Slug: "ubuntu-20-04-x64"},
				SSHKeys:           []godo.DropletCreateSSHKey{},
				Backups:           false,
				IPv6:              false,
				PrivateNetworking: false,
				Monitoring:        false,
				UserData:          "",
				VPCUUID:           "",
				WithDropletAgent:  godo.Bool(true),
			},
		},
		{
			name: "DropletAgent is not set",
			in: &Config{
				DropletName: "ubuntu-20-04-x64-build",
				Region:      "nyc3",
				Size:        "s-1vcpu-1gb",
				Image:       "ubuntu-20-04-x64",
			},
			out: &godo.DropletCreateRequest{
				Name:              "ubuntu-20-04-x64-build",
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             godo.DropletCreateImage{ID: 0, Slug: "ubuntu-20-04-x64"},
				SSHKeys:           []godo.DropletCreateSSHKey{},
				Backups:           false,
				IPv6:              false,
				PrivateNetworking: false,
				Monitoring:        false,
				UserData:          "",
				VPCUUID:           "",
			},
		},
		{
			name: "SSHKeyID set in config",
			in: &Config{
				DropletName: "ubuntu-20-04-x64-build",
				Region:      "nyc3",
				Size:        "s-1vcpu-1gb",
				Image:       "ubuntu-20-04-x64",
				SSHKeyID:    12345,
			},
			out: &godo.DropletCreateRequest{
				Name:              "ubuntu-20-04-x64-build",
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             godo.DropletCreateImage{ID: 0, Slug: "ubuntu-20-04-x64"},
				SSHKeys:           []godo.DropletCreateSSHKey{{ID: 12345, Fingerprint: ""}},
				Backups:           false,
				IPv6:              false,
				PrivateNetworking: false,
				Monitoring:        false,
				UserData:          "",
				VPCUUID:           "",
			},
		},
		{
			name:       "SSH key set in both state and config",
			addToState: map[string]interface{}{"ssh_key_id": 56789},
			in: &Config{
				DropletName: "ubuntu-20-04-x64-build",
				Region:      "nyc3",
				Size:        "s-1vcpu-1gb",
				Image:       "ubuntu-20-04-x64",
				SSHKeyID:    123456,
			},
			out: &godo.DropletCreateRequest{
				Name:   "ubuntu-20-04-x64-build",
				Region: "nyc3",
				Size:   "s-1vcpu-1gb",
				Image:  godo.DropletCreateImage{ID: 0, Slug: "ubuntu-20-04-x64"},
				SSHKeys: []godo.DropletCreateSSHKey{
					{ID: 56789, Fingerprint: ""},
					{ID: 123456, Fingerprint: ""},
				},
				Backups:           false,
				IPv6:              false,
				PrivateNetworking: false,
				Monitoring:        false,
				UserData:          "",
				VPCUUID:           "",
			},
		},
		{
			name:       "ssh_key_id set in state",
			addToState: map[string]interface{}{"ssh_key_id": 56789},
			in: &Config{
				DropletName: "ubuntu-20-04-x64-build",
				Region:      "nyc3",
				Size:        "s-1vcpu-1gb",
				Image:       "ubuntu-20-04-x64",
			},
			out: &godo.DropletCreateRequest{
				Name:              "ubuntu-20-04-x64-build",
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             godo.DropletCreateImage{ID: 0, Slug: "ubuntu-20-04-x64"},
				SSHKeys:           []godo.DropletCreateSSHKey{{ID: 56789, Fingerprint: ""}},
				Backups:           false,
				IPv6:              false,
				PrivateNetworking: false,
				Monitoring:        false,
				UserData:          "",
				VPCUUID:           "",
			},
		},
		{
			name: "image as int",
			in: &Config{
				DropletName: "ubuntu-20-04-x64-build",
				Region:      "nyc3",
				Size:        "s-1vcpu-1gb",
				Image:       "789",
				SSHKeyID:    12345,
			},
			out: &godo.DropletCreateRequest{
				Name:              "ubuntu-20-04-x64-build",
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             godo.DropletCreateImage{ID: 789, Slug: ""},
				SSHKeys:           []godo.DropletCreateSSHKey{{ID: 12345, Fingerprint: ""}},
				Backups:           false,
				IPv6:              false,
				PrivateNetworking: false,
				Monitoring:        false,
				UserData:          "",
				VPCUUID:           "",
			},
		},
	}

	for _, tt := range imageTypeTests {
		t.Run(tt.name, func(t *testing.T) {
			state := new(multistep.BasicStateBag)
			state.Put("config", tt.in)
			if tt.addToState != nil {
				for k, v := range tt.addToState {
					state.Put(k, v)
				}
			}

			step := new(stepCreateDroplet)

			req, err := step.buildDropletCreateRequest(state)
			require.NoError(t, err)

			require.Equal(t, tt.out, req)
		})
	}
}
