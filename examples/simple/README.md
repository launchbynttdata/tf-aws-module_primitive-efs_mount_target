# Simple Example: EFS Mount Target

This example demonstrates the minimal configuration required to create a single AWS EFS mount target using the primitive module.

## Key Features

- **Single Mount Target**: Creates one mount target in a single subnet
- **Primitive Module Pattern**: The module creates one mount target per invocation
- **Region-aware configuration**: Automatically constructs full AZ names from region + letter
- **Simplified AZ specification**: Use single letters ('a', 'b', 'c') instead of full AZ names
- **Minimal configuration**: Demonstrates the core required inputs

## Resources Created

- VPC with DNS support enabled
- Subnet in specified availability zone
- Security Group (NFS access on port 2049)
- Default Security Group (deny all traffic)
- EFS File System (encrypted)
- EFS Mount Target (via primitive module)

## Architecture

The primitive module is called once to create a single mount target:

```hcl
# Computed AZ from region + letter
locals {
  availability_zone = "${var.region}${var.availability_zone_letter}"  # e.g., "us-east-2a"
}

# Module creates a single mount target
module "efs_mount_target" {
  source            = "../../"
  efs_filesystem_id = module.aws_efs_file_system.file_system_id

  # Direct configuration for single mount target
  subnet_id          = aws_subnet.this.id
  security_group_ids = [aws_security_group.this.id]
}
```

## Validation Tests

The following test scenario validates the module:

1. **Initialize** - Run `terraform init` to download providers and modules
2. **Validate** - Run `terraform validate` to check configuration syntax
3. **Plan** - Review the execution plan to verify resources
4. **Apply** - Create the EFS mount target and supporting infrastructure
5. **Verify** - Confirm mount target is accessible and properly configured
6. **Destroy** - Clean teardown of all resources

## Usage

### Step-by-Step Validation

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Review execution plan
terraform plan -var-file=test.tfvars

# Create resources
terraform apply -var-file=test.tfvars

# Verify outputs
terraform output

# Clean up
terraform destroy -var-file=test.tfvars
```

### Quick Test

```bash
terraform init && \
terraform apply -var-file=test.tfvars -auto-approve && \
terraform destroy -var-file=test.tfvars -auto-approve
```

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
| <a name="output_mount_target_id"></a> [mount\_target\_id](#output\_mount\_target\_id) | The EFS mount target ID. |
| <a name="output_mount_target_subnet_id"></a> [mount\_target\_subnet\_id](#output\_mount\_target\_subnet\_id) | The subnet ID where the mount target is located. |
| <a name="output_mount_target_dns_name"></a> [mount\_target\_dns\_name](#output\_mount\_target\_dns\_name) | The EFS file system DNS name. |
| <a name="output_mount_target_az_dns_name"></a> [mount\_target\_az\_dns\_name](#output\_mount\_target\_az\_dns\_name) | The mount target AZ-specific DNS name. |
| <a name="output_mount_target_network_interface_id"></a> [mount\_target\_network\_interface\_id](#output\_mount\_target\_network\_interface\_id) | The network interface ID for the mount target. |
| <a name="output_mount_target_availability_zone_name"></a> [mount\_target\_availability\_zone\_name](#output\_mount\_target\_availability\_zone\_name) | The availability zone name where the mount target resides. |
| <a name="output_mount_target_availability_zone_id"></a> [mount\_target\_availability\_zone\_id](#output\_mount\_target\_availability\_zone\_id) | The availability zone ID where the mount target resides. |
| <a name="output_efs_file_system_id"></a> [efs\_file\_system\_id](#output\_efs\_file\_system\_id) | The ID of the EFS file system. |
| <a name="output_efs_file_system_arn"></a> [efs\_file\_system\_arn](#output\_efs\_file\_system\_arn) | The ARN of the EFS file system. |
<!-- END_TF_DOCS -->
