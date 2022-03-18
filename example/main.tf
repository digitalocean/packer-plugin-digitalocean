terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.18.0"
    }
    hcp = {
      source  = "hashicorp/hcp"
      version = "0.24.0"
    }
  }
}

// either use token below or set DIGITALOCEAN_TOKEN ENV VAR
provider "digitalocean" {
  token = "YOUR DIGITALOCEAN TOKEN"
}

data "hcp_packer_iteration" "production_digitalocean" {
  bucket_name = "digitalocean-hcp-test"
  channel     = "production"
}

data "hcp_packer_image" "production_digitalocean_image" {
  bucket_name    = "digitalocean-hcp-test"
  cloud_provider = "digitalocean"
  iteration_id   = data.hcp_packer_iteration.production_digitalocean.ulid
  region         = "nyc1"
}

resource "digitalocean_droplet" "production_digitalocean_droplet" {
  image  = data.hcp_packer_image.production_digitalocean_image.cloud_image_id
  name   = "prod-digitalocean-droplet"
  region = "nyc1"
  size   = "s-1vcpu-1gb"
}