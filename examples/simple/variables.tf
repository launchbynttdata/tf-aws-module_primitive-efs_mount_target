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

# Region variable
variable "region" {
  description = "AWS region where resources will be created."
  type        = string
  default     = "us-east-2"

  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-[0-9]{1}$", var.region))
    error_message = "Region must be a valid AWS region format (e.g., us-east-1, eu-west-2)."
  }
}

# VPC and subnet variables
variable "vpc_cidr_block" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"

  validation {
    condition     = can(regex("^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$", var.vpc_cidr_block))
    error_message = "You must provide a valid CIDR block for the VPC."
  }
}

variable "subnet_cidr_block" {
  description = "CIDR block for the subnet."
  type        = string
  default     = "10.0.1.0/24"

  validation {
    condition     = can(regex("^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$", var.subnet_cidr_block))
    error_message = "You must provide a valid CIDR block for the subnet."
  }
}

variable "availability_zone_letter" {
  description = "Availability Zone letter suffix (e.g., 'a', 'b', 'c') to be appended to the region."
  type        = string
  default     = "a"

  validation {
    condition     = can(regex("^[a-z]$", var.availability_zone_letter))
    error_message = "Availability zone letter must be a single lowercase letter (a-z)."
  }
}

# EFS file system variables
variable "efs_file_system_creation_token" {
  description = "A unique creation token for the EFS file system."
  type        = string
  default     = "example-efs"

  validation {
    condition     = length(var.efs_file_system_creation_token) > 0
    error_message = "You must provide a non-empty creation token for the EFS file system."
  }
}

variable "efs_file_system_encrypted" {
  description = "Whether to enable encryption at rest for the EFS file system."
  type        = bool
  default     = true
}
# efs_mount_target module variables
# Note: subnet_ids and security_group_ids will be set automatically by the example

variable "efs_mount_target_project_name" {
  description = "Project name for EFS mount target resources."
  type        = string
  default     = ""
}
