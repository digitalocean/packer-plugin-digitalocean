Type: `digitalocean-image`

The DigitalOcean image data source is used look up the ID of an existing DigitalOcean image
for use as a builder source.

## Required:

<!-- Code generated from the comments of the Config struct in datasource/image/data.go; DO NOT EDIT MANUALLY -->

- `api_token` (string) - The API token to used to access your account. It can also be specified via
  the DIGITALOCEAN_TOKEN or DIGITALOCEAN_ACCESS_TOKEN environment variables.

<!-- End of code generated from the comments of the Config struct in datasource/image/data.go; -->


## Optional:

<!-- Code generated from the comments of the Config struct in datasource/image/data.go; DO NOT EDIT MANUALLY -->

- `api_url` (string) - A non-standard API endpoint URL. Set this if you are  using a DigitalOcean API
  compatible service. It can also be specified via environment variable DIGITALOCEAN_API_URL.

- `http_retry_max` (\*int) - The maximum number of retries for requests that fail with a 429 or 500-level error.
  The default value is 5. Set to 0 to disable reties.

- `http_retry_wait_max` (\*float64) - The maximum wait time (in seconds) between failed API requests. Default: 30.0

- `http_retry_wait_min` (\*float64) - The minimum wait time (in seconds) between failed API requests. Default: 1.0

- `name` (string) - The name of the image to return. Only one of `name` or `name_regex` may be provided.

- `name_regex` (string) - A regex matching the name of the image to return. Only one of `name` or `name_regex` may be provided.

- `type` (string) - Filter the images searched by type. This may be one of `application`, `distribution`, or `user`.
  By default, all image types are searched.

- `region` (string) - A DigitalOcean region slug (e.g. `nyc3`). When provided, only images available in that region
  will be returned.

- `latest` (bool) - A boolean value determining how to handle multiple matching images. By default, multiple matching images
  results in an error. When set to `true`, the most recently created image is returned instead.

<!-- End of code generated from the comments of the Config struct in datasource/image/data.go; -->


## Output:

<!-- Code generated from the comments of the DatasourceOutput struct in datasource/image/data.go; DO NOT EDIT MANUALLY -->

- `image_id` (int) - The ID of the found image.

- `image_regions` ([]string) - The regions the found image is availble in.

<!-- End of code generated from the comments of the DatasourceOutput struct in datasource/image/data.go; -->


## Example Usage

```hcl
data "digitalocean-image" "example" {
    name_regex = "golden-image-2022.*"
    region     = "nyc3"
    type       = "user"
    latest     = true
}

locals {
    image_id = data.digitalocean-image.example.image_id
}

source "digitalocean" "example" {
    snapshot_name = "updated-golden-image"
    image         = local.image_id
    region        = "nyc3"
    size          = "s-1vcpu-1gb"
    ssh_username  = "root"
}

build {
  sources = ["source.digitalocean.example"]
  provisioner "shell" {
    inline = ["touch /root/provisioned-by-packer"]
  }
}
```
