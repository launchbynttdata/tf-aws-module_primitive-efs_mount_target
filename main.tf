resource "aws_efs_mount_target" "this" {
  for_each = { for idx, subnet_id in var.subnet_ids : subnet_id => subnet_id }

  file_system_id  = var.efs_filesystem_id
  subnet_id       = each.value
  security_groups = var.security_group_ids
}
