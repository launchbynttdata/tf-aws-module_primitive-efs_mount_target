# AWS EFS Mount Target Resource
# This module creates AWS EFS mount targets in one or more subnets for a given EFS file system.
#
# Key Features:
# - Creates one mount target per subnet
# - Attaches security groups to each mount target
# - Supports optional IP address and AZ name overrides
# - Validates required inputs
# - Exposes mount target IDs, DNS names, and network interface IDs


resource "aws_efs_mount_target" "this" {
  # Create a mount target for each provided subnet
  for_each = { for idx, subnet_id in var.subnet_ids : tostring(idx) => subnet_id }

  # Required: EFS file system ID and subnet ID
  file_system_id = var.efs_filesystem_id
  subnet_id      = each.value

  # Optional: Security groups to associate with the mount target (up to 5)
  security_groups = var.security_group_ids

  # Optional: Specify a static IP address for the mount target in the subnet
  ip_address = var.ip_address


  # Optional: Custom timeouts for create and delete operations
  timeouts {
    create = var.create_timeout
    delete = var.delete_timeout
  }

  # Note: efs_mount_target does NOT support tags as of provider v5.100
  # tags = local.tags
}
