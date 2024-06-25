The [DigitalOcean](https://www.digitalocean.com/) Packer plugin provides a builder for building images in
DigitalOcean, and a post-processor for importing already-existing images into
DigitalOcean.


### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    digitalocean = {
      version = ">= 1.0.4"
      source  = "github.com/digitalocean/digitalocean"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/digitalocean/digitalocean
```

### Components

#### Builders

- [digitalocean](/packer/integrations/digitalocean/digitalocean/latest/components/builder/digitalocean) - The builder takes a source image, runs any provisioning necessary on the image after launching it, then snapshots it into a reusable image. This reusable image can then be used as the foundation of new servers that are launched within DigitalOcean.

#### Data Sources

- [digitalocean-image](/packer/integrations/digitalocean/digitalocean/latest/components/data-source/digitalocean-image) - The DigitalOcean image data source is used look up the ID of an existing DigitalOcean image for use as a builder source.

#### Post-processors

- [digitalocean-import](/packer/integrations/digitalocean/digitalocean/latest/components/post-processor/digitalocean-import) - The digitalocean-import post-processor is used to import images to DigitalOcean
