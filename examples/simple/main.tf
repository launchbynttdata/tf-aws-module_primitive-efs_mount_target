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

# Minimal example for EFS mount target module

# provider "aws" {
#   region = var.region
# }

# VPC for EFS
resource "aws_vpc" "this" {
  cidr_block = var.vpc_cidr_block
}

# Subnet for EFS mount target
resource "aws_subnet" "this" {
  vpc_id            = aws_vpc.this.id
  cidr_block        = var.subnet_cidr_block
  availability_zone = var.availability_zone
}

# Security group for EFS
resource "aws_security_group" "this" {
  name   = var.efs_mount_target_project_name != "" ? "${var.efs_mount_target_project_name}-efs-sg" : "efs_mount_target_example-efs-sg"
  vpc_id = aws_vpc.this.id

  ingress {
    from_port   = 2049
    to_port     = 2049
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.this.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.this.cidr_block]
  }
}

# Default security group for VPC
# Configure the default security group to deny all traffic
resource "aws_default_security_group" "default" {
  vpc_id = aws_vpc.this.id

  # No ingress or egress rules = deny all traffic
  tags = {
    Name = "${var.efs_mount_target_project_name}-default-sg"
  }
}

# EFS file system
module "aws_efs_file_system" {
  source = "github.com/launchbynttdata/tf-aws-module_primitive-efs_file_system?ref=1.0.0"

  creation_token = var.efs_file_system_creation_token
  name           = var.efs_mount_target_project_name != "" ? "${var.efs_mount_target_project_name}-efs-fs" : "efs_mount_target_example-efs-fs"
  encrypted      = var.efs_file_system_encrypted

  tags = {
    Environment = "dev"
    Application = var.efs_mount_target_project_name != "" ? "${var.efs_mount_target_project_name}-efs-fs" : "efs_mount_target_example-efs-fs"
  }
}

# EFS mount target module usage
module "efs_mount_target" {
  source             = "../../"
  efs_filesystem_id  = module.aws_efs_file_system.file_system_id
  subnet_ids         = [aws_subnet.this.id]
  security_group_ids = [aws_security_group.this.id]
}
