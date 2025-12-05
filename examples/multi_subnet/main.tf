// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

# Multi-subnet example for EFS mount target module
# This example demonstrates:
# - Creating mount targets across multiple availability zones
# - Stable infrastructure when adding/removing subnets from the list
# - No rebuilds of existing mount targets when subnet list order changes

# VPC for EFS
resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr_block
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.project_name}-vpc"
    Environment = var.environment
  }
}

# Subnets across multiple AZs - using count for dynamic creation
resource "aws_subnet" "this" {
  count = length(local.subnet_configs_with_full_az)

  vpc_id            = aws_vpc.this.id
  cidr_block        = local.subnet_configs_with_full_az[count.index].cidr_block
  availability_zone = local.subnet_configs_with_full_az[count.index].availability_zone

  tags = {
    Name        = "${var.project_name}-subnet-${local.subnet_configs_with_full_az[count.index].availability_zone}"
    Environment = var.environment
    AZ          = local.subnet_configs_with_full_az[count.index].availability_zone
  }
}

# Security group for EFS
resource "aws_security_group" "efs" {
  name_prefix = "${var.project_name}-efs-"
  description = "Security group for EFS mount targets"
  vpc_id      = aws_vpc.this.id

  ingress {
    description = "NFS from VPC"
    from_port   = 2049
    to_port     = 2049
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.this.cidr_block]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-efs-sg"
    Environment = var.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Default security group - deny all traffic
resource "aws_default_security_group" "default" {
  vpc_id = aws_vpc.this.id

  tags = {
    Name        = "${var.project_name}-default-sg"
    Environment = var.environment
  }
}

# EFS file system
module "aws_efs_file_system" {
  source = "github.com/launchbynttdata/tf-aws-module_primitive-efs_file_system?ref=1.0.0"

  creation_token = "${var.project_name}-${var.environment}"
  name           = "${var.project_name}-efs-fs"
  encrypted      = var.efs_encrypted

  tags = {
    Name        = "${var.project_name}-efs-fs"
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Local values for subnet and mount target configuration
locals {
  # Construct full AZ names from region and letters
  subnet_configs_with_full_az = [
    for config in var.subnet_configs : {
      cidr_block        = config.cidr_block
      availability_zone = "${var.region}${config.az_letter}"
      az_letter         = config.az_letter
    }
  ]

  # Create a map of mount targets based on enabled_subnet_indices
  # Keys are static (az-a, az-b, az-c) so they're known at plan time
  # This prevents the "unknown values in for_each" error
  all_mount_targets = {
    for idx, config in local.subnet_configs_with_full_az :
    "az-${config.az_letter}" => {
      subnet_id = aws_subnet.this[idx].id
    }
  }

  # Filter based on enabled_subnet_indices if provided
  enabled_mount_targets = var.enabled_subnet_indices != null && length(var.enabled_subnet_indices) > 0 ? {
    for idx in var.enabled_subnet_indices :
    "az-${local.subnet_configs_with_full_az[idx].az_letter}" => local.all_mount_targets["az-${local.subnet_configs_with_full_az[idx].az_letter}"]
  } : local.all_mount_targets
}

# EFS mount target module usage
# The module is now called once per mount target using for_each
# Each module invocation creates a single mount target
# Using map with static keys (az-a, az-b, az-c) ensures:
# 1. Keys are known at plan time (no "unknown values in for_each" error)
# 2. Mount targets are not rebuilt when configurations change
# 3. Resource addresses are stable and readable
module "efs_mount_target" {
  source   = "../../"
  for_each = local.enabled_mount_targets

  efs_filesystem_id = module.aws_efs_file_system.file_system_id

  # Individual mount target configuration
  subnet_id = each.value.subnet_id

  # Security groups for all mount targets
  security_group_ids = [aws_security_group.efs.id]

  # Optional: Custom timeouts for testing
  create_timeout = var.mount_target_create_timeout
  delete_timeout = var.mount_target_delete_timeout
}
