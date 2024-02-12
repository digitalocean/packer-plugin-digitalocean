Type: `digitalocean`
Artifact BuilderId: `pearkes.digitalocean`

The `digitalocean` Packer builder is able to create new images for use with
[DigitalOcean](https://www.digitalocean.com). The builder takes a source image,
runs any provisioning necessary on the image after launching it, then snapshots
it into a reusable image. This reusable image can then be used as the
foundation of new servers that are launched within DigitalOcean.

The builder does _not_ manage images. Once it creates an image, it is up to you
to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

### Required:

<!-- Code generated from the comments of the Config struct in builder/digitalocean/config.go; DO NOT EDIT MANUALLY -->

- `api_token` (string) - The client TOKEN to use to access your account. It
  can also be specified via environment variable DIGITALOCEAN_TOKEN, DIGITALOCEAN_ACCESS_TOKEN, or DIGITALOCEAN_API_TOKEN if
  set. DIGITALOCEAN_API_TOKEN will be deprecated in a future release in favor of DIGITALOCEAN_TOKEN or DIGITALOCEAN_ACCESS_TOKEN.

- `region` (string) - The name (or slug) of the region to launch the droplet
  in. Consequently, this is the region where the snapshot will be available.
  See
  https://docs.digitalocean.com/reference/api/api-reference/#operation/list_all_regions
  for the accepted region names/slugs.

- `size` (string) - The name (or slug) of the droplet size to use. See
  https://docs.digitalocean.com/reference/api/api-reference/#operation/list_all_sizes
  for the accepted size names/slugs.

- `image` (string) - The name (or slug) of the base image to use. This is the
  image that will be used to launch a new droplet and provision it. See
  https://docs.digitalocean.com/reference/api/api-reference/#operation/get_images_list
  for details on how to get a list of the accepted image names/slugs.

<!-- End of code generated from the comments of the Config struct in builder/digitalocean/config.go; -->


### Optional:

<!-- Code generated from the comments of the Config struct in builder/digitalocean/config.go; DO NOT EDIT MANUALLY -->

- `api_url` (string) - Non standard api endpoint URL. Set this if you are
  using a DigitalOcean API compatible service. It can also be specified via
  environment variable DIGITALOCEAN_API_URL.

- `http_retry_max` (\*int) - The maximum number of retries for requests that fail with a 429 or 500-level error.
  The default value is 5. Set to 0 to disable reties.

- `http_retry_wait_max` (\*float64) - The maximum wait time (in seconds) between failed API requests. Default: 30.0

- `http_retry_wait_min` (\*float64) - The minimum wait time (in seconds) between failed API requests. Default: 1.0

- `private_networking` (bool) - Set to true to enable private networking
  for the droplet being created. This defaults to false, or not enabled.

- `monitoring` (bool) - Set to true to enable monitoring for the droplet
  being created. This defaults to false, or not enabled.

- `droplet_agent` (\*bool) - A boolean indicating whether to install the DigitalOcean agent used for
  providing access to the Droplet web console in the control panel. By
  default, the agent is installed on new Droplets but installation errors
  (i.e. OS not supported) are ignored. To prevent it from being installed,
  set to false. To make installation errors fatal, explicitly set it to true.

- `ipv6` (bool) - Set to true to enable ipv6 for the droplet being
  created. This defaults to false, or not enabled.

- `snapshot_name` (string) - The name of the resulting snapshot that will
  appear in your account. Defaults to `packer-{{timestamp}}` (see
  configuration templates for more info).

- `snapshot_regions` ([]string) - Additional regions that resulting snapshot should be distributed to.

- `wait_snapshot_transfer` (\*bool) - When true, Packer will block until all snapshot transfers have been completed
  and report errors. When false, Packer will initiate the snapshot transfers
  and exit successfully without waiting for completion. Defaults to true.

- `state_timeout` (duration string | ex: "1h5m2s") - The time to wait, as a duration string, for a
  droplet to enter a desired state (such as "active") before timing out. The
  default state timeout is "6m".

- `snapshot_timeout` (duration string | ex: "1h5m2s") - How long to wait for an image to be published to the shared image
  gallery before timing out. If your Packer build is failing on the
  Publishing to Shared Image Gallery step with the error `Original Error:
  context deadline exceeded`, but the image is present when you check your
  Azure dashboard, then you probably need to increase this timeout from
  its default of "60m" (valid time units include `s` for seconds, `m` for
  minutes, and `h` for hours.)

- `droplet_name` (string) - The name assigned to the droplet. DigitalOcean
  sets the hostname of the machine to this value.

- `user_data` (string) - User data to launch with the Droplet. Packer will
  not automatically wait for a user script to finish before shutting down the
  instance this must be handled in a provisioner.

- `user_data_file` (string) - Path to a file that will be used for the user
  data when launching the Droplet.

- `tags` ([]string) - Tags to apply to the droplet when it is created

- `vpc_uuid` (string) - UUID of the VPC which the droplet will be created in. Before using this,
  private_networking should be enabled.

- `connect_with_private_ip` (bool) - Wheter the communicators should use private IP or not (public IP in that case).
  If the droplet is or going to be accessible only from the local network because
  it is at behind a firewall, then communicators should use the private IP
  instead of the public IP. Before using this, private_networking should be enabled.

- `ssh_key_id` (int) - The ID of an existing SSH key on the DigitalOcean account. This should be
  used in conjunction with `ssh_private_key_file`.

<!-- End of code generated from the comments of the Config struct in builder/digitalocean/config.go; -->


## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

**HCL2**

```hcl
source "digitalocean" "example" {
  api_token    = "YOUR API KEY"
  image        = "ubuntu-22-04-x64"
  region       = "nyc3"
  size         = "s-1vcpu-1gb"
  ssh_username = "root"
}

build {
  sources = ["source.digitalocean.example"]
}
```

**JSON**

```json
{
  "type": "digitalocean",
  "api_token": "YOUR API KEY",
  "image": "ubuntu-22-04-x64",
  "region": "nyc3",
  "size": "s-1vcpu-1gb",
  "ssh_username": "root"
}
```


### Communicator Config

In addition to the builder options, a
[communicator](/docs/templates/legacy_json_templates/communicator) can be configured for this builder.

<!-- Code generated from the comments of the Config struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `communicator` (string) - Packer currently supports three kinds of communicators:
  
  -   `none` - No communicator will be used. If this is set, most
      provisioners also can't be used.
  
  -   `ssh` - An SSH connection will be established to the machine. This
      is usually the default.
  
  -   `winrm` - A WinRM connection will be established.
  
  In addition to the above, some builders have custom communicators they
  can use. For example, the Docker builder has a "docker" communicator
  that uses `docker exec` and `docker cp` to execute scripts and copy
  files.

- `pause_before_connecting` (duration string | ex: "1h5m2s") - We recommend that you enable SSH or WinRM as the very last step in your
  guest's bootstrap script, but sometimes you may have a race condition
  where you need Packer to wait before attempting to connect to your
  guest.
  
  If you end up in this situation, you can use the template option
  `pause_before_connecting`. By default, there is no pause. For example if
  you set `pause_before_connecting` to `10m` Packer will check whether it
  can connect, as normal. But once a connection attempt is successful, it
  will disconnect and then wait 10 minutes before connecting to the guest
  and beginning provisioning.

<!-- End of code generated from the comments of the Config struct in communicator/config.go; -->


<!-- Code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `temporary_key_pair_type` (string) - `dsa` | `ecdsa` | `ed25519` | `rsa` ( the default )
  
  Specifies the type of key to create. The possible values are 'dsa',
  'ecdsa', 'ed25519', or 'rsa'.
  
  NOTE: DSA is deprecated and no longer recognized as secure, please
  consider other alternatives like RSA or ED25519.

- `temporary_key_pair_bits` (int) - Specifies the number of bits in the key to create. For RSA keys, the
  minimum size is 1024 bits and the default is 4096 bits. Generally, 3072
  bits is considered sufficient. DSA keys must be exactly 1024 bits as
  specified by FIPS 186-2. For ECDSA keys, bits determines the key length
  by selecting from one of three elliptic curve sizes: 256, 384 or 521
  bits. Attempting to use bit lengths other than these three values for
  ECDSA keys will fail. Ed25519 keys have a fixed length and bits will be
  ignored.
  
  NOTE: DSA is deprecated and no longer recognized as secure as specified
  by FIPS 186-5, please consider other alternatives like RSA or ED25519.

<!-- End of code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; -->


<!-- Code generated from the comments of the SSH struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `ssh_host` (string) - The address to SSH to. This usually is automatically configured by the
  builder.

- `ssh_port` (int) - The port to connect to SSH. This defaults to `22`.

- `ssh_username` (string) - The username to connect to SSH with. Required if using SSH.

- `ssh_password` (string) - A plaintext password to use to authenticate with SSH.

- `ssh_ciphers` ([]string) - This overrides the value of ciphers supported by default by Golang.
  The default value is [
    "aes128-gcm@openssh.com",
    "chacha20-poly1305@openssh.com",
    "aes128-ctr", "aes192-ctr", "aes256-ctr",
  ]
  
  Valid options for ciphers include:
  "aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com",
  "chacha20-poly1305@openssh.com",
  "arcfour256", "arcfour128", "arcfour", "aes128-cbc", "3des-cbc",

- `ssh_clear_authorized_keys` (bool) - If true, Packer will attempt to remove its temporary key from
  `~/.ssh/authorized_keys` and `/root/.ssh/authorized_keys`. This is a
  mostly cosmetic option, since Packer will delete the temporary private
  key from the host system regardless of whether this is set to true
  (unless the user has set the `-debug` flag). Defaults to "false";
  currently only works on guests with `sed` installed.

- `ssh_key_exchange_algorithms` ([]string) - If set, Packer will override the value of key exchange (kex) algorithms
  supported by default by Golang. Acceptable values include:
  "curve25519-sha256@libssh.org", "ecdh-sha2-nistp256",
  "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
  "diffie-hellman-group14-sha1", and "diffie-hellman-group1-sha1".

- `ssh_certificate_file` (string) - Path to user certificate used to authenticate with SSH.
  The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_pty` (bool) - If `true`, a PTY will be requested for the SSH connection. This defaults
  to `false`.

- `ssh_timeout` (duration string | ex: "1h5m2s") - The time to wait for SSH to become available. Packer uses this to
  determine when the machine has booted so this is usually quite long.
  Example value: `10m`.
  This defaults to `5m`, unless `ssh_handshake_attempts` is set.

- `ssh_disable_agent_forwarding` (bool) - If true, SSH agent forwarding will be disabled. Defaults to `false`.

- `ssh_handshake_attempts` (int) - The number of handshakes to attempt with SSH once it can connect.
  This defaults to `10`, unless a `ssh_timeout` is set.

- `ssh_bastion_host` (string) - A bastion host to use for the actual SSH connection.

- `ssh_bastion_port` (int) - The port of the bastion host. Defaults to `22`.

- `ssh_bastion_agent_auth` (bool) - If `true`, the local SSH agent will be used to authenticate with the
  bastion host. Defaults to `false`.

- `ssh_bastion_username` (string) - The username to connect to the bastion host.

- `ssh_bastion_password` (string) - The password to use to authenticate with the bastion host.

- `ssh_bastion_interactive` (bool) - If `true`, the keyboard-interactive used to authenticate with bastion host.

- `ssh_bastion_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with the
  bastion host. The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_bastion_certificate_file` (string) - Path to user certificate used to authenticate with bastion host.
  The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_file_transfer_method` (string) - `scp` or `sftp` - How to transfer files, Secure copy (default) or SSH
  File Transfer Protocol.
  
  **NOTE**: Guests using Windows with Win32-OpenSSH v9.1.0.0p1-Beta, scp
  (the default protocol for copying data) returns a a non-zero error code since the MOTW
  cannot be set, which cause any file transfer to fail. As a workaround you can override the transfer protocol
  with SFTP instead `ssh_file_transfer_protocol = "sftp"`.

- `ssh_proxy_host` (string) - A SOCKS proxy host to use for SSH connection

- `ssh_proxy_port` (int) - A port of the SOCKS proxy. Defaults to `1080`.

- `ssh_proxy_username` (string) - The optional username to authenticate with the proxy server.

- `ssh_proxy_password` (string) - The optional password to use to authenticate with the proxy server.

- `ssh_keep_alive_interval` (duration string | ex: "1h5m2s") - How often to send "keep alive" messages to the server. Set to a negative
  value (`-1s`) to disable. Example value: `10s`. Defaults to `5s`.

- `ssh_read_write_timeout` (duration string | ex: "1h5m2s") - The amount of time to wait for a remote command to end. This might be
  useful if, for example, packer hangs on a connection after a reboot.
  Example: `5m`. Disabled by default.

- `ssh_remote_tunnels` ([]string) - 

- `ssh_local_tunnels` ([]string) - 

<!-- End of code generated from the comments of the SSH struct in communicator/config.go; -->


- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.
