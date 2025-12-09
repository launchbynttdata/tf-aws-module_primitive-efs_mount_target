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
  description = "The ID of the EFS mount target."
  value       = aws_efs_mount_target.this.id
}

output "mount_target_subnet_id" {
  description = "The ID of the subnet the mount target is in."
  value       = aws_efs_mount_target.this.subnet_id
}

output "mount_target_dns_name" {
  description = "The DNS name of the EFS file system (file-system-id.efs.aws-region.amazonaws.com)."
  value       = aws_efs_mount_target.this.dns_name
}

output "mount_target_az_dns_name" {
  description = "The mount target's availability zone-specific DNS name (availability-zone.file-system-id.efs.aws-region.amazonaws.com)."
  value       = aws_efs_mount_target.this.mount_target_dns_name
}

output "mount_target_file_system_arn" {
  description = "Amazon Resource Name (ARN) of the EFS file system."
  value       = aws_efs_mount_target.this.file_system_arn
}

output "mount_target_network_interface_id" {
  description = "The ID of the network interface created for the EFS mount target."
  value       = aws_efs_mount_target.this.network_interface_id
}

output "mount_target_availability_zone_name" {
  description = "The name of the Availability Zone (AZ) that the mount target resides in."
  value       = aws_efs_mount_target.this.availability_zone_name
}

output "mount_target_availability_zone_id" {
  description = "The unique identifier of the Availability Zone (AZ) that the mount target resides in."
  value       = aws_efs_mount_target.this.availability_zone_id
}

output "mount_target_owner_id" {
  description = "AWS account ID that owns the mount target resource."
  value       = aws_efs_mount_target.this.owner_id
}
