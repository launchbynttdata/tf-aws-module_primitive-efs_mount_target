# tf-aws-module_primitive-efs_mount_target

## What is a Primitive Module?

A **primitive module** is a thin, focused Terraform wrapper around a single AWS resource type. Primitive modules:

- Wrap a **single AWS resource** (e.g., `aws_eks_cluster`, `aws_kms_key`, `aws_s3_bucket`)
- Provide sensible defaults while maintaining full configurability
- Include comprehensive validation rules
- Follow consistent patterns for inputs, outputs, and tagging
- Include automated testing using Terratest
- Serve as building blocks for higher-level composite modules

For examples of well-structured primitive modules, see:

- [tf-aws-module_primitive-eks_cluster](https://github.com/launchbynttdata/tf-aws-module_primitive-eks_cluster)
- [tf-aws-module_primitive-kms_key](https://github.com/launchbynttdata/tf-aws-module_primitive-kms_key)

---

## Getting Started with This Template

### 1. Create Your New Module Repository

1. Click the "Use this template" button on GitHub
2. Name your repository following the naming convention: `tf-aws-module_primitive-<resource_name>`
   - Examples: `tf-aws-module_primitive-s3_bucket`, `tf-aws-module_primitive-lambda_function`
3. Clone your new repository locally

### 2. Initialize and Clean Up Template References

After cloning, run the cleanup target to update template references with your actual repository information:

```bash
make init-module
```

This command will:

- Update the `go.mod` file with your repository's GitHub URL
- Update test imports to reference your new module name
- Remove template-specific placeholders

### 3. Configure Your Environment

Install required development dependencies:

```bash
make configure-dependencies
make configure-git-hooks
```

This installs:

- Terraform
- Go
- Pre-commit hooks
- Other development tools specified in `.tool-versions`

---

## HOWTO: Developing a Primitive Module

### Step 1: Define Your Resource

1. **Identify the AWS resource** you're wrapping (e.g., `aws_eks_cluster`)
2. **Review AWS documentation** for the resource to understand all available parameters
3. **Study similar primitive modules** for patterns and best practices

### Step 2: Create the Module Structure

Your primitive module should include these core files:

#### `main.tf`

- Contains the primary resource declaration
- Should be clean and focused on the single resource
- Example:

```hcl
resource "aws_eks_cluster" "this" {
  name     = var.name
  role_arn = var.role_arn
  version  = var.kubernetes_version

  vpc_config {
    subnet_ids              = var.vpc_config.subnet_ids
    security_group_ids      = var.vpc_config.security_group_ids
    endpoint_private_access = var.vpc_config.endpoint_private_access
    endpoint_public_access  = var.vpc_config.endpoint_public_access
    public_access_cidrs     = var.vpc_config.public_access_cidrs
  }

  tags = merge(
    var.tags,
    local.default_tags
  )
}
```

#### `variables.tf`

- Define all configurable parameters
- Include clear descriptions for each variable
- Set sensible defaults where appropriate
- Use validation rules to enforce constraints, but only when the validations can be made precise.
- Alternatively, use [`check`](https://developer.hashicorp.com/terraform/language/block/check) blocks to create more complicated validations. (Requires terraform ~> 1.12)
- Example:

```hcl
variable "name" {
  description = "Name of the EKS cluster"
  type        = string

  validation {
    condition     = length(var.name) <= 100
    error_message = "Cluster name must be 100 characters or less"
  }
}

variable "kubernetes_version" {
  description = "Kubernetes version to use for the EKS cluster"
  type        = string
  default     = null

  validation {
    condition     = var.kubernetes_version == null || can(regex("^1\\.(2[89]|[3-9][0-9])$", var.kubernetes_version))
    error_message = "Kubernetes version must be 1.28 or higher"
  }
}
```

#### `outputs.tf`

- Export all useful attributes of the resource
- Include comprehensive outputs for downstream consumption
- Document what each output provides
- Example:

```hcl
output "id" {
  description = "The ID of the EKS cluster"
  value       = aws_eks_cluster.this.id
}

output "arn" {
  description = "The ARN of the EKS cluster"
  value       = aws_eks_cluster.this.arn
}

output "endpoint" {
  description = "The endpoint for the EKS cluster API server"
  value       = aws_eks_cluster.this.endpoint
}
```

#### `locals.tf`

- Define local values and transformations
- Include standard tags (e.g., `provisioner = "Terraform"`)
- Example:

```hcl
locals {
  default_tags = {
    provisioner = "Terraform"
  }
}
```

#### `versions.tf`

- Specify required Terraform and provider versions
- Example:

```hcl
terraform {
  required_version = "~> 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.100"
    }
  }
}
```

### Step 3: Create Examples

Create example configurations in the `examples/` directory:

#### `examples/simple/`

- Minimal, working configuration
- Uses only required variables
- Good for quick starts and basic testing

#### `examples/complete/`

- Comprehensive configuration showing all features
- Demonstrates advanced options
- Includes comments explaining choices

Each example should include:

- `main.tf` - The module invocation
- `variables.tf` - Example variables
- `outputs.tf` - Pass-through outputs
- `test.tfvars` - Test values for automated testing
- `README.md` - Documentation for the example

### Step 4: Write Tests

Update the test files in `tests/`:

#### `tests/testimpl/test_impl.go`

Write functional tests that verify:

- The resource is created successfully
- Resource properties match expectations
- Outputs are correct
- Integration with AWS SDK to verify actual state

#### `tests/testimpl/types.go`

Define the configuration structure for your tests:

```go
type ThisTFModuleConfig struct {
    Name              string `json:"name"`
    KubernetesVersion string `json:"kubernetes_version"`
    // ... other fields
}
```

#### `tests/post_deploy_functional/main_test.go`

- Update test names to match your module
- Configure test flags (e.g., idempotency settings)
- Adjust test context as needed

### Step 5: Update Documentation

1. **Update README.md** with:
   - Overview of the module
   - Feature list
   - Usage examples
   - Input/output documentation
   - Validation rules

2. **Document validation rules** clearly so users understand constraints.

### Step 6: Test Your Module

1. **Run local validation**:

```bash
make check
```

This runs:

- Terraform fmt, validate, and plan
- Go tests with Terratest
- Pre-commit hooks
- Security scans

1. **Test with real infrastructure**:

```bash
cd examples/simple
terraform init
terraform plan -var-file=test.tfvars -out=the.tfplan
terraform apply the.tfplan
```

1. **Verify outputs**:

```bash
terraform output
```

1. **Clean up**:

```bash
terraform destroy -var-file=test.tfvars
```

### Step 7: Document and Release

1. **Write a comprehensive README** following the pattern in the example modules
1. **Add files to commit** `git add .`
1. **Run pre-commit hooks manually** `pre-commit run`
1. **Resolve any pre-commit issues**
1. **Push branch to github**

---

## Module Best Practices

### Naming Conventions

- Repository: `tf-aws-module_primitive-<resource_name>`
- Resource identifier: Use `this` for the primary resource.
- Variables: Use snake_case.
- Match AWS resource parameter names where possible.

### Input Variables

- Provide sensible defaults when safe to do so.
- Use `null` as default for optional complex objects.
- Include validation rules with clear error messages.
- Group related parameters using object types.
- Document expected formats and constraints.

### Outputs

- Export all significant resource attributes.
- Use clear, descriptive output names.
- Include descriptions for all outputs.
- Consider downstream module needs.

### Tags

- Always include a `tags` variable, unless the resource does not support tags.
- Merge with `local.default_tags` including `provisioner = "Terraform"`.
- Use provider default tags when appropriate.

### Validation

- Validate input constraints at the variable level.
- Provide helpful error messages.
- Check for common misconfigurations.
- Validate relationships between variables.

### Testing

- Test the minimal example (required parameters only).
- Test the complete example (all features).
- Verify resource creation and properties.
- Test idempotency where applicable.
- Test validation rules by expecting failures.

### Documentation

- Clear overview of the module's purpose.
- Feature list highlighting key capabilities.
- Multiple usage examples (minimal and complete).
- Comprehensive input/output tables.
- Document validation rules and constraints.
- Include links to relevant AWS documentation.

---

## File Structure

After initialization, your module should have this structure:

```
tf-aws-module_primitive-<resource_name>/
├── .github/
│   └── workflows/          # CI/CD workflows
├── examples/
│   ├── simple/            # Minimal example
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── test.tfvars
│   │   └── README.md
│   └── complete/          # Comprehensive example
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       ├── test.tfvars
│       └── README.md
├── tests/
│   ├── post_deploy_functional/
│   │   └── main_test.go
│   ├── testimpl/
│   │   ├── test_impl.go
│   │   └── types.go
├── .gitignore
├── .pre-commit-config.yaml
├── .tool-versions
├── go.mod
├── go.sum
├── LICENSE
├── locals.tf
├── main.tf
├── Makefile
├── outputs.tf
├── README.md
├── variables.tf
└── versions.tf
```

---

## Common Makefile Targets

| Target | Description |
|--------|-------------|
| `make init-module` | Initialize new module from template (run once after creating from template) |
| `make configure-dependencies` | Install required development tools |
| `make configure-git-hooks` | Set up pre-commit hooks |
| `make check` | Run all validation and tests |
| `make configure` | Full setup (dependencies + hooks + repo sync) |
| `make clean` | Remove downloaded components |

---

## Getting Help

- Review example modules: [EKS Cluster](https://github.com/launchbynttdata/tf-aws-module_primitive-eks_cluster), [KMS Key](https://github.com/launchbynttdata/tf-aws-module_primitive-kms_key)
- Check the Launch Common Automation Framework documentation.
- Reach out to the platform team for guidance.

---

## Contributing

Follow the established patterns in existing primitive modules. All modules should:

- Pass `make check` validation.
- Include comprehensive tests.
- Follow naming conventions.
- Include clear documentation.
- Use semantic versioning.

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

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_hello_message"></a> [hello\_message](#input\_hello\_message) | A friendly hello message. | `string` | `"Hello, Terraform!"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_account_id"></a> [account\_id](#output\_account\_id) | n/a |
| <a name="output_arn"></a> [arn](#output\_arn) | n/a |
| <a name="output_hello_message"></a> [hello\_message](#output\_hello\_message) | n/a |
<!-- END_TF_DOCS -->
