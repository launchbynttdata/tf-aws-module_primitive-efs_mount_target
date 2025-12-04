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

output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.this.id
}

output "subnet_ids" {
  description = "List of all created subnet IDs"
  value       = aws_subnet.this[*].id
}

output "subnet_availability_zones" {
  description = "Map of subnet IDs to their availability zones"
  value = {
    for subnet in aws_subnet.this : subnet.id => subnet.availability_zone
  }
}

output "enabled_mount_target_keys" {
  description = "List of mount target keys that are enabled"
  value       = keys(local.enabled_mount_targets)
}

output "enabled_subnet_ids" {
  description = "Map of mount target key to subnet ID where mount targets are enabled"
  value       = { for k, v in local.enabled_mount_targets : k => v.subnet_id }
}

output "efs_file_system_id" {
  description = "The ID of the EFS file system"
  value       = module.aws_efs_file_system.file_system_id
}

output "efs_file_system_arn" {
  description = "The ARN of the EFS file system"
  value       = module.aws_efs_file_system.file_system_arn
}

output "mount_target_ids" {
  description = "Map of mount target key (e.g., 'az-a') to EFS mount target ID"
  value       = module.efs_mount_target.mount_target_ids
}

output "mount_target_subnet_ids" {
  description = "Map of mount target key to subnet ID"
  value       = module.efs_mount_target.mount_target_subnet_ids
}

output "mount_target_dns_names" {
  description = "Map of mount target key to EFS file system DNS name"
  value       = module.efs_mount_target.mount_target_dns_names
}

output "mount_target_az_dns_names" {
  description = "Map of mount target key to mount target AZ-specific DNS name"
  value       = module.efs_mount_target.mount_target_az_dns_names
}

output "mount_target_network_interface_ids" {
  description = "Map of mount target key to network interface ID"
  value       = module.efs_mount_target.mount_target_network_interface_ids
}

output "mount_target_availability_zones" {
  description = "Map of mount target key to availability zone name"
  value       = module.efs_mount_target.mount_target_availability_zone_names
}

output "security_group_id" {
  description = "The ID of the EFS security group"
  value       = aws_security_group.efs.id
}
