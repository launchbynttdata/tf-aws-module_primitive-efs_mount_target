variable "subnet_ids" {
  description = "List of subnet IDs where EFS mount targets should be created"
  type        = list(string)
  default     = []
}

variable "security_group_ids" {
  description = "List of security group IDs for the EFS mount targets"
  type        = list(string)
  default     = []
}

variable "efs_filesystem_id" {
  description = "The ID of the EFS file system"
  type        = string
}