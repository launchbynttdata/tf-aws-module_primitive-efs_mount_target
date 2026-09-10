# tf-aws-module_primitive-efs_mount_target

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![License: CC BY-NC-ND 4.0](https://img.shields.io/badge/License-CC_BY--NC--ND_4.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc-nd/4.0/)

## Overview

This primitive module creates a **single** AWS EFS mount target in a specified subnet for a given EFS file system. It wraps the [`aws_efs_mount_target`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/efs_mount_target) resource and is designed to be called once per mount target, typically using `for_each` when multiple mount targets are needed.

## Features

- **Single Mount Target Per Invocation**: Creates one mount target per module call
- **Security Group Support**: Attach multiple security groups to the mount target
- **Optional IP Address**: Specify a static IP or let AWS assign one automatically
- **Input Validation**: Validates required parameters
- **Comprehensive Outputs**: Exposes mount target ID, DNS names, network interface ID, and more

## Usage

### Single Mount Target

```hcl
module "efs_mount_target" {
  source = "../../"

  efs_filesystem_id  = module.aws_efs_file_system.file_system_id
  subnet_id          = aws_subnet.this.id
  security_group_ids = [aws_security_group.this.id]
}
```

### Multiple Mount Targets (Recommended Pattern)

```hcl
locals {
  mount_targets = {
    "az-a" = { subnet_id = aws_subnet.az_a.id }
    "az-b" = { subnet_id = aws_subnet.az_b.id }
  }
}

module "efs_mount_target" {
  source   = "../../"
  for_each = local.mount_targets

  efs_filesystem_id  = module.aws_efs_file_system.file_system_id
  subnet_id          = each.value.subnet_id
  security_group_ids = [aws_security_group.this.id]
}
```

## Examples

See [examples/simple](./examples/simple) and [examples/multi_subnet](./examples/multi_subnet).

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | ~> 1.5 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.100 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_efs_mount_target.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/efs_mount_target) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_create_timeout"></a> [create\_timeout](#input\_create\_timeout) | (Optional) Timeout for creating the EFS mount target (e.g., '30m'). | `string` | `"30m"` | no |
| <a name="input_delete_timeout"></a> [delete\_timeout](#input\_delete\_timeout) | (Optional) Timeout for deleting the EFS mount target (e.g., '10m'). | `string` | `"10m"` | no |
| <a name="input_subnet_id"></a> [subnet\_id](#input\_subnet\_id) | (Required) The ID of the subnet in which to create the mount target. One mount target should be created per availability zone for high availability. | `string` | n/a | yes |
| <a name="input_ip_address"></a> [ip\_address](#input\_ip\_address) | (Optional) Static IPv4 address for the mount target within the subnet's CIDR range. If not specified, AWS automatically assigns an available IP address from the subnet. | `string` | `null` | no |
| <a name="input_security_group_ids"></a> [security\_group\_ids](#input\_security\_group\_ids) | (Optional) List of security group IDs for the mount target. If not provided, AWS will use the VPC's default security group. | `list(string)` | `null` | no |
| <a name="input_efs_filesystem_id"></a> [efs\_filesystem\_id](#input\_efs\_filesystem\_id) | The ID of the EFS file system. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_mount_target_id"></a> [mount\_target\_id](#output\_mount\_target\_id) | The ID of the EFS mount target. |
| <a name="output_mount_target_subnet_id"></a> [mount\_target\_subnet\_id](#output\_mount\_target\_subnet\_id) | The ID of the subnet the mount target is in. |
| <a name="output_mount_target_dns_name"></a> [mount\_target\_dns\_name](#output\_mount\_target\_dns\_name) | The DNS name of the EFS file system (file-system-id.efs.aws-region.amazonaws.com). |
| <a name="output_mount_target_az_dns_name"></a> [mount\_target\_az\_dns\_name](#output\_mount\_target\_az\_dns\_name) | The mount target's availability zone-specific DNS name (availability-zone.file-system-id.efs.aws-region.amazonaws.com). |
| <a name="output_mount_target_file_system_arn"></a> [mount\_target\_file\_system\_arn](#output\_mount\_target\_file\_system\_arn) | Amazon Resource Name (ARN) of the EFS file system. |
| <a name="output_mount_target_network_interface_id"></a> [mount\_target\_network\_interface\_id](#output\_mount\_target\_network\_interface\_id) | The ID of the network interface created for the EFS mount target. |
| <a name="output_mount_target_availability_zone_name"></a> [mount\_target\_availability\_zone\_name](#output\_mount\_target\_availability\_zone\_name) | The name of the Availability Zone (AZ) that the mount target resides in. |
| <a name="output_mount_target_availability_zone_id"></a> [mount\_target\_availability\_zone\_id](#output\_mount\_target\_availability\_zone\_id) | The unique identifier of the Availability Zone (AZ) that the mount target resides in. |
| <a name="output_mount_target_owner_id"></a> [mount\_target\_owner\_id](#output\_mount\_target\_owner\_id) | AWS account ID that owns the mount target resource. |
<!-- END_TF_DOCS -->

## Module Development

### Pre-Requisites

The following commands should be available on your system:

- `asdf` or `mise`
- `make`
- `python3` (for pre-commit)

Additionally, your `git` user and email must be configured. Run the `make configure` command from the root of the repository to ensure that you meet these requirements.

### Pre-Commit hooks

The [.pre-commit-config.yaml](.pre-commit-config.yaml) file defines certain `pre-commit` hooks that are relevant to Terraform and Golang, as well as some common linting tasks. These will be configured for you when you run `make configure`.

### Local Validation

You should validate the changes you make to any module locally, prior to pushing your changes in a branch to GitHub.

1. Ensure that you have run `make configure` successfully.

2. Ensure you are signed into the appropriate cloud provider (e.g. AWS or Azure) for the module under test in your current console session.

3. Run the Terraform and Golang linters with the following command:

```
make lint
```

4. Once you have satisfied the linters, the following command will build example infrastructure in your configured cloud, run the tests, and then tear down the infrastructure it created:

```
make test
```

The pre-commit validations, as well as the `make lint` and `make test` targets, will all be performed in CI. Running these validations locally prior to opening a PR helps ensure a smooth review and merge process.

### Review & Merge Process

Once your change has been tested locally and your branch pushed up, open a new Pull Request for your branch to the default (main) branch of this repository.

The title of your Pull Request will determine the version bump for this change, and the title must be in [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#specification) format in order to merge. A breaking change will trigger a major version bump, a feature will trigger a minor version bump, and all other types will trigger a patch version bump.

Ensure your CI workflows are passing; seek approval from teammates and address any feedback; seek any explicit approvals required by the CODEOWNERS file. You may merge the PR as soon as all requirements are met, and a new release and tag will be automatically created for you.

### Automatic Updates

The shared configuration and workflow files in this repository are largely managed through the [launch-terraform-skeleton](https://github.com/launchbynttdata/launch-terraform-skeleton) repository. Outside of perhaps the `.gitignore` to account for specific files being generated by certain Terraform modules (e.g. Lambda functions), there should not be much cause to update these files on a per-repo basis, and making changes to them individually is discouraged.

If desired, you can check for and run these updates locally in a branch if you have the `copier` tool installed. Some example commands are included below:

```
# Check for updates, optionally checking prerelease versions
copier check-update [--prereleases]

# Run an update, using default answers if there are any. We use tasks, which requires --trust to be set.
copier update --defaults --trust [--prereleases]

# Recopy from the source, and --overwrite all templated files in the process
copier recopy --defaults --trust --overwrite [--prereleases]
```

Automatic updates will run through a scheduled workflow, and if the post-update tests are successful, the Pull Request created will automatically merge. Conflicts in the update or failures to test may leave a Pull Request outstanding, which needs to be addressed by a Launch Engineer.
