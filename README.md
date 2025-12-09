# tf-aws-module_primitive-efs_mount_target

## What is a Primitive Module?

A **primitive module** is a thin, focused Terraform wrapper around a single AWS resource type. Primitive modules:

- Wrap a **single AWS resource** (e.g., `aws_eks_cluster`, `aws_kms_key`, `aws_s3_bucket`)
- Provide sensible defaults while maintaining full configurability
- Include comprehensive validation rules
- Follow consistent patterns for inputs, outputs, and tagging
- Include automated testing using Terratest
- Serve as building blocks for higher-level composite modules

For examples of well-structured primitive modules, see:

- [tf-aws-module_primitive-eks_cluster](https://github.com/launchbynttdata/tf-aws-module_primitive-eks_cluster)
- [tf-aws-module_primitive-kms_key](https://github.com/launchbynttdata/tf-aws-module_primitive-kms_key)

---

## Getting Started with This Template

### 1. Create Your New Module Repository

1. Click the "Use this template" button on GitHub
2. Name your repository following the naming convention: `tf-aws-module_primitive-<resource_name>`
   - Examples: `tf-aws-module_primitive-s3_bucket`, `tf-aws-module_primitive-lambda_function`
3. Clone your new repository locally

### 2. Initialize and Clean Up Template References

After cloning, run the cleanup target to update template references with your actual repository information:

```bash
make init-module
```

This command will:

- Update the `go.mod` file with your repository's GitHub URL
- Update test imports to reference your new module name
- Remove template-specific placeholders

### 3. Configure Your Environment

Install required development dependencies:

```bash
make configure-dependencies
make configure-git-hooks
```

This installs:

- Terraform
- Go
- Pre-commit hooks
- Other development tools specified in `.tool-versions`

---

## HOWTO: Developing a Primitive Module

### Step 1: Define Your Resource

1. **Identify the AWS resource** you're wrapping (e.g., `aws_eks_cluster`)
2. **Review AWS documentation** for the resource to understand all available parameters
3. **Study similar primitive modules** for patterns and best practices

### Step 2: Create the Module Structure

Your primitive module should include these core files:

#### `main.tf`

- Contains the primary resource declaration
- Should be clean and focused on the single resource
- Example:

```hcl
resource "aws_eks_cluster" "this" {
  name     = var.name
  role_arn = var.role_arn
  version  = var.kubernetes_version

  vpc_config {
    subnet_ids              = var.vpc_config.subnet_ids
    security_group_ids      = var.vpc_config.security_group_ids
    endpoint_private_access = var.vpc_config.endpoint_private_access
    endpoint_public_access  = var.vpc_config.endpoint_public_access
    public_access_cidrs     = var.vpc_config.public_access_cidrs
  }

  tags = merge(
    var.tags,
    local.default_tags
  )
}
```

#### `variables.tf`

- Define all configurable parameters
- Include clear descriptions for each variable
- Set sensible defaults where appropriate
- Use validation rules to enforce constraints, but only when the validations can be made precise.
- Alternatively, use [`check`](https://developer.hashicorp.com/terraform/language/block/check) blocks to create more complicated validations. (Requires terraform ~> 1.12)
- Example:

```hcl
variable "name" {
  description = "Name of the EKS cluster"
  type        = string

  validation {
    condition     = length(var.name) <= 100
    error_message = "Cluster name must be 100 characters or less"
  }
}

variable "kubernetes_version" {
  description = "Kubernetes version to use for the EKS cluster"
  type        = string
  default     = null

  validation {
    condition     = var.kubernetes_version == null || can(regex("^1\\.(2[89]|[3-9][0-9])$", var.kubernetes_version))
    error_message = "Kubernetes version must be 1.28 or higher"
  }
}
```

#### `outputs.tf`

- Export all useful attributes of the resource
- Include comprehensive outputs for downstream consumption
- Document what each output provides
- Example:

```hcl
output "id" {
  description = "The ID of the EKS cluster"
  value       = aws_eks_cluster.this.id
}

output "arn" {
  description = "The ARN of the EKS cluster"
  value       = aws_eks_cluster.this.arn
}

output "endpoint" {
  description = "The endpoint for the EKS cluster API server"
  value       = aws_eks_cluster.this.endpoint
}
```

#### `locals.tf`

- Define local values and transformations
- Include standard tags (e.g., `provisioner = "Terraform"`)
- Example:

```hcl
locals {
  default_tags = {
    provisioner = "Terraform"
  }
}

# tf-aws-module_primitive-efs_mount_target

## Overview

This primitive module creates a **single** AWS EFS mount target in a specified subnet for a given EFS file system. It wraps the [`aws_efs_mount_target`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/efs_mount_target) resource and is designed to be called once per mount target, typically using `for_each` when multiple mount targets are needed.

## Features

- **Single Mount Target Per Invocation**: Creates one mount target per module call
- **Security Group Support**: Attach multiple security groups to the mount target
- **Optional IP Address**: Specify a static IP or let AWS assign one automatically
- **Input Validation**: Validates required parameters
- **Comprehensive Outputs**: Exposes mount target ID, DNS names, network interface ID, and more
- **Follows Launch by NTT DATA Standards**: Consistent patterns and best practices

## Usage

### Single Mount Target

```hcl
module "efs_mount_target" {
  source             = "launchbynttdata/efs_mount_target/aws"
  efs_filesystem_id  = "fs-12345678"
  subnet_id          = "subnet-abc123"
  security_group_ids = ["sg-12345678"]
}
```

### Multiple Mount Targets (Recommended Pattern)

Use `for_each` to create multiple mount targets across availability zones:

```hcl
locals {
  mount_targets = {
    "az-a" = { subnet_id = "subnet-abc123" }
    "az-b" = { subnet_id = "subnet-def456" }
    "az-c" = { subnet_id = "subnet-ghi789" }
  }
}

module "efs_mount_target" {
  source   = "launchbynttdata/efs_mount_target/aws"
  for_each = local.mount_targets

  efs_filesystem_id  = "fs-12345678"
  subnet_id          = each.value.subnet_id
  security_group_ids = ["sg-12345678"]
}

# Access outputs
output "mount_target_ids" {
  value = { for k, v in module.efs_mount_target : k => v.mount_target_id }
}
```

## Inputs

| Name                | Description                                                        | Type         | Default | Required |
|---------------------|--------------------------------------------------------------------|--------------|---------|:--------:|
| efs_filesystem_id   | The ID of the EFS file system                                      | string       | n/a     | yes      |
| subnet_id           | The subnet ID where the mount target will be created               | string       | n/a     | yes      |
| security_group_ids  | List of security group IDs for the mount target                    | list(string) | `null`  | no       |
| ip_address          | Static IPv4 address for the mount target (optional)                | string       | `null`  | no       |
| create_timeout      | Timeout for creating the mount target                              | string       | `"30m"` | no       |
| delete_timeout      | Timeout for deleting the mount target                              | string       | `"10m"` | no       |

## Outputs

| Name                               | Description                                              |
|------------------------------------|----------------------------------------------------------|
| mount_target_id                    | The ID of the EFS mount target                           |
| mount_target_subnet_id             | The subnet ID where the mount target is located          |
| mount_target_dns_name              | The DNS name for the EFS file system                     |
| mount_target_az_dns_name           | The AZ-specific DNS name for the mount target            |
| mount_target_file_system_arn       | The ARN of the EFS file system                           |
| mount_target_network_interface_id  | The network interface ID for the mount target            |
| mount_target_availability_zone_name| The availability zone name                               |
| mount_target_availability_zone_id  | The availability zone ID                                 |
| mount_target_owner_id              | The AWS account ID that owns the mount target            |

## Validation Rules

- `subnet_id` must be a non-empty string
- `security_group_ids` must be null or contain at least one security group ID
- `efs_filesystem_id` must be a non-empty string

## Why This Pattern?

### Primitive Module Design

This module creates **one mount target per invocation** following primitive module best practices:

- **Simplicity**: Each module call has a clear, single responsibility
- **Flexibility**: Callers control how mount targets are organized using `for_each` or `count`
- **Stability**: Resource addresses are determined by the caller's `for_each` keys
- **Composability**: Easy to integrate into higher-level modules

### Benefits

1. **Stable Infrastructure**: Using static keys in `for_each` prevents unnecessary rebuilds
2. **Predictable Behavior**: Adding/removing mount targets only affects those specific resources
3. **Clear Resource Addressing**: `module.efs_mount_target["az-a"]` is explicit and readable
4. **Testability**: Simple to test individual mount target creation

## Examples

This repository includes two examples demonstrating different use cases:

- **[simple](./examples/simple/)**: Basic single mount target deployment
- **[multi_subnet](./examples/multi_subnet/)**: Multiple mount targets across availability zones using `for_each`

## Testing

Run all validation and tests:

```bash
make check
```

To deploy an example:

```bash
cd examples/simple
terraform init
terraform plan -var-file=test.tfvars -out=the.tfplan
terraform apply the.tfplan
terraform destroy -var-file=test.tfvars
```

## AWS Documentation

- [aws_efs_mount_target](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/efs_mount_target)
- [Amazon EFS: How It Works](https://docs.aws.amazon.com/efs/latest/ug/how-it-works.html)

## License

Apache 2.0. See LICENSE and NOTICE files for details.
- Adjust test context as needed

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
