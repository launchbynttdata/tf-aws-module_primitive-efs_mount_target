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

This module creates AWS EFS mount targets in one or more subnets for a given EFS file system. It is a primitive module, focused on wrapping the [`aws_efs_mount_target`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/efs_mount_target) resource, and is designed for use as a building block in larger infrastructure compositions.

## Features

- Creates one EFS mount target per subnet
- Supports multiple security groups
- Validates required inputs
- Exposes mount target IDs, DNS names, and network interface IDs
- Follows Launch by NTT DATA module standards

## Usage

```hcl
module "efs_mount_target" {
  source             = "launchbynttdata/efs_mount_target/aws"
  efs_filesystem_id  = "fs-12345678"
  subnet_ids         = ["subnet-abc123", "subnet-def456"]
  security_group_ids = ["sg-12345678"]
}
```

## Inputs

| Name                | Description                                                        | Type         | Default | Required |
|---------------------|--------------------------------------------------------------------|--------------|---------|:--------:|
| subnet_ids          | List of subnet IDs where EFS mount targets should be created        | list(string) | n/a     | yes      |
| security_group_ids  | List of security group IDs for the EFS mount targets               | list(string) | n/a     | yes      |
| efs_filesystem_id   | The ID of the EFS file system                                      | string       | n/a     | yes      |
| tags                | Map of tags to assign to resources that support tagging (unused)   | map(string)  | `{}`    | no       |

## Outputs

| Name                           | Description                                              |
|--------------------------------|----------------------------------------------------------|
| mount_target_ids               | Map of subnet ID to EFS mount target ID                  |
| mount_target_dns_names         | Map of subnet ID to EFS mount target DNS name            |
| mount_target_network_interface_ids | Map of subnet ID to network interface ID created for the EFS mount target |

## Validation Rules

- `subnet_ids` must be a non-empty list
- `security_group_ids` must be a non-empty list
- `efs_filesystem_id` must be a non-empty string

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
```

## AWS Documentation

- [aws_efs_mount_target](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/efs_mount_target)

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
| <a name="input_ip_address"></a> [ip\_address](#input\_ip\_address) | (Optional) A specific IP address within the subnet to be used for the EFS mount target. Defaults to AWS-assigned. | `string` | `null` | no |
| <a name="input_subnet_ids"></a> [subnet\_ids](#input\_subnet\_ids) | List of subnet IDs where EFS mount targets should be created. Must be non-empty. | `list(string)` | n/a | yes |
| <a name="input_security_group_ids"></a> [security\_group\_ids](#input\_security\_group\_ids) | List of security group IDs for the EFS mount targets. Must be non-empty. | `list(string)` | n/a | yes |
| <a name="input_efs_filesystem_id"></a> [efs\_filesystem\_id](#input\_efs\_filesystem\_id) | The ID of the EFS file system. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_mount_target_ids"></a> [mount\_target\_ids](#output\_mount\_target\_ids) | Map of subnet ID to EFS mount target ID. |
| <a name="output_mount_target_dns_names"></a> [mount\_target\_dns\_names](#output\_mount\_target\_dns\_names) | Map of subnet ID to EFS file system DNS name (file-system-id.efs.aws-region.amazonaws.com). |
| <a name="output_mount_target_az_dns_names"></a> [mount\_target\_az\_dns\_names](#output\_mount\_target\_az\_dns\_names) | Map of subnet ID to mount target AZ-specific DNS name (availability-zone.file-system-id.efs.aws-region.amazonaws.com). |
| <a name="output_mount_target_file_system_arns"></a> [mount\_target\_file\_system\_arns](#output\_mount\_target\_file\_system\_arns) | Map of subnet ID to EFS file system ARN. |
| <a name="output_mount_target_network_interface_ids"></a> [mount\_target\_network\_interface\_ids](#output\_mount\_target\_network\_interface\_ids) | Map of subnet ID to network interface ID created for the EFS mount target. |
| <a name="output_mount_target_availability_zone_names"></a> [mount\_target\_availability\_zone\_names](#output\_mount\_target\_availability\_zone\_names) | Map of subnet ID to the name of the Availability Zone (AZ) that the mount target resides in. |
| <a name="output_mount_target_availability_zone_ids"></a> [mount\_target\_availability\_zone\_ids](#output\_mount\_target\_availability\_zone\_ids) | Map of subnet ID to the unique identifier of the Availability Zone (AZ) that the mount target resides in. |
| <a name="output_mount_target_owner_ids"></a> [mount\_target\_owner\_ids](#output\_mount\_target\_owner\_ids) | Map of subnet ID to AWS account ID that owns the mount target resource. |
<!-- END_TF_DOCS -->
