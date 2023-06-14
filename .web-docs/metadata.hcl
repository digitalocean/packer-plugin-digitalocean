# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "TODO"
  description = "TODO"
  identifier = "packer/BrandonRomano/digitalocean"
  component {
    type = "data-source"
    name = "DigitalOcean Image"
    slug = "digitalocen-image"
  }
  component {
    type = "builder"
    name = "DigitalOcean"
    slug = "digitalocean"
  }
  component {
    type = "post-processor"
    name = "DigitalOcean Import"
    slug = "digitalocean-import"
  }
}
