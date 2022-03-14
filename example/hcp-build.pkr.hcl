packer {
  required_plugins {
    digitalocean = {
      source  = "github.com/hashicorp/digitalocean"
    }
  }
}

// Be sure to export your DIGITALOCEAN_TOKEN
// to your environment or use the below 'api_token'
// field.

source "digitalocean" "example" {
  api_token        = "YOUR API KEY"
  image            = "centos-stream-8-x64"
  region           = "nyc1"
  size             = "s-1vcpu-1gb"
  ssh_username     = "root"
  snapshot_regions = ["nyc1", "nyc3"]
}

build {
  hcp_packer_registry {
    bucket_name = "do-hcp-test"
    description = "A nice test description"
    bucket_labels = {
      "foo" = "bar"
    }
  }
  sources = ["source.digitalocean.example"]
}