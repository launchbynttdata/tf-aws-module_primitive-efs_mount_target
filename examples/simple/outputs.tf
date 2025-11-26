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

output "aws_efs_mount_target_id" {
  description = "The ID of the EFS mount target created."
  value       = module.efs_mount_target.mount_target_ids
}

output "aws_efs_mount_target_dns_name" {
  description = "The DNS name for the EFS file system."
  value       = module.efs_mount_target.mount_target_dns_names
}

output "aws_efs_mount_target_network_interface_id" {
  description = "The network interface ID for the EFS mount target."
  value       = module.efs_mount_target.mount_target_network_interface_ids
}
