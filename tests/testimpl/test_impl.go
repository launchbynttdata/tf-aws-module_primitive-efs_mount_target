package testimpl

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gruntwork-io/terratest/modules/terraform"
	testTypes "github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiSubnetWithChanges tests the multi-subnet example with dynamic subnet changes
// This validates that mount targets are not rebuilt when subnet configurations change
// Test scenario:
// 1. Deploy all 3 mount targets (az-a, az-b, az-c)
// 2. Remove az-b - verify az-a and az-c are unchanged
// 3. Re-add az-b and remove az-a - verify az-c is unchanged
func TestMultiSubnetWithChanges(t *testing.T, ctx testTypes.TestContext) {
	t.Log("=======================================================")
	t.Log("=== Stage 1: Deploy all 3 mount targets ===")
	t.Log("=======================================================")

	opts := ctx.TerratestTerraformOptions()

	// Get AWS config for API calls
	awsConfig := GetAWSConfig(t, "us-west-2") // Will use region from test config
	efsClient := efs.NewFromConfig(awsConfig)

	// Initial deployment with all subnets - get outputs from Terraform
	mountTargetIDsOutput := terraform.OutputMap(t, opts, "mount_target_ids")
	efsFileSystemID := terraform.Output(t, opts, "efs_file_system_id")
	efsFileSystemARN := terraform.Output(t, opts, "efs_file_system_arn")

	// Verify all 3 mount targets are created
	assert.Len(t, mountTargetIDsOutput, 3, "Should have 3 mount targets initially")
	assert.Contains(t, mountTargetIDsOutput, "az-a", "Should have mount target for az-a")
	assert.Contains(t, mountTargetIDsOutput, "az-b", "Should have mount target for az-b")
	assert.Contains(t, mountTargetIDsOutput, "az-c", "Should have mount target for az-c")
	assert.NotEmpty(t, efsFileSystemID, "EFS file system ID should not be empty")
	assert.NotEmpty(t, efsFileSystemARN, "EFS file system ARN should not be empty")

	// Store initial mount target IDs from Terraform output
	initialMountTargetAZ_A := mountTargetIDsOutput["az-a"]
	initialMountTargetAZ_B := mountTargetIDsOutput["az-b"]
	initialMountTargetAZ_C := mountTargetIDsOutput["az-c"]

	t.Logf("Initial mount target IDs from Terraform output:")
	t.Logf("  az-a: %s", initialMountTargetAZ_A)
	t.Logf("  az-b: %s", initialMountTargetAZ_B)
	t.Logf("  az-c: %s", initialMountTargetAZ_C)

	// Get full mount target details from AWS API (not Terraform output)
	// This is faster than terraform output and provides real-time AWS state
	stage1MountTargets := GetMountTargetsByFileSystem(t, efsFileSystemID, efsClient)
	assert.Len(t, stage1MountTargets, 3, "Should have 3 mount targets in AWS")

	// Verify all mount targets are available in AWS
	for _, mt := range stage1MountTargets {
		mtID := aws.ToString(mt.MountTargetId)
		assert.Equal(t, "available", string(mt.LifeCycleState),
			"Mount target %s should be available in AWS", mtID)
		t.Logf("✅ Mount target %s validated in AWS: IP=%s, SubnetID=%s, AZ=%s",
			mtID, aws.ToString(mt.IpAddress), aws.ToString(mt.SubnetId), aws.ToString(mt.AvailabilityZoneName))
	}

	t.Log("=======================================================")
	t.Log("=== Stage 2: Remove az-b - Verify az-a and az-c unchanged ===")
	t.Log("=======================================================")

	opts.SetVarsAfterVarFiles = true
	opts.Vars = map[string]interface{}{
		"enabled_subnet_indices": []int{0, 2}, // Keep indices 0 (az-a) and 2 (az-c), remove 1 (az-b)
	}

	// Apply with the new configuration
	terraform.Apply(t, opts)

	// Get mount target IDs from Terraform output
	mountTargetIDsOutput = terraform.OutputMap(t, opts, "mount_target_ids")

	// Verify only 2 mount targets remain in Terraform output
	assert.Len(t, mountTargetIDsOutput, 2, "Should have 2 mount targets after removing az-b")
	assert.Contains(t, mountTargetIDsOutput, "az-a", "Should still have mount target for az-a")
	assert.NotContains(t, mountTargetIDsOutput, "az-b", "Should NOT have mount target for az-b")
	assert.Contains(t, mountTargetIDsOutput, "az-c", "Should still have mount target for az-c")

	// CRITICAL TEST: Verify IDs haven't changed (no rebuild)
	assert.Equal(t, initialMountTargetAZ_A, mountTargetIDsOutput["az-a"],
		"Mount target az-a should NOT be rebuilt (ID should remain %s)", initialMountTargetAZ_A)
	assert.Equal(t, initialMountTargetAZ_C, mountTargetIDsOutput["az-c"],
		"Mount target az-c should NOT be rebuilt (ID should remain %s)", initialMountTargetAZ_C)

	t.Logf("✅ SUCCESS: Mount targets az-a (%s) and az-c (%s) were NOT rebuilt",
		mountTargetIDsOutput["az-a"], mountTargetIDsOutput["az-c"])

	// Verify in AWS using API (faster than terraform output)
	stage2MountTargets := GetMountTargetsByFileSystem(t, efsFileSystemID, efsClient)
	assert.Len(t, stage2MountTargets, 2, "Should have 2 mount targets in AWS after removing az-b")

	// Validate the remaining mount targets are available and match expected IDs
	stage2MountTargetIDs := make(map[string]bool)
	for _, mt := range stage2MountTargets {
		mtID := aws.ToString(mt.MountTargetId)
		stage2MountTargetIDs[mtID] = true
		assert.Equal(t, "available", string(mt.LifeCycleState),
			"Mount target %s should be available in AWS", mtID)
	}

	// Verify az-a and az-c still exist with same IDs in AWS
	assert.True(t, stage2MountTargetIDs[initialMountTargetAZ_A],
		"Mount target az-a (%s) should still exist in AWS", initialMountTargetAZ_A)
	assert.True(t, stage2MountTargetIDs[initialMountTargetAZ_C],
		"Mount target az-c (%s) should still exist in AWS", initialMountTargetAZ_C)
	assert.False(t, stage2MountTargetIDs[initialMountTargetAZ_B],
		"Mount target az-b (%s) should NOT exist in AWS", initialMountTargetAZ_B)

	t.Log("✅ AWS API confirmed: az-b removed, az-a and az-c unchanged")

	t.Log("=======================================================")
	t.Log("=== Stage 3: Re-add az-b and remove az-a - Verify az-c unchanged ===")
	t.Log("=======================================================")

	opts.Vars = map[string]interface{}{
		"enabled_subnet_indices": []int{1, 2}, // Keep indices 1 (az-b) and 2 (az-c), remove 0 (az-a)
	}

	// Apply with the new configuration
	terraform.Apply(t, opts)

	// Get mount target IDs from Terraform output
	mountTargetIDsOutput = terraform.OutputMap(t, opts, "mount_target_ids")

	// Verify only 2 mount targets remain in Terraform output
	assert.Len(t, mountTargetIDsOutput, 2, "Should have 2 mount targets after re-adding az-b and removing az-a")
	assert.NotContains(t, mountTargetIDsOutput, "az-a", "Should NOT have mount target for az-a")
	assert.Contains(t, mountTargetIDsOutput, "az-b", "Should have mount target for az-b")
	assert.Contains(t, mountTargetIDsOutput, "az-c", "Should still have mount target for az-c")

	// Store new az-b ID (it was recreated)
	newMountTargetAZ_B := mountTargetIDsOutput["az-b"]

	// CRITICAL TEST: Verify az-c ID hasn't changed (no rebuild)
	assert.Equal(t, initialMountTargetAZ_C, mountTargetIDsOutput["az-c"],
		"Mount target az-c should STILL NOT be rebuilt (ID should remain %s)", initialMountTargetAZ_C)

	// Verify az-b has a NEW ID (it was recreated)
	assert.NotEqual(t, initialMountTargetAZ_B, newMountTargetAZ_B,
		"Mount target az-b should have a NEW ID (was recreated)")

	t.Logf("✅ SUCCESS: Mount target az-c (%s) was NOT rebuilt, az-b recreated with new ID (%s)",
		mountTargetIDsOutput["az-c"], newMountTargetAZ_B)

	// Verify in AWS using API
	stage3MountTargets := GetMountTargetsByFileSystem(t, efsFileSystemID, efsClient)
	assert.Len(t, stage3MountTargets, 2, "Should have 2 mount targets in AWS")

	// Validate the mount targets are available and match expected IDs
	stage3MountTargetIDs := make(map[string]bool)
	for _, mt := range stage3MountTargets {
		mtID := aws.ToString(mt.MountTargetId)
		stage3MountTargetIDs[mtID] = true
		assert.Equal(t, "available", string(mt.LifeCycleState),
			"Mount target %s should be available in AWS", mtID)
		t.Logf("  Mount target in AWS: %s (Subnet: %s, AZ: %s)",
			mtID, aws.ToString(mt.SubnetId), aws.ToString(mt.AvailabilityZoneName))
	}

	// Verify az-c still exists with same ID, az-a removed, az-b recreated
	assert.False(t, stage3MountTargetIDs[initialMountTargetAZ_A],
		"Mount target az-a (%s) should NOT exist in AWS", initialMountTargetAZ_A)
	assert.True(t, stage3MountTargetIDs[newMountTargetAZ_B],
		"New mount target az-b (%s) should exist in AWS", newMountTargetAZ_B)
	assert.True(t, stage3MountTargetIDs[initialMountTargetAZ_C],
		"Mount target az-c (%s) should still exist in AWS with same ID", initialMountTargetAZ_C)

	t.Log("✅ AWS API confirmed: az-a removed, az-b recreated, az-c unchanged")
	t.Log("=======================================================")
	t.Log("=== TEST PASSED: All mount target lifecycle changes validated ===")
	t.Log("=== Summary: ===")
	t.Logf("  - Initial: Created 3 mount targets (az-a, az-b, az-c)")
	t.Logf("  - Stage 2: Removed az-b, az-a and az-c unchanged")
	t.Logf("  - Stage 3: Re-added az-b, removed az-a, az-c unchanged")
	t.Logf("  - az-c (%s) was NEVER rebuilt across all changes", initialMountTargetAZ_C)
	t.Log("=======================================================")
}

// GetMountTargetsByFileSystem retrieves all mount targets for a given EFS file system from AWS
func GetMountTargetsByFileSystem(t *testing.T, fileSystemID string, efsClient *efs.Client) []types.MountTargetDescription {
	input := &efs.DescribeMountTargetsInput{
		FileSystemId: aws.String(fileSystemID),
	}

	result, err := efsClient.DescribeMountTargets(context.TODO(), input)
	require.NoError(t, err, "Failed to describe mount targets for file system %s", fileSystemID)

	return result.MountTargets
}

// TestSimpleExample tests the simple example with a single mount target
func TestSimpleExample(t *testing.T, ctx testTypes.TestContext) {
	t.Log("=== Testing Simple Example (Single Mount Target) ===")

	opts := ctx.TerratestTerraformOptions()

	// Get all outputs - now direct values instead of maps
	mountTargetID := terraform.Output(t, opts, "mount_target_id")
	mountTargetSubnetID := terraform.Output(t, opts, "mount_target_subnet_id")
	mountTargetDNSName := terraform.Output(t, opts, "mount_target_dns_name")
	mountTargetAZDNSName := terraform.Output(t, opts, "mount_target_az_dns_name")
	mountTargetNetworkInterfaceID := terraform.Output(t, opts, "mount_target_network_interface_id")
	mountTargetAZName := terraform.Output(t, opts, "mount_target_availability_zone_name")
	mountTargetAZID := terraform.Output(t, opts, "mount_target_availability_zone_id")
	efsFileSystemID := terraform.Output(t, opts, "efs_file_system_id")
	efsFileSystemARN := terraform.Output(t, opts, "efs_file_system_arn")

	// Verify all outputs are populated
	assert.NotEmpty(t, mountTargetID, "Mount target ID should not be empty")
	assert.NotEmpty(t, mountTargetSubnetID, "Subnet ID should not be empty")
	assert.NotEmpty(t, mountTargetDNSName, "DNS name should not be empty")
	assert.NotEmpty(t, mountTargetAZDNSName, "AZ-specific DNS name should not be empty")
	assert.NotEmpty(t, mountTargetNetworkInterfaceID, "Network interface ID should not be empty")
	assert.NotEmpty(t, mountTargetAZName, "Availability zone name should not be empty")
	assert.NotEmpty(t, mountTargetAZID, "Availability zone ID should not be empty")
	assert.NotEmpty(t, efsFileSystemID, "EFS file system ID should not be empty")
	assert.NotEmpty(t, efsFileSystemARN, "EFS file system ARN should not be empty")

	// Log output values
	t.Logf("Mount target ID: %s", mountTargetID)
	t.Logf("Subnet ID: %s", mountTargetSubnetID)
	t.Logf("DNS name: %s", mountTargetDNSName)
	t.Logf("AZ-specific DNS name: %s", mountTargetAZDNSName)
	t.Logf("Network interface ID: %s", mountTargetNetworkInterfaceID)
	t.Logf("Availability zone: %s (%s)", mountTargetAZName, mountTargetAZID)
	t.Logf("EFS file system: %s (%s)", efsFileSystemID, efsFileSystemARN)

	// Validate mount target exists in AWS using AWS API
	region := mountTargetAZName[:len(mountTargetAZName)-1] // Extract region from AZ name
	awsConfig := GetAWSConfig(t, region)
	efsClient := efs.NewFromConfig(awsConfig)
	ec2Client := ec2.NewFromConfig(awsConfig)

	t.Logf("Validating mount target %s exists in AWS via API", mountTargetID)

	// Validate mount target via EFS API
	mtInput := &efs.DescribeMountTargetsInput{
		MountTargetId: aws.String(mountTargetID),
	}

	mtResult, err := efsClient.DescribeMountTargets(context.TODO(), mtInput)
	require.NoError(t, err, "Failed to describe mount target via AWS API")
	require.Len(t, mtResult.MountTargets, 1, "Should return exactly one mount target from AWS API")

	mountTarget := mtResult.MountTargets[0]
	assert.Equal(t, mountTargetID, *mountTarget.MountTargetId, "Mount target ID from AWS should match Terraform output")
	assert.Equal(t, mountTargetSubnetID, *mountTarget.SubnetId, "Subnet ID from AWS should match Terraform output")
	assert.Equal(t, efsFileSystemID, *mountTarget.FileSystemId, "File system ID from AWS should match Terraform output")
	assert.Equal(t, "available", string(mountTarget.LifeCycleState), "Mount target should be available in AWS")
	assert.NotEmpty(t, *mountTarget.IpAddress, "Mount target should have an IP address")
	assert.NotEmpty(t, *mountTarget.AvailabilityZoneId, "Mount target should have an availability zone ID")
	assert.NotEmpty(t, *mountTarget.AvailabilityZoneName, "Mount target should have an availability zone name")
	t.Logf("✅ Mount target validated in AWS: ID=%s, IP=%s, State=%s",
		*mountTarget.MountTargetId, *mountTarget.IpAddress, mountTarget.LifeCycleState)

	// Validate network interface exists via EC2 API
	t.Logf("Validating network interface %s exists in AWS via EC2 API", mountTargetNetworkInterfaceID)

	niInput := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{mountTargetNetworkInterfaceID},
	}

	niResult, err := ec2Client.DescribeNetworkInterfaces(context.TODO(), niInput)
	require.NoError(t, err, "Failed to describe network interface via AWS EC2 API")
	require.Len(t, niResult.NetworkInterfaces, 1, "Should return exactly one network interface from AWS API")

	netInterface := niResult.NetworkInterfaces[0]
	assert.Equal(t, mountTargetNetworkInterfaceID, *netInterface.NetworkInterfaceId, "Network interface ID from AWS should match Terraform output")
	assert.Equal(t, mountTargetSubnetID, *netInterface.SubnetId, "Network interface subnet ID should match")
	assert.Equal(t, "in-use", string(netInterface.Status), "Network interface should be in-use")
	assert.NotEmpty(t, *netInterface.PrivateIpAddress, "Network interface should have a private IP")
	require.NotEmpty(t, netInterface.Groups, "Network interface should have security groups attached")
	t.Logf("✅ Network interface validated in AWS: ID=%s, IP=%s, Status=%s, SecurityGroups=%d",
		*netInterface.NetworkInterfaceId, *netInterface.PrivateIpAddress, netInterface.Status, len(netInterface.Groups))

	t.Log("✅ Simple example validated successfully in both Terraform outputs and AWS API")
}

func GetAWSSTSClient(t *testing.T, region string) *sts.Client {
	awsSTSClient := sts.NewFromConfig(GetAWSConfig(t, region))
	return awsSTSClient
}

func GetAWSConfig(t *testing.T, region string) (cfg aws.Config) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	require.NoErrorf(t, err, "unable to load SDK config, %v", err)
	t.Logf("AWS SDK configured for region: %s", region)
	return cfg
}

// GetRegionFromTerraform extracts the AWS region from Terraform configuration
// This function works with both simple and multi-subnet examples
func GetRegionFromTerraform(t *testing.T, opts *terraform.Options) string {
	outputs := terraform.OutputAll(t, opts)

	// Try to get region from availability zone names
	// Simple example: mount_target_availability_zone_name (single string)
	// Multi-subnet example: mount_target_availability_zones (map)
	if azName, exists := outputs["mount_target_availability_zone_name"]; exists {
		// Simple example - direct string output
		azStr := azName.(string)
		if len(azStr) > 0 {
			region := azStr[:len(azStr)-1]
			t.Logf("Detected region %s from availability zone %s", region, azStr)
			return region
		}
	}

	if azNamesOutput, exists := outputs["mount_target_availability_zones"]; exists {
		// Multi-subnet example - map output
		azNames := azNamesOutput.(map[string]interface{})
		for _, azVal := range azNames {
			az := azVal.(string)
			if len(az) > 0 {
				region := az[:len(az)-1]
				t.Logf("Detected region %s from availability zone %s", region, az)
				return region
			}
		}
	}

	// Fallback: try to get region from vars or return default
	if region, exists := opts.Vars["region"]; exists {
		regionStr := region.(string)
		t.Logf("Using region %s from terraform vars", regionStr)
		return regionStr
	}

	// If all else fails, return us-west-2 as default (matches test.tfvars)
	t.Logf("WARNING: Could not detect region from Terraform, using default us-west-2")
	return "us-west-2"
}
