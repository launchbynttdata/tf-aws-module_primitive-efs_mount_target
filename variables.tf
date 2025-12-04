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
variable "create_timeout" {
  description = "(Optional) Timeout for creating the EFS mount target (e.g., '30m')."
  type        = string
  default     = "30m"
}

variable "delete_timeout" {
  description = "(Optional) Timeout for deleting the EFS mount target (e.g., '10m')."
  type        = string
  default     = "10m"
}
variable "mount_targets" {
  description = <<-EOT
    Map of mount target configurations. Each key identifies a mount target (e.g., 'az-a', 'subnet-1').
    This key-based approach ensures mount targets are not rebuilt when configurations change.

    Each mount target supports:
    - subnet_id: (Required) The subnet ID where the mount target will be created
    - ip_address: (Optional) Static IP address for the mount target
    - security_group_ids: (Optional) Security group IDs for this specific mount target
  EOT
  type = map(object({
    subnet_id          = string
    ip_address         = optional(string, null)
    security_group_ids = optional(list(string), null)
  }))

  validation {
    condition     = length(var.mount_targets) > 0
    error_message = "You must provide at least one mount target configuration."
  }

  validation {
    condition = alltrue([
      for k, v in var.mount_targets : v.subnet_id != null && v.subnet_id != ""
    ])
    error_message = "Each mount target must have a valid subnet_id."
  }
}

variable "security_group_ids" {
  description = "(Optional) Default list of security group IDs for mount targets. Can be overridden per mount target. If not provided here or per mount target, AWS will use the VPC's default security group."
  type        = list(string)
  default     = null

  validation {
    condition     = var.security_group_ids == null || length(var.security_group_ids) > 0
    error_message = "If security_group_ids is provided, it must contain at least one security group ID."
  }
}

variable "efs_filesystem_id" {
  description = "The ID of the EFS file system."
  type        = string

  validation {
    condition     = length(var.efs_filesystem_id) > 0
    error_message = "You must provide a valid EFS file system ID."
  }
}

# Not used for aws_efs_mount_target.
# variable "tags" {
#   description = "A map of tags to assign to resources that support tagging."
#   type        = map(string)
#   default     = {}
# }
