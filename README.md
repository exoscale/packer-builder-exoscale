# Packer Builder plugin: Exoscale

[![Actions Status](https://github.com/exoscale/packer-builder-exoscale/workflows/CI/badge.svg)](https://github.com/exoscale/packer-builder-exoscale/actions?query=workflow%3ACI)

The `exoscale` builder plugin can be used with HashiCorp [Packer][packer]
to create a new [instance template][customtemplatesdoc] from a Compute
instance volume snapshot. This plugin creates a Compute instance in your
Exoscale organization, logs into it via SSH to execute configured
provisioners and then snapshots the storage volume of the instance to
register a custom template from the exported snapshot.

**Note:** the `exoscale` Packer builder only supports UNIX-like operating
systems (e.g. GNU/Linux, *BSD...). To build Exoscale custom templates for
other OS, we recommend using the [QEMU][packerqemu] builder combined with the
[exoscale-import][exoscale-import] Packer post-processor plugin.


## Installation

### Using pre-built releases

You can find pre-built releases of the plugin [here][releases].
Once you have downloaded the latest archive corresponding to your target OS,
uncompress it to retrieve the `packer-builder-exoscale` plugin binary file.


### From Sources

If you prefer to build the plugin from sources, clone the GitHub repository
locally and run the command `make build` from the root of the sources
directory. Upon successful compilation, a `packer-builder-exoscale` plugin
binary file can be found in the `/bin` directory.


## Configuration

To use the plugin with Packer, please follow the official documentation on
[installing a Packer plugin][packerplugindoc].

Here is the list of supported configuration parameters by the post-processor.


### Required parameters

- `api_key` (string) - The API key used to communicate with Exoscale services.

- `api_secret` (string) - The API secret used to communicate with Exoscale
  services.

- `instance_template` (string) - The name of the template to use when creating
  the Compute instance.

- `template_zone` (string) - The Exoscale [zone][zones] in which to create the
  template.
  
- `template_name` (string) - The name of the template.


### Optional parameters

- `api_endpoint` (string) - The API endpoint used to communicate with the
  Exoscale API. Defaults to `https://api.exoscale.com/v1`.

- `instance_type` (string) - The instance type of the Compute instance.
  Defaults to `Medium`.

- `instance_name` (string) - The name of the Compute instance.
  Defaults to `packer-<BUILD ID>`.

- `instance_zone` (string) - The Exoscale zone in which to create the Compute
  instance. Defaults to the value of `template_zone`.

- `instance_template_filter` (string) - The template filter to specify for the
  `instance_template` parameter. Defaults to `featured`.

- `instance_disk_size` (int) - Volume disk size in GB of the Compute instance
  to create. Defaults to `50`.

- `instance_security_group` (string) - Security Group to use for the Compute
  instance. Defaults to `default`.

- `instance_ssh_key` (string) - Name of the Exoscale SSH key to use with the
  Compute instance. If unset, a throwaway SSH key named `packer-<BUILD ID>`
  will be created before creating the instance, and destroyed after a
  successful build.

- `template_description` (string) - The description of the template.

- `template_username` (string) - An optional username to be used to log into
  Compute instances using this template.

- `template_disable_password` (boolean) - Whether the template should disable
  Compute instance password reset. Defaults to `false`.

- `template_disable_sshkey` (boolean) - Whether the template should disable
  SSH key installation during Compute instance creation. Defaults to `false`.

In addition to plugin-specific configuration parameters, you can also adjust
the [SSH communicator][packerssh] settings to configure how Packer will log
into the Compute instance.


## Usage

Here is an example of a simple Packer configuration using the `exoscale`
builder:

```json
{
  "variables": {
    "api_key": "{{env `EXOSCALE_API_KEY`}}",
    "api_secret": "{{env `EXOSCALE_API_SECRET`}}"
  },

  "builders": [{
    "type": "exoscale",
    "api_key": "{{user `api_key`}}",
    "api_secret": "{{user `api_secret`}}",
    "instance_template": "Linux Ubuntu 20.04 LTS 64-bit",
    "template_zone": "ch-gva-2",
    "template_name": "my-app",
    "template_username": "ubuntu",
    "ssh_username": "ubuntu"
  }],

  "provisioners": [{
    "type": "shell",
    "execute_command": "chmod +x {{.Path}}; sudo {{.Path}}",
    "scripts": ["install.sh"] 
  }]
}
```

The same configuration in [HCL][packerhcl] format (only with Packer >= 1.5.0):

```hcl
variable "api_key" { default = "" }
variable "api_secret" { default = "" }

source "exoscale" "my-app" {
  api_key = var.api_key
  api_secret = var.api_secret
  instance_template = "Linux Ubuntu 20.04 LTS 64-bit"
  instance_security_group = "packer"
  template_zone = "ch-gva-2"
  template_name = "my-app"
  template_username = "ubuntu"
  ssh_username = "ubuntu"
}

build {
  sources = ["source.exoscale.test"]

  provisioner "shell" {
    execute_command = "chmod +x {{.Path}}; sudo {{.Path}}"
    scripts = ["install.sh"]
  }
}
```

To build your template using Packer, run the following command:

```console
$ packer build my-app.pkr.hcl
exoscale: output will be in this color.

==> exoscale: Build ID: brjpaeh8d3b2liin7pm0
==> exoscale: Creating Compute instance
==> exoscale: Using ssh communicator to connect: 159.100.242.23
==> exoscale: Waiting for SSH to become available...
==> exoscale: Connected to SSH!
==> exoscale: Provisioning with shell script: install.sh
==> exoscale: Stopping Compute instance
==> exoscale: Creating Compute instance snapshot
==> exoscale: Exporting Compute instance snapshot
==> exoscale: Registering Compute instance template
==> exoscale: Destroying Compute instance template
Build 'exoscale' finished.

==> Builds finished. The artifacts of successful builds are:
--> exoscale: my-app @ ch-gva-2 (423e0bda-f127-417e-9c10-4e412d596478)

$ exo vm template show 423e0bda-f127-417e-9c10-4e412d596478
  ┼───────────────┼──────────────────────────────────────┼
  │   TEMPLATE    │                                      │
  ┼───────────────┼──────────────────────────────────────┼
  │ ID            │ 423e0bda-f127-417e-9c10-4e412d596478 │
  │ Name          │ my-app                               │
  │ OS Type       │ Other (64-bit)                       │
  │ Creation Date │ 2020-06-15T17:40:27+0200             │
  │ Zone          │ ch-gva-2                             │
  │ Disk Size     │ 50 GiB                               │
  │ Username      │ ubuntu                               │
  │ Password?     │ true                                 │
  ┼───────────────┼──────────────────────────────────────┼
```

[releases]: https://github.com/exoscale/packer-builder-exoscale/releases
[packer]: https://www.packer.io/
[packerintro]: https://www.packer.io/intro/
[packerqemu]: https://www.packer.io/docs/builders/qemu/
[customtemplatesdoc]: https://community.exoscale.com/documentation/compute/custom-templates/
[packerplugindoc]: https://www.packer.io/docs/extending/plugins/#installing-plugins
[packerhcl]: https://www.packer.io/guides/hcl/
[packerssh]: https://www.packer.io/docs/communicators/ssh/
[exoscale-import]: https://github.com/exoscale/packer-post-processor-exoscale-import
[zones]: https://www.exoscale.com/datacenters/
