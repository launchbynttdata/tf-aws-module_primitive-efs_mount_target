# AWS EFS Mount Target Resource
# This module creates a single AWS EFS mount target in a specified subnet for a given EFS file system.
#
# Key Features:
# - Creates one mount target per module invocation
# - Attaches security groups to the mount target
# - Supports optional IP address override
# - Validates required inputs
# - Exposes mount target IDs, DNS names, and network interface IDs
# - Designed to be called multiple times using for_each at the caller level


resource "aws_efs_mount_target" "this" {
  # Mount targets enable EC2 instances in a VPC subnet to access the EFS file system
  # One mount target should be created per availability zone for high availability

  # Required: The ID of the EFS file system for which to create the mount target
  file_system_id = var.efs_filesystem_id

  # Required: The ID of the subnet in which to create the mount target
  # One mount target should be created per availability zone for high availability
  subnet_id = var.subnet_id

  # Optional: Security groups to associate with the mount target (maximum of 5)
  # These control inbound/outbound traffic to the mount target's network interface
  # Typically allows NFS traffic (port 2049) from application instances
  security_groups = var.security_group_ids

  # Optional: Specify a static IPv4 address for the mount target within the subnet's CIDR range
  # If not specified, AWS automatically assigns an available IP address from the subnet
  # Note: Cannot be set to empty string; use null or omit the parameter
  ip_address = var.ip_address

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
