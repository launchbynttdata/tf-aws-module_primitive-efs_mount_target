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
  value       = { for k, v in aws_efs_mount_target.this : k => v.id }
}

output "mount_target_subnet_ids" {
  description = "Map of mount target key to subnet ID."
  value       = { for k, v in aws_efs_mount_target.this : k => v.subnet_id }
}

output "mount_target_dns_names" {
  description = "Map of mount target key to EFS file system DNS name (file-system-id.efs.aws-region.amazonaws.com)."
  value       = { for k, v in aws_efs_mount_target.this : k => v.dns_name }
}

output "mount_target_az_dns_names" {
  description = "Map of mount target key to mount target AZ-specific DNS name (availability-zone.file-system-id.efs.aws-region.amazonaws.com)."
  value       = { for k, v in aws_efs_mount_target.this : k => v.mount_target_dns_name }
}

output "mount_target_file_system_arns" {
  description = "Map of mount target key to EFS file system ARN."
  value       = { for k, v in aws_efs_mount_target.this : k => v.file_system_arn }
}

output "mount_target_network_interface_ids" {
  description = "Map of mount target key to network interface ID created for the EFS mount target."
  value       = { for k, v in aws_efs_mount_target.this : k => v.network_interface_id }
}

output "mount_target_availability_zone_names" {
  description = "Map of mount target key to the name of the Availability Zone (AZ) that the mount target resides in."
  value       = { for k, v in aws_efs_mount_target.this : k => v.availability_zone_name }
}

output "mount_target_availability_zone_ids" {
  description = "Map of mount target key to the unique identifier of the Availability Zone (AZ) that the mount target resides in."
  value       = { for k, v in aws_efs_mount_target.this : k => v.availability_zone_id }
}

output "mount_target_owner_ids" {
  description = "Map of mount target key to AWS account ID that owns the mount target resource."
  value       = { for k, v in aws_efs_mount_target.this : k => v.owner_id }
}
