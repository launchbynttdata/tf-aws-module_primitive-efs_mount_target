output "mount_target_ids" {
  description = "List of EFS mount target IDs"
  value       = { for k, v in aws_efs_mount_target.this : k => v.id }
}

output "mount_target_dns_names" {
  description = "List of EFS mount target DNS names"
  value       = { for k, v in aws_efs_mount_target.this : k => v.dns_name }
}

output "mount_target_network_interface_ids" {
  description = "List of network interface IDs that Amazon EFS created"
  value       = { for k, v in aws_efs_mount_target.this : k => v.network_interface_id }
}
