# Simple Example: EFS Mount Target

This example demonstrates the minimal configuration required to create an AWS EFS mount target using the module.

## Resources Created

- VPC
- Subnet
- Security Group (NFS access)
- EFS File System
- EFS Mount Target (via module)

## Overview

This example demonstrates the minimal configuration required to create an AWS EFS mount target using the module. It showcases:

- **Region-aware configuration**: Automatically constructs full AZ names from region + letter
- **Simplified AZ specification**: Use single letters ('a', 'b', 'c') instead of full AZ names
- **Single subnet deployment**: Creates one mount target in a specified availability zone

## Usage

```hcl
provider "aws" {
  region = var.region  # e.g., "us-east-2"
}

# Data source for available AZs
data "aws_availability_zones" "available" {
  state = "available"
}

# Computed AZ from region + letter
locals {
  availability_zone = "${var.region}${var.availability_zone_letter}"  # e.g., "us-east-2a"
}

module "efs_mount_target" {
  source            = "../../"
  efs_filesystem_id = module.aws_efs_file_system.file_system_id
  mount_targets = {
    "primary" = {
      subnet_id = aws_subnet.this.id
    }
  }
  security_group_ids = [aws_security_group.this.id]
}
```

## Inputs

| Name                     | Description                                                  | Type   | Default      |
|--------------------------|--------------------------------------------------------------|--------|--------------|
| region                   | AWS region to deploy resources in                            | string | us-east-2    |
| availability_zone_letter | AZ letter suffix (e.g., 'a', 'b', 'c') appended to region    | string | a            |

## Outputs

| Name                              | Description                                  |
|-----------------------------------|----------------------------------------------|
| efs_mount_target_id               | The ID of the EFS mount target               |
| efs_mount_target_dns_name         | The DNS name for the EFS file system         |
| efs_mount_target_network_interface_id | The network interface ID for the mount target |

## How to Run

```sh
terraform init
terraform plan -var-file=test.tfvars
terraform apply -var-file=test.tfvars
```

## Cleanup

```sh
terraform destroy -var-file=test.tfvars
```
# Simple Example

This example provides a basic test case for the `tf-aws-module_primitive-vpc_security_group_ingress_rule` module, used primarily for integration testing.

## Features

- Single SSH ingress rule (port 22)
- IPv4 CIDR source
- Basic configuration

## Usage

```bash
terraform init
terraform plan -var-file=test.tfvars
terraform apply -var-file=test.tfvars
terraform destroy -var-file=test.tfvars
```

## Resources Created

- 1 VPC
- 1 Security Group
- 1 Security Group Ingress Rule

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | ~> 1.5 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.100 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 5.100.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_aws_efs_file_system"></a> [aws\_efs\_file\_system](#module\_aws\_efs\_file\_system) | github.com/launchbynttdata/tf-aws-module_primitive-efs_file_system | 1.0.0 |
| <a name="module_efs_mount_target"></a> [efs\_mount\_target](#module\_efs\_mount\_target) | ../../ | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_default_security_group.default](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/default_security_group) | resource |
| [aws_security_group.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_subnet.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet) | resource |
| [aws_vpc.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_region"></a> [region](#input\_region) | AWS region where resources will be created. | `string` | `"us-east-2"` | no |
| <a name="input_vpc_cidr_block"></a> [vpc\_cidr\_block](#input\_vpc\_cidr\_block) | CIDR block for the VPC. | `string` | `"10.0.0.0/16"` | no |
| <a name="input_subnet_cidr_block"></a> [subnet\_cidr\_block](#input\_subnet\_cidr\_block) | CIDR block for the subnet. | `string` | `"10.0.1.0/24"` | no |
| <a name="input_availability_zone_letter"></a> [availability\_zone\_letter](#input\_availability\_zone\_letter) | Availability Zone letter suffix (e.g., 'a', 'b', 'c') to be appended to the region. | `string` | `"a"` | no |
| <a name="input_efs_file_system_creation_token"></a> [efs\_file\_system\_creation\_token](#input\_efs\_file\_system\_creation\_token) | A unique creation token for the EFS file system. | `string` | `"example-efs"` | no |
| <a name="input_efs_file_system_encrypted"></a> [efs\_file\_system\_encrypted](#input\_efs\_file\_system\_encrypted) | Whether to enable encryption at rest for the EFS file system. | `bool` | `true` | no |
| <a name="input_efs_mount_target_project_name"></a> [efs\_mount\_target\_project\_name](#input\_efs\_mount\_target\_project\_name) | Project name for EFS mount target resources. | `string` | `""` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_mount_target_ids"></a> [mount\_target\_ids](#output\_mount\_target\_ids) | Map of mount target key to EFS mount target ID. |
| <a name="output_mount_target_subnet_ids"></a> [mount\_target\_subnet\_ids](#output\_mount\_target\_subnet\_ids) | Map of mount target key to subnet ID. |
| <a name="output_mount_target_dns_names"></a> [mount\_target\_dns\_names](#output\_mount\_target\_dns\_names) | Map of mount target key to EFS file system DNS name. |
| <a name="output_mount_target_az_dns_names"></a> [mount\_target\_az\_dns\_names](#output\_mount\_target\_az\_dns\_names) | Map of mount target key to mount target AZ-specific DNS name. |
| <a name="output_mount_target_network_interface_ids"></a> [mount\_target\_network\_interface\_ids](#output\_mount\_target\_network\_interface\_ids) | Map of mount target key to network interface ID. |
| <a name="output_mount_target_availability_zone_names"></a> [mount\_target\_availability\_zone\_names](#output\_mount\_target\_availability\_zone\_names) | Map of mount target key to availability zone name. |
| <a name="output_mount_target_availability_zone_ids"></a> [mount\_target\_availability\_zone\_ids](#output\_mount\_target\_availability\_zone\_ids) | Map of mount target key to availability zone ID. |
| <a name="output_efs_file_system_id"></a> [efs\_file\_system\_id](#output\_efs\_file\_system\_id) | The ID of the EFS file system. |
| <a name="output_efs_file_system_arn"></a> [efs\_file\_system\_arn](#output\_efs\_file\_system\_arn) | The ARN of the EFS file system. |
<!-- END_TF_DOCS -->
