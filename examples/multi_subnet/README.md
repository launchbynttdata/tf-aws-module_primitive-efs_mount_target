# Multi-Subnet EFS Mount Target Example

This example demonstrates creating EFS mount targets across multiple availability zones with stable infrastructure when subnets are added or removed.

## Key Features

- **Multiple Availability Zones**: Creates mount targets in 3 different AZs
- **Stable Infrastructure**: Mount targets use subnet IDs as keys, preventing rebuilds when subnet list order changes
- **Flexible Testing**: Control which subnets have mount targets via `enabled_subnet_indices`
- **No Unnecessary Rebuilds**: Removing a subnet from the middle of the list doesn't force rebuilds of other mount targets

## Architecture

```
VPC (10.1.0.0/16)
├── Subnet 1 (10.1.1.0/24) - us-east-2a
│   └── Mount Target 1
├── Subnet 2 (10.1.2.0/24) - us-east-2b
│   └── Mount Target 2
└── Subnet 3 (10.1.3.0/24) - us-east-2c
    └── Mount Target 3
```

## Usage

### Deploy All Mount Targets

```bash
terraform init
terraform plan -var-file=test.tfvars
terraform apply -var-file=test.tfvars
```

### Test Subnet Removal (No Rebuild)

1. **Initial deployment** with all 3 subnets:
```bash
terraform apply -var-file=test.tfvars
```

2. **Remove the middle subnet** (index 1) - Mount targets in subnets 0 and 2 should NOT rebuild:
```bash
terraform apply -var-file=test.tfvars -var='enabled_subnet_indices=[0,2]'
```

3. **Verify no rebuilds** - Terraform should only destroy mount target in subnet 1, leaving others untouched:
```
Plan: 0 to add, 0 to change, 1 to destroy
```

4. **Add it back** - Only creates mount target in subnet 1:
```bash
terraform apply -var-file=test.tfvars -var='enabled_subnet_indices=[0,1,2]'
```

### Test Different Configurations

#### Only First Two Subnets
```bash
terraform apply -var-file=test.tfvars -var='enabled_subnet_indices=[0,1]'
```

#### Only Last Two Subnets
```bash
terraform apply -var-file=test.tfvars -var='enabled_subnet_indices=[1,2]'
```

#### Single Subnet
```bash
terraform apply -var-file=test.tfvars -var='enabled_subnet_indices=[0]'
```

## Testing from Go

### Example Test Scenarios

```go
// Test 1: Deploy all mount targets
func TestEFSMountTarget_AllSubnets(t *testing.T) {
    terraformOptions := &terraform.Options{
        TerraformDir: "../examples/multi_subnet",
        VarFiles:     []string{"test.tfvars"},
    }

    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)

    // Verify 3 mount targets created
    mountTargets := terraform.OutputMap(t, terraformOptions, "mount_target_ids")
    assert.Equal(t, 3, len(mountTargets))
}

// Test 2: Remove middle subnet without rebuilding others
func TestEFSMountTarget_RemoveMiddleSubnet_NoRebuild(t *testing.T) {
    terraformOptions := &terraform.Options{
        TerraformDir: "../examples/multi_subnet",
        VarFiles:     []string{"test.tfvars"},
    }

    // Initial deployment with all subnets
    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)

    initialOutputs := terraform.OutputMap(t, terraformOptions, "mount_target_ids")
    subnet0MountTarget := initialOutputs[getSubnetId(t, terraformOptions, 0)]
    subnet2MountTarget := initialOutputs[getSubnetId(t, terraformOptions, 2)]

    // Remove middle subnet
    terraformOptions.Vars = map[string]interface{}{
        "enabled_subnet_indices": []int{0, 2},
    }

    terraform.Apply(t, terraformOptions)

    // Verify mount targets for subnets 0 and 2 remain unchanged
    newOutputs := terraform.OutputMap(t, terraformOptions, "mount_target_ids")
    assert.Equal(t, subnet0MountTarget, newOutputs[getSubnetId(t, terraformOptions, 0)])
    assert.Equal(t, subnet2MountTarget, newOutputs[getSubnetId(t, terraformOptions, 2)])
    assert.Equal(t, 2, len(newOutputs)) // Only 2 mount targets now
}

// Test 3: Add subnet back
func TestEFSMountTarget_AddSubnetBack(t *testing.T) {
    terraformOptions := &terraform.Options{
        TerraformDir: "../examples/multi_subnet",
        VarFiles:     []string{"test.tfvars"},
        Vars: map[string]interface{}{
            "enabled_subnet_indices": []int{0, 2},
        },
    }

    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)

    // Add back middle subnet
    terraformOptions.Vars = map[string]interface{}{
        "enabled_subnet_indices": []int{0, 1, 2},
    }

    terraform.Apply(t, terraformOptions)

    // Verify 3 mount targets exist
    mountTargets := terraform.OutputMap(t, terraformOptions, "mount_target_ids")
    assert.Equal(t, 3, len(mountTargets))
}
```

## Important Notes

### Why Subnet IDs as Keys?

The module uses subnet IDs as Terraform resource keys:

```hcl
for_each = { for subnet_id in var.subnet_ids : subnet_id => subnet_id }
```

This approach ensures:
- **Stable resource addresses**: Each mount target's Terraform address is based on its subnet ID, not list position
- **No rebuilds on reordering**: Changing subnet order in the list doesn't affect existing mount targets
- **Predictable operations**: Adding/removing subnets only affects those specific mount targets

### Alternative Approach (Not Recommended)

Using list indices as keys would cause rebuilds:

```hcl
# DON'T DO THIS - causes rebuilds when list order changes
for_each = { for idx, subnet_id in var.subnet_ids : idx => subnet_id }
```

With this approach, removing subnet at index 1 would:
1. Destroy mount target at index 1
2. **Destroy and recreate** mount target at index 2 (because it becomes index 1)

## Variables

| Name | Description | Type | Default |
|------|-------------|------|---------|
| `project_name` | Project name for resource naming | `string` | `"efs-mt-multi"` |
| `environment` | Environment name | `string` | `"test"` |
| `vpc_cidr_block` | VPC CIDR block | `string` | `"10.1.0.0/16"` |
| `subnet_configs` | List of subnet configurations | `list(object)` | See test.tfvars |
| `enabled_subnet_indices` | Indices of subnets to enable mount targets in | `list(number)` | `null` (all) |
| `efs_encrypted` | Enable EFS encryption | `bool` | `true` |

## Outputs

| Name | Description |
|------|-------------|
| `mount_target_ids` | Map of subnet ID to mount target ID |
| `mount_target_dns_names` | Map of subnet ID to DNS name |
| `mount_target_availability_zones` | Map of subnet ID to AZ |
| `subnet_ids` | List of all created subnet IDs |
| `enabled_subnet_ids` | List of subnets with mount targets |

## Cleanup

```bash
terraform destroy -var-file=test.tfvars
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | ~> 1.5 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 5.100 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 5.100.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_aws_efs_file_system"></a> [aws\_efs\_file\_system](#module\_aws\_efs\_file\_system) | github.com/launchbynttdata/tf-aws-module_primitive-efs_file_system | 1.0.0 |
| <a name="module_efs_mount_target"></a> [efs\_mount\_target](#module\_efs\_mount\_target) | ../../ | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_default_security_group.default](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/default_security_group) | resource |
| [aws_security_group.efs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_subnet.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet) | resource |
| [aws_vpc.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_project_name"></a> [project\_name](#input\_project\_name) | Project name used for resource naming and tagging | `string` | `"efs-mt-multi"` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Environment name (e.g., dev, test, prod) | `string` | `"test"` | no |
| <a name="input_vpc_cidr_block"></a> [vpc\_cidr\_block](#input\_vpc\_cidr\_block) | CIDR block for the VPC | `string` | `"10.0.0.0/16"` | no |
| <a name="input_subnet_configs"></a> [subnet\_configs](#input\_subnet\_configs) | List of subnet configurations with CIDR blocks and availability zones | <pre>list(object({<br/>    cidr_block        = string<br/>    availability_zone = string<br/>  }))</pre> | n/a | yes |
| <a name="input_enabled_subnet_indices"></a> [enabled\_subnet\_indices](#input\_enabled\_subnet\_indices) | List of subnet indices (0-based) to create mount targets in.<br/>If null or empty, mount targets will be created in all subnets.<br/><br/>Example: [0, 2] creates mount targets only in the 1st and 3rd subnets.<br/>This allows testing subnet removal/addition without rebuilding existing mount targets. | `list(number)` | `null` | no |
| <a name="input_efs_encrypted"></a> [efs\_encrypted](#input\_efs\_encrypted) | Whether to enable encryption at rest for the EFS file system | `bool` | `true` | no |
| <a name="input_mount_target_create_timeout"></a> [mount\_target\_create\_timeout](#input\_mount\_target\_create\_timeout) | Timeout for creating EFS mount targets | `string` | `"30m"` | no |
| <a name="input_mount_target_delete_timeout"></a> [mount\_target\_delete\_timeout](#input\_mount\_target\_delete\_timeout) | Timeout for deleting EFS mount targets | `string` | `"10m"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | The ID of the VPC |
| <a name="output_subnet_ids"></a> [subnet\_ids](#output\_subnet\_ids) | List of all created subnet IDs |
| <a name="output_subnet_availability_zones"></a> [subnet\_availability\_zones](#output\_subnet\_availability\_zones) | Map of subnet IDs to their availability zones |
| <a name="output_enabled_mount_target_keys"></a> [enabled\_mount\_target\_keys](#output\_enabled\_mount\_target\_keys) | List of mount target keys that are enabled |
| <a name="output_enabled_subnet_ids"></a> [enabled\_subnet\_ids](#output\_enabled\_subnet\_ids) | Map of mount target key to subnet ID where mount targets are enabled |
| <a name="output_efs_file_system_id"></a> [efs\_file\_system\_id](#output\_efs\_file\_system\_id) | The ID of the EFS file system |
| <a name="output_efs_file_system_arn"></a> [efs\_file\_system\_arn](#output\_efs\_file\_system\_arn) | The ARN of the EFS file system |
| <a name="output_mount_target_ids"></a> [mount\_target\_ids](#output\_mount\_target\_ids) | Map of mount target key (e.g., 'az-a') to EFS mount target ID |
| <a name="output_mount_target_subnet_ids"></a> [mount\_target\_subnet\_ids](#output\_mount\_target\_subnet\_ids) | Map of mount target key to subnet ID |
| <a name="output_mount_target_dns_names"></a> [mount\_target\_dns\_names](#output\_mount\_target\_dns\_names) | Map of mount target key to EFS file system DNS name |
| <a name="output_mount_target_az_dns_names"></a> [mount\_target\_az\_dns\_names](#output\_mount\_target\_az\_dns\_names) | Map of mount target key to mount target AZ-specific DNS name |
| <a name="output_mount_target_network_interface_ids"></a> [mount\_target\_network\_interface\_ids](#output\_mount\_target\_network\_interface\_ids) | Map of mount target key to network interface ID |
| <a name="output_mount_target_availability_zones"></a> [mount\_target\_availability\_zones](#output\_mount\_target\_availability\_zones) | Map of mount target key to availability zone name |
| <a name="output_security_group_id"></a> [security\_group\_id](#output\_security\_group\_id) | The ID of the EFS security group |
<!-- END_TF_DOCS -->
