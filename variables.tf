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
variable "subnet_id" {
  description = "(Required) The ID of the subnet in which to create the mount target. One mount target should be created per availability zone for high availability."
  type        = string

  validation {
    condition     = var.subnet_id != null && var.subnet_id != ""
    error_message = "You must provide a valid subnet_id."
  }
}

variable "ip_address" {
  description = "(Optional) Static IPv4 address for the mount target within the subnet's CIDR range. If not specified, AWS automatically assigns an available IP address from the subnet."
  type        = string
  default     = null
}

variable "security_group_ids" {
  description = "(Optional) List of security group IDs for the mount target. If not provided, AWS will use the VPC's default security group."
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
