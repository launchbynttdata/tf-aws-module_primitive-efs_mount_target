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

variable "project_name" {
  description = "Project name used for resource naming and tagging"
  type        = string
  default     = "efs-mt-multi"
}

variable "environment" {
  description = "Environment name (e.g., dev, test, prod)"
  type        = string
  default     = "test"
}

variable "vpc_cidr_block" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"

  validation {
    condition     = can(cidrhost(var.vpc_cidr_block, 0))
    error_message = "Must be a valid IPv4 CIDR block."
  }
}

variable "subnet_configs" {
  description = "List of subnet configurations with CIDR blocks and availability zones"
  type = list(object({
    cidr_block        = string
    availability_zone = string
  }))

  validation {
    condition     = length(var.subnet_configs) > 0
    error_message = "At least one subnet configuration must be provided."
  }

  validation {
    condition = alltrue([
      for config in var.subnet_configs : can(cidrhost(config.cidr_block, 0))
    ])
    error_message = "All CIDR blocks must be valid IPv4 CIDR blocks."
  }

  validation {
    condition = alltrue([
      for config in var.subnet_configs : length(config.availability_zone) > 0
    ])
    error_message = "All availability zones must be non-empty strings."
  }
}

variable "enabled_subnet_indices" {
  description = <<-EOT
    List of subnet indices (0-based) to create mount targets in.
    If null or empty, mount targets will be created in all subnets.

    Example: [0, 2] creates mount targets only in the 1st and 3rd subnets.
    This allows testing subnet removal/addition without rebuilding existing mount targets.
  EOT
  type        = list(number)
  default     = null

  validation {
    condition = var.enabled_subnet_indices == null || alltrue([
      for idx in var.enabled_subnet_indices : idx >= 0
    ])
    error_message = "All subnet indices must be non-negative integers."
  }
}

variable "efs_encrypted" {
  description = "Whether to enable encryption at rest for the EFS file system"
  type        = bool
  default     = true
}

variable "mount_target_create_timeout" {
  description = "Timeout for creating EFS mount targets"
  type        = string
  default     = "30m"
}

variable "mount_target_delete_timeout" {
  description = "Timeout for deleting EFS mount targets"
  type        = string
  default     = "10m"
}
