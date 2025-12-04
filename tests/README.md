# Test Suite Documentation

This directory contains the automated test suite for the EFS Mount Target Terraform module using [Terratest](https://terratest.gruntwork.io/).

## Test Structure

The test suite follows a modular pattern with three main components:

```
tests/
├── post_deploy_functional/        # Full lifecycle tests (deploy → test → destroy)
│   └── main_test.go
├── post_deploy_functional_readonly/ # Tests on pre-existing infrastructure
│   └── main_test.go
└── testimpl/                       # Shared test implementation
    ├── test_impl.go               # Test logic
    └── types.go                   # Type definitions
```

### Test Folders

#### `post_deploy_functional/`
**Purpose**: Full end-to-end integration tests that deploy infrastructure, run validations, and clean up.

**What it does**:
1. Runs `terraform init` and `terraform apply` using the example configuration
2. Executes all test cases defined in `testimpl/test_impl.go`
3. Runs `terraform destroy` to clean up resources

**When to use**:
- During development to validate changes
- In CI/CD pipelines for pull requests
- Before releasing new versions

#### `post_deploy_functional_readonly/`
**Purpose**: Run tests against already-deployed infrastructure without managing the lifecycle.

**What it does**:
1. Connects to existing Terraform state
2. Executes all test cases without deploying or destroying resources

**When to use**:
- Testing against long-lived environments
- Debugging issues in deployed infrastructure
- Running tests multiple times without redeployment

#### `testimpl/`
**Purpose**: Contains the actual test logic shared by both test folders.

**Components**:
- `test_impl.go`: Implements `TestComposableComplete()` function with all test cases
- `types.go`: Defines the module-specific test configuration type

## Test Cases

The test suite validates both Terraform outputs and actual AWS resource state:

### Output Validation Tests
1. **TestMountTargetIDs**: Verifies mount target IDs are not empty
2. **TestMountTargetDNSNames**: Validates DNS names contain `.efs.` and are properly formatted
3. **TestMountTargetNetworkInterfaceIDs**: Confirms network interface IDs exist

### AWS Resource Tests
4. **TestMountTargetExistsInAWS**: Queries EFS API to verify:
   - Mount target exists and is in 'available' state
   - Properties match expected values (subnet ID, file system ID, etc.)
   - Has valid IP address and availability zone information

5. **TestNetworkInterfaceExistsInAWS**: Queries EC2 API to verify:
   - Network interface exists and is 'in-use'
   - Associated with correct subnet
   - Has security groups attached
   - Has a private IP address

6. **TestMountTargetDNSResolution**: Validates DNS naming convention:
   - Format: `fs-xxxxxxxx.efs.region.amazonaws.com`

7. **TestMountTargetSecurityGroups**: Verifies:
   - Security groups are attached to the network interface
   - Each security group exists and is properly configured

## Running Tests

### Prerequisites

1. **Install dependencies**:
   ```bash
   cd /workspace
   go mod download
   ```

2. **AWS credentials**: Configure AWS credentials with permissions to:
   - Create/read/delete EFS mount targets
   - Describe EC2 network interfaces and security groups
   - Read EFS file systems

3. **Example configuration**: Ensure `examples/simple/test.tfvars` contains valid values
   - Set `region` to your target AWS region (e.g., `us-east-2`)
   - Set `availability_zone_letter` to a valid AZ suffix (e.g., `a`, `b`, or `c`)
   - The full AZ name will be automatically constructed as `{region}{availability_zone_letter}`

### Running Full Integration Tests

From the repository root:

```bash
# Run post_deploy_functional tests (full lifecycle)
go test -v -timeout 30m ./tests/post_deploy_functional/

# Run with specific example folder
cd tests/post_deploy_functional
go test -v -timeout 30m
```

**Expected flow**:
1. Terraform applies the simple configuration
2. Tests execute against deployed resources
3. Terraform destroys all resources
4. Test results are displayed

### Running Read-Only Tests

Useful when you have already deployed infrastructure:

```bash
# Run against existing infrastructure
go test -v -timeout 10m ./tests/post_deploy_functional_readonly/
```

**Note**: Ensure Terraform state exists in simple before running.

### Running Specific Test Cases

Run only specific subtests:

```bash
go test -v -timeout 30m ./tests/post_deploy_functional/ \
  -run TestModule/TestMountTargetExistsInAWS
```

### Test Verbosity

Control output detail:

```bash
# Minimal output
go test ./tests/post_deploy_functional/

# Verbose output (recommended)
go test -v ./tests/post_deploy_functional/

# Very verbose with Terraform output
TF_LOG=DEBUG go test -v ./tests/post_deploy_functional/
```

## Test Configuration

Tests use the `lcaf-component-terratest` framework which provides:

- Terraform lifecycle management
- State handling
- Output retrieval
- Error handling and cleanup

### Configuration Parameters

Set in main_test.go:

```go
const (
    testConfigsExamplesFolderDefault = "../../examples"  // Path to examples
    infraTFVarFileNameDefault        = "test.tfvars"     // Vars file name
)
```

### Environment Variables

Control test behavior via environment variables:

```bash
# Show Terraform debug logs
export TF_LOG=DEBUG

# Set AWS region (examples will construct full AZ names)
export AWS_REGION=us-east-2

# Use specific AWS profile
export AWS_PROFILE=my-profile
```

### Testing Across Different Regions

The examples support region-aware configuration. To test in different regions:

```bash
# Test in us-west-2
terraform apply -var-file=test.tfvars -var='region=us-west-2'

# Test in eu-west-1
terraform apply -var-file=test.tfvars -var='region=eu-west-1'
```

The availability zone names are automatically constructed from the region and AZ letter suffix (e.g., `us-west-2` + `a` = `us-west-2a`)

## Debugging Tests

### View Terraform Plans

```bash
cd examples/simple
terraform init
terraform plan -var-file=test.tfvars
```

### Preserve Test Resources

To keep resources for inspection, comment out the cleanup in test code or use `-debug`:

```bash
# Skip cleanup on failure
go test -v -timeout 30m ./tests/post_deploy_functional/ -failfast
```

Then manually inspect:
```bash
cd examples/simple
terraform show
aws efs describe-mount-targets --mount-target-id <id>
```

### Common Issues

1. **Timeout errors**: Increase timeout with `-timeout 45m`
2. **State locking**: Ensure no other processes are using the state
3. **Resource already exists**: Run `terraform destroy` in simple
4. **Permission errors**: Verify AWS credentials have required permissions

## Adding New Tests

To add new test cases:

1. **Edit** test_impl.go
2. **Add** a new `t.Run()` block in `TestComposableComplete()`
3. **Use** the existing AWS clients (`efsClient`, `ec2Client`)
4. **Follow** the pattern of existing tests

Example:

```go
t.Run("TestMyNewValidation", func(t *testing.T) {
    for subnet, mountTargetID := range mountTargetIDs {
        // Your test logic here
        input := &efs.DescribeMountTargetsInput{
            MountTargetId: aws.String(mountTargetID),
        }
        result, err := efsClient.DescribeMountTargets(context.TODO(), input)
        require.NoError(t, err)

        // Add assertions
        assert.NotNil(t, result)
    }
})
```

## CI/CD Integration

Tests run automatically on:

- **Pull Requests**: Via pull-request-terraform-check.yml
- **Pre-commit**: Linting via `golangci-lint` (configured in .golangci.yaml)

### GitHub Actions

The workflow runs:
1. Terraform formatting check
2. Terraform validation
3. Go tests in `post_deploy_functional/`

Configure in repository settings:
- `TERRAFORM_CHECK_AWS_ASSUME_ROLE_ARN`: IAM role for tests
- `TERRAFORM_CHECK_AWS_REGION`: Target region

## Best Practices

1. **Always use timeouts**: EFS resources can take time to provision
2. **Test both outputs and AWS state**: Outputs can be misleading
3. **Clean up resources**: Use `defer` or framework cleanup
4. **Use subtests**: Makes it easier to identify failures
5. **Verify states**: Check resources are 'available', 'in-use', etc.
6. **Test security**: Verify security group attachments

## Additional Resources

- [Terratest Documentation](https://terratest.gruntwork.io/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [lcaf-component-terratest](https://github.com/launchbynttdata/lcaf-component-terratest)
