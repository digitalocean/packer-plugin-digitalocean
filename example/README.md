# Examples

## Simple Packer Build

After cloning this repo, move to the `example` directory by running:

```shell
cd packer-plugin-digitalocean/example
```

Either modify `basic_digitalocean.pkr.hcl` to reflect your DigitalOcean token or comment that out and set the `DIGITALOCEAN_TOKEN` environment variable by running:

```shell
export DIGITALOCEAN_TOKEN=<your digitalocean token>
```

Then run the following commands to build a simple DigitalOcean image via Packer:

```shell
packer init basic_digitalocean.pkr.hcl
packer build basic_digitalocean.pkr.hcl
```

## HCP Packer Build

### Prerequisites

- In order to complete this example, you must have Packer and Terraform installed.

HCP Packer gives users the ability to store metadata about their Packer builds and have Terraform consume the image IDs from HCP Packer to deploy virtual machines with that specific image. Learn more about HCP Packer [here](https://cloud.hashicorp.com/docs/packer).

After cloning this repo move to the `example` directory by running:

```shell
cd packer-plugin-digitalocean/example
```

HCP Packer requires a HCP account and the creation of a service principal. Documentation on creating a service principal can be found [here](https://cloud.hashicorp.com/docs/hcp/access-control/service-principals).

Once your service principal is created, either modify `hcp_packer_digitalocean.pkr.hcl` to reflect your DigitalOcean token or comment those lines out and run the following to set the `DIGITALOCEAN_TOKEN`:

```shell
export DIGITALOCEAN_TOKEN=<your digitalocean token>
```

Then run the following commands to build a simple DigitalOcean image via Packer and store the metadata about that image in HCP:

```shell
packer init hcp_packer_digitalocean.pkr.hcl
packer build hcp_packer_digitalocean.pkr.hcl
```

Now the metadata about your DigitalOcean image is stored within HCP Packer.

### Deploy DigitalOcean Droplet with Terraform

Navigate to your HCP Packer bucket `digitalocean-hcp-test` in a web browser.

Under `channels` create a new channel named `production` and assign the most recent iteration to it.

Back in your terminal navigate to the `example` directory by running:
```shell
cd packer-plugin-digitalocean/example
```

Modify the `digitalocean` and `hcp` provider blocks within the Terraform file `main.tf` with DigitalOcean Token, HCP Client ID, and HCP Client Secret or comment those lines out and run the following to set the `DIGITALOCEAN_TOKEN`, `HCP_CLIENT_ID`, and `HCP_CLIENT_SECRET` environment variables:

```shell
export DIGITALOCEAN_TOKEN=<your digitalocean token>
export HCP_CLIENT_ID=<your HCP client ID>
export HCP_CLIENT_SECRET=<your HCP client secret>
```

In order to deploy the DigitalOcean instance using the image based on HCP Packer metadata run the following commands:

```shell
terraform init
terraform apply
```