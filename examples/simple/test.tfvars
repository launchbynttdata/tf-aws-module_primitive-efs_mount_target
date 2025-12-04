# Example configuration for EFS mount target module
# Copy this file to test.tfvars and customize the values as needed


# VPC and subnet variables
vpc_cidr_block    = "10.0.0.0/16"
subnet_cidr_block = "10.0.1.0/24"
availability_zone = "us-west-2a"

# EFS file system variables
efs_file_system_creation_token = "efs_mount_target_example_efs-fs-token"
efs_file_system_encrypted      = true

# efs_mount_target module variables
# Note: subnet_ids and security_group_ids will be set automatically by the example
efs_mount_target_project_name = "efs_mt_example"
