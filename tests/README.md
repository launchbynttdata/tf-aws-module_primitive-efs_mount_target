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
- `test_impl.go`: Implements test functions with AWS API validation
- `types.go`: Defines the module-specific test configuration type

## Test Cases

The test suite validates both Terraform outputs and actual AWS resource state using AWS API calls for performance.

### TestSimpleExample
Tests the simple example with a single mount target:

1. **Output Validation**: Verifies all Terraform outputs are populated
   - Mount target ID, subnet ID, DNS names
   - Network interface ID, availability zone information
   - EFS file system ID and ARN

2. **AWS EFS API Validation**: Queries EFS API to verify:
   - Mount target exists and is in 'available' state
   - Properties match Terraform outputs (subnet ID, file system ID, etc.)
   - Has valid IP address and availability zone information

3. **AWS EC2 API Validation**: Queries EC2 API to verify:
   - Network interface exists and is 'in-use'
   - Associated with correct subnet
   - Has security groups attached
   - Has a private IP address

### TestMultiSubnetWithChanges
Tests the multi-subnet example with infrastructure stability validation:

**Stage 1: Deploy all 3 mount targets (az-a, az-b, az-c)**
- Verifies 3 mount targets created via Terraform outputs
- Stores initial mount target IDs for comparison
- Validates all mount targets are available via AWS EFS API

**Stage 2: Remove az-b - Verify az-a and az-c unchanged**
- Applies with `enabled_subnet_indices = [0, 2]`
- **Critical Test**: Asserts az-a and az-c IDs remain identical (no rebuild)
- Uses AWS API to confirm only 2 mount targets exist
- Validates az-b was removed, az-a and az-c are unchanged

**Stage 3: Re-add az-b and remove az-a - Verify az-c unchanged**
- Applies with `enabled_subnet_indices = [1, 2]`
- **Critical Test**: Asserts az-c ID remains identical (never rebuilt across all changes)
- Verifies az-b has a NEW ID (was recreated)
- Uses AWS API to confirm final state
- **Test passes when az-c remains stable throughout all changes**

### Helper Functions

- **GetMountTargetsByFileSystem**: Retrieves all mount targets for an EFS file system via AWS API (faster than terraform output)
- **GetAWSConfig**: Creates AWS SDK configuration with proper region
- **GetRegionFromTerraform**: Extracts AWS region from Terraform outputs or variables

## Running Tests

### Prerequisites

1. **Install dependencies**:
   ```bash
   cd /workspace
   go mod download
   ```

2. **AWS credentials**: Configure AWS credentials with permissions to:
   - Create/read/delete EFS mount targets and file systems
   - Create/read/delete VPCs, subnets, and security groups
   - Describe EC2 network interfaces and security groups
   - Read EFS file systems

3. **Example configuration**:
   - `examples/simple/test.tfvars` - Single mount target example
   - `examples/multi_subnet/test.tfvars` - Multiple mount targets with dynamic changes
   - Set `region` to your target AWS region (e.g., `us-west-2`)
   - The full AZ names are automatically constructed as `{region}{availability_zone_letter}`

### Running Full Integration Tests

From the repository root:

```bash
# Run all post_deploy_functional tests (both examples, full lifecycle)
go test -v -timeout 45m ./tests/post_deploy_functional/

# Run specific test only
go test -v -timeout 30m ./tests/post_deploy_functional/ -run TestSimpleExample
go test -v -timeout 45m ./tests/post_deploy_functional/ -run TestMultiSubnetExample
```

**Expected flow**:
1. Terraform applies the example configuration
2. Tests execute against deployed resources using AWS API
3. For multi-subnet test: applies multiple terraform configurations to test stability
4. Terraform destroys all resources
5. Test results are displayed with detailed AWS API validation

### Running Read-Only Tests

Useful when you have already deployed infrastructure:

```bash
# Run against existing simple example infrastructure
go test -v -timeout 10m ./tests/post_deploy_functional_readonly/ -run TestSimpleExampleReadOnly

# Run against existing multi-subnet infrastructure
go test -v -timeout 15m ./tests/post_deploy_functional_readonly/ -run TestMultiSubnetExampleReadOnly
```

**Note**: Ensure Terraform state exists in the example directories before running read-only tests.

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

1. **Edit** `testimpl/test_impl.go`
2. **Create** a new test function following the pattern of existing tests
3. **Use** AWS API calls for validation (faster than terraform output)
4. **Add** test invocation in `post_deploy_functional/main_test.go` and `post_deploy_functional_readonly/main_test.go`

Example:

```go
// In testimpl/test_impl.go
func TestMyNewExample(t *testing.T, ctx testTypes.TestContext) {
    t.Log("=== Testing My New Example ===")

    opts := ctx.TerratestTerraformOptions()

    // Get Terraform outputs
    mountTargetID := terraform.Output(t, opts, "mount_target_id")

    // Validate via AWS API
    region := "us-west-2" // or extract from outputs
    awsConfig := GetAWSConfig(t, region)
    efsClient := efs.NewFromConfig(awsConfig)

    input := &efs.DescribeMountTargetsInput{
        MountTargetId: aws.String(mountTargetID),
    }
    result, err := efsClient.DescribeMountTargets(context.TODO(), input)
    require.NoError(t, err)
    require.Len(t, result.MountTargets, 1)

    // Add assertions
    mt := result.MountTargets[0]
    assert.Equal(t, "available", string(mt.LifeCycleState))
}
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

1. **Always use timeouts**: EFS resources can take time to provision (use `-timeout 30m` or more)
2. **Use AWS API for validation**: Much faster than `terraform output` commands
3. **Test actual AWS state**: Terraform outputs can be cached; AWS API returns real-time state
4. **Clean up resources**: Framework handles cleanup automatically
5. **Verify lifecycle states**: Check resources are 'available', 'in-use', etc.
6. **Test infrastructure stability**: Verify resource IDs don't change when configuration changes
7. **Use GetMountTargetsByFileSystem**: Retrieve all mount targets in one API call instead of individual queries

## Additional Resources

- [Terratest Documentation](https://terratest.gruntwork.io/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [lcaf-component-terratest](https://github.com/launchbynttdata/lcaf-component-terratest)
