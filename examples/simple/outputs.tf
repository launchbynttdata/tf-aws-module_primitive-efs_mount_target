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

output "mount_target_id" {
  description = "The EFS mount target ID."
  value       = module.efs_mount_target.mount_target_id
}

output "mount_target_subnet_id" {
  description = "The subnet ID where the mount target is located."
  value       = module.efs_mount_target.mount_target_subnet_id
}

output "mount_target_dns_name" {
  description = "The EFS file system DNS name."
  value       = module.efs_mount_target.mount_target_dns_name
}

output "mount_target_az_dns_name" {
  description = "The mount target AZ-specific DNS name."
  value       = module.efs_mount_target.mount_target_az_dns_name
}

output "mount_target_network_interface_id" {
  description = "The network interface ID for the mount target."
  value       = module.efs_mount_target.mount_target_network_interface_id
}

output "mount_target_availability_zone_name" {
  description = "The availability zone name where the mount target resides."
  value       = module.efs_mount_target.mount_target_availability_zone_name
}

output "mount_target_availability_zone_id" {
  description = "The availability zone ID where the mount target resides."
  value       = module.efs_mount_target.mount_target_availability_zone_id
}

output "efs_file_system_id" {
  description = "The ID of the EFS file system."
  value       = module.aws_efs_file_system.file_system_id
}

output "efs_file_system_arn" {
  description = "The ARN of the EFS file system."
  value       = module.aws_efs_file_system.file_system_arn
}
