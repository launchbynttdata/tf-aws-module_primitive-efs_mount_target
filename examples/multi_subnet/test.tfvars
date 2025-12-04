# Multi-subnet example configuration
# This demonstrates EFS mount targets across multiple availability zones

project_name = "efs-mt-test"
environment  = "test"

vpc_cidr_block = "10.1.0.0/16"

# Define three subnets across different AZs
subnet_configs = [
  {
    cidr_block        = "10.1.1.0/24"
    availability_zone = "us-east-2a"
  },
  {
    cidr_block        = "10.1.2.0/24"
    availability_zone = "us-east-2b"
  },
  {
    cidr_block        = "10.1.3.0/24"
    availability_zone = "us-east-2c"
  }
]

# Test scenarios:
# 1. All subnets enabled (default): enabled_subnet_indices = null or [0, 1, 2]
# 2. Remove middle subnet: enabled_subnet_indices = [0, 2]
# 3. Only first and second: enabled_subnet_indices = [0, 1]
# 4. Single subnet: enabled_subnet_indices = [1]

# Start with all subnets enabled
enabled_subnet_indices = null

efs_encrypted = true

# Timeouts
mount_target_create_timeout = "30m"
mount_target_delete_timeout = "10m"
