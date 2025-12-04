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

output "mount_target_ids" {
  description = "Map of mount target key to EFS mount target ID."
  value       = module.efs_mount_target.mount_target_ids
}

output "mount_target_subnet_ids" {
  description = "Map of mount target key to subnet ID."
  value       = module.efs_mount_target.mount_target_subnet_ids
}

output "mount_target_dns_names" {
  description = "Map of mount target key to EFS file system DNS name."
  value       = module.efs_mount_target.mount_target_dns_names
}

output "mount_target_az_dns_names" {
  description = "Map of mount target key to mount target AZ-specific DNS name."
  value       = module.efs_mount_target.mount_target_az_dns_names
}

output "mount_target_network_interface_ids" {
  description = "Map of mount target key to network interface ID."
  value       = module.efs_mount_target.mount_target_network_interface_ids
}

output "mount_target_availability_zone_names" {
  description = "Map of mount target key to availability zone name."
  value       = module.efs_mount_target.mount_target_availability_zone_names
}

output "mount_target_availability_zone_ids" {
  description = "Map of mount target key to availability zone ID."
  value       = module.efs_mount_target.mount_target_availability_zone_ids
}

output "efs_file_system_id" {
  description = "The ID of the EFS file system."
  value       = module.aws_efs_file_system.file_system_id
}

output "efs_file_system_arn" {
  description = "The ARN of the EFS file system."
  value       = module.aws_efs_file_system.file_system_arn
}
