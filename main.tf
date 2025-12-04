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
  # Create a mount target for each provided mount target configuration
  # Mount targets enable EC2 instances in a VPC subnet to access the EFS file system
  # Using a map with user-defined keys ensures mount targets are not rebuilt when configurations change
  for_each = var.mount_targets

  # Required: The ID of the EFS file system for which to create the mount target
  file_system_id = var.efs_filesystem_id

  # Required: The ID of the subnet in which to create the mount target
  # One mount target should be created per availability zone for high availability
  subnet_id = each.value.subnet_id

  # Optional: Security groups to associate with the mount target (maximum of 5)
  # These control inbound/outbound traffic to the mount target's network interface
  # Typically allows NFS traffic (port 2049) from application instances
  # Can be specified per mount target or use default security groups
  security_groups = coalesce(each.value.security_group_ids, var.security_group_ids)

  # Optional: Specify a static IPv4 address for the mount target within the subnet's CIDR range
  # If not specified, AWS automatically assigns an available IP address from the subnet
  # Note: Cannot be set to empty string; use null or omit the parameter
  ip_address = each.value.ip_address

  # Optional: Custom timeouts for create and delete operations
  # Adjust these if you experience timeouts in large or complex VPC configurations
  timeouts {
    create = var.create_timeout
    delete = var.delete_timeout
  }

  # Note: aws_efs_mount_target does NOT support tags as of AWS provider v5.100
  # Tags are inherited from the parent EFS file system resource
  # tags = local.tags
}
