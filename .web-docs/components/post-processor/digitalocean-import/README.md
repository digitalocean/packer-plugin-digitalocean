Type: `digitalocean-import`
Artifact BuilderId: `packer.post-processor.digitalocean-import`

The Packer DigitalOcean Import post-processor is used to import images created by other Packer builders to DigitalOcean.

~> Note: Users looking to create custom images, and reusable snapshots, directly on DigitalOcean can use
the [DigitalOcean builder](/docs/builder/digitalocean) without this post-processor.

## How Does it Work?

The import process operates uploading a temporary copy of the image to
DigitalOcean Spaces and then importing it as a custom image via the
DigialOcean API. The temporary copy in Spaces can be discarded after the
import is complete.

For information about the requirements to use an image for a DigitalOcean
Droplet, see DigitalOcean's [Custom Images documentation](https://www.digitalocean.com/docs/images/custom-images).

## Configuration

There are some configuration options available for the post-processor.

Required:

<!-- Code generated from the comments of the Config struct in post-processor/digitalocean-import/post-processor.go; DO NOT EDIT MANUALLY -->

- `api_token` (string) - A personal access token used to communicate with the DigitalOcean v2 API.
  This may also be set using the `DIGITALOCEAN_TOKEN` or
  `DIGITALOCEAN_ACCESS_TOKEN` environmental variables.

- `spaces_key` (string) - The access key used to communicate with Spaces. This may also be set using
  the `DIGITALOCEAN_SPACES_ACCESS_KEY` environmental variable.

- `spaces_secret` (string) - The secret key used to communicate with Spaces. This may also be set using
  the `DIGITALOCEAN_SPACES_SECRET_KEY` environmental variable.

- `spaces_region` (string) - The name of the region, such as `nyc3`, in which to upload the image to Spaces.

- `space_name` (string) - The name of the specific Space where the image file will be copied to for
  import. This Space must exist when the post-processor is run.

- `image_name` (string) - The name to be used for the resulting DigitalOcean custom image.

- `image_regions` ([]string) - A list of DigitalOcean regions, such as `nyc3`, where the resulting image
  will be available for use in creating Droplets.

<!-- End of code generated from the comments of the Config struct in post-processor/digitalocean-import/post-processor.go; -->


Optional:

<!-- Code generated from the comments of the Config struct in post-processor/digitalocean-import/post-processor.go; DO NOT EDIT MANUALLY -->

- `api_token` (string) - A personal access token used to communicate with the DigitalOcean v2 API.
  This may also be set using the `DIGITALOCEAN_TOKEN` or
  `DIGITALOCEAN_ACCESS_TOKEN` environmental variables.

- `spaces_key` (string) - The access key used to communicate with Spaces. This may also be set using
  the `DIGITALOCEAN_SPACES_ACCESS_KEY` environmental variable.

- `spaces_secret` (string) - The secret key used to communicate with Spaces. This may also be set using
  the `DIGITALOCEAN_SPACES_SECRET_KEY` environmental variable.

- `spaces_region` (string) - The name of the region, such as `nyc3`, in which to upload the image to Spaces.

- `space_name` (string) - The name of the specific Space where the image file will be copied to for
  import. This Space must exist when the post-processor is run.

- `image_name` (string) - The name to be used for the resulting DigitalOcean custom image.

- `image_regions` ([]string) - A list of DigitalOcean regions, such as `nyc3`, where the resulting image
  will be available for use in creating Droplets.

<!-- End of code generated from the comments of the Config struct in post-processor/digitalocean-import/post-processor.go; -->


- `keep_input_artifact` (boolean) - if true, do not delete the source virtual
  machine image after importing it to the cloud. Defaults to false.

## Basic Example

Here is a basic example:

**JSON**

```json
{
  "type": "digitalocean-import",
  "api_token": "{{user `token`}}",
  "spaces_key": "{{user `key`}}",
  "spaces_secret": "{{user `secret`}}",
  "spaces_region": "nyc3",
  "space_name": "import-bucket",
  "image_name": "ubuntu-18.10-minimal-amd64",
  "image_description": "Packer import {{timestamp}}",
  "image_regions": ["nyc3", "nyc2"],
  "image_tags": ["custom", "packer"]
}
```

**HCL2**

```hcl
post-processor "digitalocean-import" {
  api_token         = "{{user `token`}}"
  spaces_key        = "{{user `key`}}"
  spaces_secret     = "{{user `secret`}}"
  spaces_region     = "nyc3"
  space_name        = "import-bucket"
  image_name        = "ubuntu-18.10-minimal-amd64"
  image_description = "Packer import {{timestamp}}"
  image_regions     = ["nyc3", "nyc2"]
  image_tags        = ["custom", "packer"]
}
```
