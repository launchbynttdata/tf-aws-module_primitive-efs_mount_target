package testimpl

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gruntwork-io/terratest/modules/terraform"
	testTypes "github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposableComplete(t *testing.T, ctx testTypes.TestContext) {
	opts := ctx.TerratestTerraformOptions()

	// Get outputs from Terraform - using correct output names for map-based module
	mountTargetIDs := terraform.OutputMap(t, opts, "mount_target_ids")
	mountTargetDNSNames := terraform.OutputMap(t, opts, "mount_target_dns_names")
	mountTargetNetworkInterfaceIDs := terraform.OutputMap(t, opts, "mount_target_network_interface_ids")

	// Initialize AWS clients with correct region
	region := GetRegionFromTerraform(t, opts)
	awsConfig := GetAWSConfig(t, region)
	efsClient := efs.NewFromConfig(awsConfig)
	ec2Client := ec2.NewFromConfig(awsConfig)

	t.Run("TestMountTargetIDs", func(t *testing.T) {
		for subnetID, mountTargetID := range mountTargetIDs {
			assert.NotEmpty(t, mountTargetID, "Mount target ID for subnet %s should not be empty", subnetID)
			t.Logf("Found mount target %s in subnet %s", mountTargetID, subnetID)
		}
	})

	t.Run("TestMountTargetDNSNames", func(t *testing.T) {
		for subnetID, dns := range mountTargetDNSNames {
			assert.NotEmpty(t, dns, "Mount target DNS name for subnet %s should not be empty", subnetID)
			assert.Contains(t, dns, ".efs.", "DNS name for subnet %s should contain '.efs.'", subnetID)
			t.Logf("Found DNS name %s for subnet %s", dns, subnetID)
		}
	})

	t.Run("TestMountTargetNetworkInterfaceIDs", func(t *testing.T) {
		for subnetID, ni := range mountTargetNetworkInterfaceIDs {
			assert.NotEmpty(t, ni, "Network interface ID for subnet %s should not be empty", subnetID)
			t.Logf("Found network interface %s in subnet %s", ni, subnetID)
		}
	})

	t.Run("TestMountTargetExistsInAWS", func(t *testing.T) {
		for subnetID, mountTargetID := range mountTargetIDs {
			t.Logf("Validating mount target %s in subnet %s via AWS EFS API", mountTargetID, subnetID)

			// Describe the mount target using AWS EFS API to validate it exists
			input := &efs.DescribeMountTargetsInput{
				MountTargetId: aws.String(mountTargetID),
			}

			result, err := efsClient.DescribeMountTargets(context.TODO(), input)
			require.NoError(t, err, "Failed to describe mount target %s for subnet %s via AWS API", mountTargetID, subnetID)
			require.NotNil(t, result, "DescribeMountTargets result should not be nil")
			require.Len(t, result.MountTargets, 1, "Should return exactly one mount target from AWS API")

			mountTarget := result.MountTargets[0]

			// Verify mount target properties from AWS match Terraform outputs
			assert.Equal(t, mountTargetID, *mountTarget.MountTargetId, "Mount target ID from AWS should match Terraform output")
			assert.Equal(t, subnetID, *mountTarget.SubnetId, "Subnet ID from AWS should match Terraform output - expected %s, got %s", subnetID, *mountTarget.SubnetId)
			assert.NotEmpty(t, *mountTarget.FileSystemId, "File system ID from AWS should not be empty")
			assert.NotEmpty(t, *mountTarget.IpAddress, "IP address from AWS should not be empty")
			assert.NotEmpty(t, *mountTarget.NetworkInterfaceId, "Network interface ID from AWS should not be empty")
			assert.NotEmpty(t, *mountTarget.AvailabilityZoneId, "Availability zone ID from AWS should not be empty")
			assert.NotEmpty(t, *mountTarget.AvailabilityZoneName, "Availability zone name from AWS should not be empty")

			// Verify lifecycle state in AWS
			assert.Equal(t, "available", string(mountTarget.LifeCycleState),
				"Mount target %s should be in 'available' state in AWS", mountTargetID)

			// Verify the network interface ID from AWS matches the Terraform output
			expectedNI := mountTargetNetworkInterfaceIDs[subnetID]
			assert.Equal(t, expectedNI, *mountTarget.NetworkInterfaceId,
				"Network interface ID from AWS should match Terraform output")

			t.Logf("✅ Mount target %s validated in AWS: IP=%s, State=%s, NI=%s",
				mountTargetID, *mountTarget.IpAddress, mountTarget.LifeCycleState, *mountTarget.NetworkInterfaceId)
		}
	})

	t.Run("TestNetworkInterfaceExistsInAWS", func(t *testing.T) {
		for subnetID, networkInterfaceID := range mountTargetNetworkInterfaceIDs {
			t.Logf("Validating network interface %s in subnet %s via AWS EC2 API", networkInterfaceID, subnetID)

			// Describe the network interface using AWS EC2 API to validate it exists
			input := &ec2.DescribeNetworkInterfacesInput{
				NetworkInterfaceIds: []string{networkInterfaceID},
			}

			result, err := ec2Client.DescribeNetworkInterfaces(context.TODO(), input)
			require.NoError(t, err, "Failed to describe network interface %s for subnet %s via AWS EC2 API", networkInterfaceID, subnetID)
			require.NotNil(t, result, "DescribeNetworkInterfaces result should not be nil")
			require.Len(t, result.NetworkInterfaces, 1, "Should return exactly one network interface from AWS API")

			netInterface := result.NetworkInterfaces[0]

			// Verify network interface properties from AWS match Terraform outputs
			assert.Equal(t, networkInterfaceID, *netInterface.NetworkInterfaceId, "Network interface ID from AWS should match Terraform output")
			assert.Equal(t, subnetID, *netInterface.SubnetId, "Subnet ID from AWS should match Terraform output - expected %s, got %s", subnetID, *netInterface.SubnetId)
			assert.NotEmpty(t, *netInterface.PrivateIpAddress, "Network interface from AWS should have a private IP address")
			assert.Equal(t, "in-use", string(netInterface.Status),
				"Network interface should be in 'in-use' state in AWS")

			// Verify security groups are attached in AWS
			require.NotEmpty(t, netInterface.Groups, "Network interface in AWS should have at least one security group attached")
			for _, group := range netInterface.Groups {
				assert.NotEmpty(t, *group.GroupId, "Security group ID from AWS should not be empty")
				assert.NotEmpty(t, *group.GroupName, "Security group name from AWS should not be empty")
			}

			t.Logf("✅ Network interface %s validated in AWS: IP=%s, Status=%s, SecurityGroups=%d",
				networkInterfaceID, *netInterface.PrivateIpAddress, netInterface.Status, len(netInterface.Groups))
		}
	})

	t.Run("TestMountTargetDNSResolution", func(t *testing.T) {
		for subnetID, dns := range mountTargetDNSNames {
			// Verify DNS name format follows AWS EFS naming convention
			// Format: fs-xxxxxxxx.efs.region.amazonaws.com
			assert.Regexp(t, `^fs-[a-f0-9]+\.efs\.[a-z0-9-]+\.amazonaws\.com$`, dns,
				"DNS name for subnet %s should follow AWS EFS format", subnetID)
		}
	})

	t.Run("TestMountTargetSecurityGroups", func(t *testing.T) {
		for subnetID, mountTargetID := range mountTargetIDs {
			t.Logf("Validating security groups for mount target %s in subnet %s", mountTargetID, subnetID)

			// Describe the mount target to get security groups
			input := &efs.DescribeMountTargetsInput{
				MountTargetId: aws.String(mountTargetID),
			}

			result, err := efsClient.DescribeMountTargets(context.TODO(), input)
			require.NoError(t, err, "Failed to describe mount target %s for subnet %s", mountTargetID, subnetID)
			require.NotNil(t, result, "DescribeMountTargets result should not be nil")
			require.Len(t, result.MountTargets, 1, "Should return exactly one mount target")

			// Get security groups from the network interface
			networkInterfaceID := mountTargetNetworkInterfaceIDs[subnetID]
			niInput := &ec2.DescribeNetworkInterfacesInput{
				NetworkInterfaceIds: []string{networkInterfaceID},
			}

			niResult, err := ec2Client.DescribeNetworkInterfaces(context.TODO(), niInput)
			require.NoError(t, err, "Failed to describe network interface for mount target %s", mountTargetID)
			require.NotNil(t, niResult, "DescribeNetworkInterfaces result should not be nil")
			require.Len(t, niResult.NetworkInterfaces, 1, "Should return exactly one network interface")

			securityGroups := niResult.NetworkInterfaces[0].Groups
			require.NotEmpty(t, securityGroups, "Mount target should have security groups attached")

			// Verify each security group exists and is configured
			for _, sg := range securityGroups {
				sgInput := &ec2.DescribeSecurityGroupsInput{
					GroupIds: []string{*sg.GroupId},
				}

				sgResult, err := ec2Client.DescribeSecurityGroups(context.TODO(), sgInput)
				require.NoError(t, err, "Failed to describe security group %s", *sg.GroupId)
				require.NotNil(t, sgResult, "DescribeSecurityGroups result should not be nil")
				require.Len(t, sgResult.SecurityGroups, 1, "Should return exactly one security group")

				securityGroup := sgResult.SecurityGroups[0]
				assert.Equal(t, *sg.GroupId, *securityGroup.GroupId, "Security group ID should match")
				assert.NotEmpty(t, *securityGroup.GroupName, "Security group name should not be empty")
			}
		}
	})

	// Add validation that we deployed the expected number of mount targets
	t.Run("TestExpectedNumberOfMountTargets", func(t *testing.T) {
		expectedCount := len(mountTargetIDs)
		assert.Greater(t, expectedCount, 0, "Should have at least one mount target deployed")
		assert.Equal(t, len(mountTargetDNSNames), expectedCount, "DNS names count should match mount targets count")
		assert.Equal(t, len(mountTargetNetworkInterfaceIDs), expectedCount, "Network interfaces count should match mount targets count")
		t.Logf("Validated %d mount targets across all subnets", expectedCount)
	})
}

// TestMultiSubnetWithChanges tests the multi-subnet example with dynamic subnet changes
// This validates that mount targets are not rebuilt when subnet configurations change
func TestMultiSubnetWithChanges(t *testing.T, ctx testTypes.TestContext) {
	t.Log("=======================================================",
		"=== Stage 1: Deploy all 3 mount targets ===",
		"=======================================================")

	opts := ctx.TerratestTerraformOptions()

	// Initial deployment with all subnets - get all outputs
	mountTargetIDs := terraform.OutputMap(t, opts, "mount_target_ids")
	mountTargetSubnetIDs := terraform.OutputMap(t, opts, "mount_target_subnet_ids")
	mountTargetDNSNames := terraform.OutputMap(t, opts, "mount_target_dns_names")
	mountTargetAZDNSNames := terraform.OutputMap(t, opts, "mount_target_az_dns_names")
	mountTargetNetworkInterfaceIDs := terraform.OutputMap(t, opts, "mount_target_network_interface_ids")
	mountTargetAZNames := terraform.OutputMap(t, opts, "mount_target_availability_zones")
	efsFileSystemID := terraform.Output(t, opts, "efs_file_system_id")
	efsFileSystemARN := terraform.Output(t, opts, "efs_file_system_arn")

	// Verify all 3 mount targets are created
	assert.Len(t, mountTargetIDs, 3, "Should have 3 mount targets initially")
	assert.Contains(t, mountTargetIDs, "az-a", "Should have mount target for az-a")
	assert.Contains(t, mountTargetIDs, "az-b", "Should have mount target for az-b")
	assert.Contains(t, mountTargetIDs, "az-c", "Should have mount target for az-c")

	// Verify all output maps have 3 entries
	assert.Len(t, mountTargetSubnetIDs, 3, "Should have 3 subnet IDs")
	assert.Len(t, mountTargetDNSNames, 3, "Should have 3 DNS names")
	assert.Len(t, mountTargetAZDNSNames, 3, "Should have 3 AZ-specific DNS names")
	assert.Len(t, mountTargetNetworkInterfaceIDs, 3, "Should have 3 network interface IDs")
	assert.Len(t, mountTargetAZNames, 3, "Should have 3 availability zone names")
	assert.NotEmpty(t, efsFileSystemID, "EFS file system ID should not be empty")
	assert.NotEmpty(t, efsFileSystemARN, "EFS file system ARN should not be empty")

	// Verify all values are populated for each mount target
	for _, key := range []string{"az-a", "az-b", "az-c"} {
		assert.NotEmpty(t, mountTargetIDs[key], fmt.Sprintf("Mount target ID for %s should not be empty", key))
		assert.NotEmpty(t, mountTargetSubnetIDs[key], fmt.Sprintf("Subnet ID for %s should not be empty", key))
		assert.NotEmpty(t, mountTargetDNSNames[key], fmt.Sprintf("DNS name for %s should not be empty", key))
		assert.NotEmpty(t, mountTargetAZDNSNames[key], fmt.Sprintf("AZ DNS name for %s should not be empty", key))
		assert.NotEmpty(t, mountTargetNetworkInterfaceIDs[key], fmt.Sprintf("Network interface ID for %s should not be empty", key))
		assert.NotEmpty(t, mountTargetAZNames[key], fmt.Sprintf("AZ name for %s should not be empty", key))
	} // Store initial IDs for comparison
	initialMountTargetAZ_A := mountTargetIDs["az-a"]
	initialMountTargetAZ_B := mountTargetIDs["az-b"]
	initialMountTargetAZ_C := mountTargetIDs["az-c"]
	initialSubnetAZ_A := mountTargetSubnetIDs["az-a"]
	initialSubnetAZ_C := mountTargetSubnetIDs["az-c"]

	t.Logf("Initial mount target IDs: az-a=%s, az-b=%s, az-c=%s",
		initialMountTargetAZ_A, initialMountTargetAZ_B, initialMountTargetAZ_C)

	// Validate all mount targets exist in AWS
	ValidateMountTargetsInAWS(t, mountTargetIDs, mountTargetSubnetIDs, opts)

	t.Log("=======================================================",
		"=== Stage 2: Remove middle subnet (az-b) - Should not rebuild others ===",
		"=======================================================")

	opts.SetVarsAfterVarFiles = true
	opts.Vars = map[string]interface{}{
		"enabled_subnet_indices": []int{0, 2}, // Remove subnet at index 1 (az-b)
	}

	// Apply with the new configuration
	terraform.Apply(t, opts)

	// Get new outputs after removal
	mountTargetIDs = terraform.OutputMap(t, opts, "mount_target_ids")
	mountTargetSubnetIDs = terraform.OutputMap(t, opts, "mount_target_subnet_ids")
	mountTargetDNSNames = terraform.OutputMap(t, opts, "mount_target_dns_names")
	mountTargetAZDNSNames = terraform.OutputMap(t, opts, "mount_target_az_dns_names")
	mountTargetNetworkInterfaceIDs = terraform.OutputMap(t, opts, "mount_target_network_interface_ids")
	mountTargetAZNames = terraform.OutputMap(t, opts, "mount_target_availability_zones")

	// Verify only 2 mount targets remain
	assert.Len(t, mountTargetIDs, 2, "Should have 2 mount targets after removal")
	assert.Contains(t, mountTargetIDs, "az-a", "Should still have mount target for az-a")
	assert.NotContains(t, mountTargetIDs, "az-b", "Should not have mount target for az-b")
	assert.Contains(t, mountTargetIDs, "az-c", "Should still have mount target for az-c")

	// Verify all output maps have 2 entries
	assert.Len(t, mountTargetSubnetIDs, 2, "Should have 2 subnet IDs")
	assert.Len(t, mountTargetDNSNames, 2, "Should have 2 DNS names")
	assert.Len(t, mountTargetAZDNSNames, 2, "Should have 2 AZ-specific DNS names")
	assert.Len(t, mountTargetNetworkInterfaceIDs, 2, "Should have 2 network interface IDs")
	assert.Len(t, mountTargetAZNames, 2, "Should have 2 availability zone names") // CRITICAL TEST: Verify IDs haven't changed (no rebuild)
	assert.Equal(t, initialMountTargetAZ_A, mountTargetIDs["az-a"],
		"Mount target az-a should NOT be rebuilt (ID should remain the same)")
	assert.Equal(t, initialMountTargetAZ_C, mountTargetIDs["az-c"],
		"Mount target az-c should NOT be rebuilt (ID should remain the same)")
	assert.Equal(t, initialSubnetAZ_A, mountTargetSubnetIDs["az-a"],
		"Subnet ID for az-a should remain the same")
	assert.Equal(t, initialSubnetAZ_C, mountTargetSubnetIDs["az-c"],
		"Subnet ID for az-c should remain the same")

	t.Log("✅ SUCCESS: Mount targets az-a and az-c were NOT rebuilt")

	// Validate remaining mount targets exist in AWS
	ValidateMountTargetsInAWS(t, mountTargetIDs, mountTargetSubnetIDs, opts)

	t.Log("=======================================================",
		"=== Stage 3: Remove first subnet (az-a) - Should not rebuild others ===",
		"=======================================================")

	opts.SetVarsAfterVarFiles = true
	opts.Vars = map[string]interface{}{
		"enabled_subnet_indices": []int{2}, // Remove subnet at index 0 (az-a)
	}

	// Apply with the new configuration
	terraform.Apply(t, opts)

	// Get new outputs after removal
	mountTargetIDs = terraform.OutputMap(t, opts, "mount_target_ids")
	mountTargetSubnetIDs = terraform.OutputMap(t, opts, "mount_target_subnet_ids")
	mountTargetDNSNames = terraform.OutputMap(t, opts, "mount_target_dns_names")
	mountTargetAZDNSNames = terraform.OutputMap(t, opts, "mount_target_az_dns_names")
	mountTargetNetworkInterfaceIDs = terraform.OutputMap(t, opts, "mount_target_network_interface_ids")
	mountTargetAZNames = terraform.OutputMap(t, opts, "mount_target_availability_zones")

	// Verify only 1 mount targets remain
	assert.Len(t, mountTargetIDs, 1, "Should have 1 mount target after removal")
	assert.NotContains(t, mountTargetIDs, "az-a", "Should not have mount target for az-a")
	assert.NotContains(t, mountTargetIDs, "az-b", "Should still not have mount target for az-b")
	assert.Contains(t, mountTargetIDs, "az-c", "Should still have mount target for az-c")

	// Verify all output maps have 1 entries
	assert.Len(t, mountTargetSubnetIDs, 1, "Should have 1 subnet IDs")
	assert.Len(t, mountTargetDNSNames, 1, "Should have 1 DNS names")
	assert.Len(t, mountTargetAZDNSNames, 1, "Should have 1 AZ-specific DNS names")
	assert.Len(t, mountTargetNetworkInterfaceIDs, 1, "Should have 1 network interface IDs")
	assert.Len(t, mountTargetAZNames, 1, "Should have 1 availability zone names") // CRITICAL TEST: Verify IDs haven't changed (no rebuild)
	assert.Equal(t, initialMountTargetAZ_C, mountTargetIDs["az-c"],
		"Mount target az-c should NOT be rebuilt (ID should remain the same)")
	assert.Equal(t, initialSubnetAZ_C, mountTargetSubnetIDs["az-c"],
		"Subnet ID for az-c should remain the same")

	t.Log("✅ SUCCESS: Mount target az-c was NOT rebuilt")

	// Validate remaining mount targets exist in AWS
	ValidateMountTargetsInAWS(t, mountTargetIDs, mountTargetSubnetIDs, opts)

	t.Log("=======================================================",
		"=== Stage 4: Add back other mount targets (az-a and az-b) ===",
		"=======================================================")

	opts.Vars = map[string]interface{}{
		"enabled_subnet_indices": []int{0, 1, 2}, // Add back subnet at index 0 (az-a) index 1 (az-b)
	}

	// Apply with the new configuration
	terraform.Apply(t, opts)

	// Get new outputs after adding back
	mountTargetIDs = terraform.OutputMap(t, opts, "mount_target_ids")
	mountTargetSubnetIDs = terraform.OutputMap(t, opts, "mount_target_subnet_ids")
	mountTargetDNSNames = terraform.OutputMap(t, opts, "mount_target_dns_names")
	mountTargetAZDNSNames = terraform.OutputMap(t, opts, "mount_target_az_dns_names")
	mountTargetNetworkInterfaceIDs = terraform.OutputMap(t, opts, "mount_target_network_interface_ids")
	mountTargetAZNames = terraform.OutputMap(t, opts, "mount_target_availability_zones")

	// Verify all 3 mount targets exist again
	assert.Len(t, mountTargetIDs, 3, "Should have 3 mount targets after adding back")
	assert.Contains(t, mountTargetIDs, "az-a", "Should have mount target for az-a")
	assert.Contains(t, mountTargetIDs, "az-b", "Should have mount target for az-b")
	assert.Contains(t, mountTargetIDs, "az-c", "Should have mount target for az-c")

	// Verify all output maps have 3 entries again
	assert.Len(t, mountTargetSubnetIDs, 3, "Should have 3 subnet IDs")
	assert.Len(t, mountTargetDNSNames, 3, "Should have 3 DNS names")
	assert.Len(t, mountTargetAZDNSNames, 3, "Should have 3 AZ-specific DNS names")
	assert.Len(t, mountTargetNetworkInterfaceIDs, 3, "Should have 3 network interface IDs")
	assert.Len(t, mountTargetAZNames, 3, "Should have 3 availability zone names")
	// Verify az-a has a NEW ID (it was destroyed and recreated)
	assert.NotEqual(t, initialMountTargetAZ_A, mountTargetIDs["az-a"],
		"Mount target az-a was rebuilt with a new id")
	// Verify az-b has a NEW ID (it was destroyed and recreated)
	assert.NotEqual(t, initialMountTargetAZ_B, mountTargetIDs["az-b"],
		"Mount target az-b should have a NEW ID (was destroyed and recreated)")
	// Verify az-c still hasen't been rebuilt
	assert.Equal(t, initialMountTargetAZ_C, mountTargetIDs["az-c"],
		"Mount target az-c should STILL not be rebuilt")

	t.Log("✅ SUCCESS: Mount targets az-a and az-b were rebuilt with new IDs, az-c remains stable")

	// Validate all mount targets exist in AWS
	ValidateMountTargetsInAWS(t, mountTargetIDs, mountTargetSubnetIDs, opts)

	t.Log("=======================================================",
		"=== All stages completed successfully ===",
		"=======================================================")
}

// ValidateMountTargetsInAWS validates that all mount targets exist and are available in AWS
// Mount targets should already be created and available after terraform apply completes
func ValidateMountTargetsInAWS(t *testing.T, mountTargetIDs, subnetIDs map[string]string, opts *terraform.Options) {
	// Get the region from Terraform configuration
	region := GetRegionFromTerraform(t, opts)
	awsConfig := GetAWSConfig(t, region)
	efsClient := efs.NewFromConfig(awsConfig)

	for key, mountTargetID := range mountTargetIDs {
		subnetID := subnetIDs[key]
		t.Logf("Validating mount target %s (key=%s, subnet=%s) in AWS", mountTargetID, key, subnetID)

		input := &efs.DescribeMountTargetsInput{
			MountTargetId: aws.String(mountTargetID),
		}

		result, err := efsClient.DescribeMountTargets(context.TODO(), input)
		require.NoError(t, err, "Failed to describe mount target %s", mountTargetID)
		require.Len(t, result.MountTargets, 1, "Should return exactly one mount target")

		mountTarget := result.MountTargets[0]

		// Verify properties
		assert.Equal(t, mountTargetID, aws.ToString(mountTarget.MountTargetId), "Mount target ID should match")
		assert.Equal(t, subnetID, aws.ToString(mountTarget.SubnetId), "Subnet ID should match")
		assert.Equal(t, "available", string(mountTarget.LifeCycleState),
			"Mount target %s should be in 'available' state", mountTargetID)

		t.Logf("✅ Mount target %s (key=%s) validated successfully in AWS", mountTargetID, key)
	}
}

// TestSimpleExample tests the simple example with a single mount target
func TestSimpleExample(t *testing.T, ctx testTypes.TestContext) {
	t.Log("=== Testing Simple Example (Single Mount Target) ===")

	opts := ctx.TerratestTerraformOptions()

	// Get all outputs
	mountTargetIDs := terraform.OutputMap(t, opts, "mount_target_ids")
	mountTargetSubnetIDs := terraform.OutputMap(t, opts, "mount_target_subnet_ids")
	mountTargetDNSNames := terraform.OutputMap(t, opts, "mount_target_dns_names")
	mountTargetAZDNSNames := terraform.OutputMap(t, opts, "mount_target_az_dns_names")
	mountTargetNetworkInterfaceIDs := terraform.OutputMap(t, opts, "mount_target_network_interface_ids")
	mountTargetAZNames := terraform.OutputMap(t, opts, "mount_target_availability_zone_names")
	mountTargetAZIDs := terraform.OutputMap(t, opts, "mount_target_availability_zone_ids")
	efsFileSystemID := terraform.Output(t, opts, "efs_file_system_id")
	efsFileSystemARN := terraform.Output(t, opts, "efs_file_system_arn")

	// Verify exactly one mount target
	assert.Len(t, mountTargetIDs, 1, "Simple example should have exactly 1 mount target")
	assert.Contains(t, mountTargetIDs, "primary", "Simple example should use 'primary' as the key")

	// Verify all outputs are populated for the 'primary' mount target
	assert.NotEmpty(t, mountTargetIDs["primary"], "Mount target ID should not be empty")
	assert.NotEmpty(t, mountTargetSubnetIDs["primary"], "Subnet ID should not be empty")
	assert.NotEmpty(t, mountTargetDNSNames["primary"], "DNS name should not be empty")
	assert.NotEmpty(t, mountTargetAZDNSNames["primary"], "AZ-specific DNS name should not be empty")
	assert.NotEmpty(t, mountTargetNetworkInterfaceIDs["primary"], "Network interface ID should not be empty")
	assert.NotEmpty(t, mountTargetAZNames["primary"], "Availability zone name should not be empty")
	assert.NotEmpty(t, mountTargetAZIDs["primary"], "Availability zone ID should not be empty")
	assert.NotEmpty(t, efsFileSystemID, "EFS file system ID should not be empty")
	assert.NotEmpty(t, efsFileSystemARN, "EFS file system ARN should not be empty")

	// Log output values
	t.Logf("Mount target ID: %s", mountTargetIDs["primary"])
	t.Logf("Subnet ID: %s", mountTargetSubnetIDs["primary"])
	t.Logf("DNS name: %s", mountTargetDNSNames["primary"])
	t.Logf("AZ-specific DNS name: %s", mountTargetAZDNSNames["primary"])
	t.Logf("Network interface ID: %s", mountTargetNetworkInterfaceIDs["primary"])
	t.Logf("Availability zone: %s (%s)", mountTargetAZNames["primary"], mountTargetAZIDs["primary"])
	t.Logf("EFS file system: %s (%s)", efsFileSystemID, efsFileSystemARN)

	// Validate mount target exists in AWS using AWS API
	region := GetRegionFromTerraform(t, opts)
	awsConfig := GetAWSConfig(t, region)
	efsClient := efs.NewFromConfig(awsConfig)
	ec2Client := ec2.NewFromConfig(awsConfig)

	mountTargetID := mountTargetIDs["primary"]
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
	assert.Equal(t, mountTargetSubnetIDs["primary"], *mountTarget.SubnetId, "Subnet ID from AWS should match Terraform output")
	assert.Equal(t, efsFileSystemID, *mountTarget.FileSystemId, "File system ID from AWS should match Terraform output")
	assert.Equal(t, "available", string(mountTarget.LifeCycleState), "Mount target should be available in AWS")
	assert.NotEmpty(t, *mountTarget.IpAddress, "Mount target should have an IP address")
	assert.NotEmpty(t, *mountTarget.AvailabilityZoneId, "Mount target should have an availability zone ID")
	assert.NotEmpty(t, *mountTarget.AvailabilityZoneName, "Mount target should have an availability zone name")
	t.Logf("✅ Mount target validated in AWS: ID=%s, IP=%s, State=%s",
		*mountTarget.MountTargetId, *mountTarget.IpAddress, mountTarget.LifeCycleState)

	// Validate network interface exists via EC2 API
	networkInterfaceID := mountTargetNetworkInterfaceIDs["primary"]
	t.Logf("Validating network interface %s exists in AWS via EC2 API", networkInterfaceID)

	niInput := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{networkInterfaceID},
	}

	niResult, err := ec2Client.DescribeNetworkInterfaces(context.TODO(), niInput)
	require.NoError(t, err, "Failed to describe network interface via AWS EC2 API")
	require.Len(t, niResult.NetworkInterfaces, 1, "Should return exactly one network interface from AWS API")

	netInterface := niResult.NetworkInterfaces[0]
	assert.Equal(t, networkInterfaceID, *netInterface.NetworkInterfaceId, "Network interface ID from AWS should match Terraform output")
	assert.Equal(t, mountTargetSubnetIDs["primary"], *netInterface.SubnetId, "Network interface subnet ID should match")
	assert.Equal(t, "in-use", string(netInterface.Status), "Network interface should be in-use")
	assert.NotEmpty(t, *netInterface.PrivateIpAddress, "Network interface should have a private IP")
	require.NotEmpty(t, netInterface.Groups, "Network interface should have security groups attached")
	t.Logf("✅ Network interface validated in AWS: ID=%s, IP=%s, Status=%s, SecurityGroups=%d",
		*netInterface.NetworkInterfaceId, *netInterface.PrivateIpAddress, netInterface.Status, len(netInterface.Groups))

	t.Log("✅ Simple example validated successfully in both Terraform outputs and AWS API")
} // GetMountTargetDetails retrieves detailed information about a mount target
func GetMountTargetDetails(t *testing.T, mountTargetID string, region string) map[string]interface{} {
	awsConfig := GetAWSConfig(t, region)
	efsClient := efs.NewFromConfig(awsConfig)

	input := &efs.DescribeMountTargetsInput{
		MountTargetId: aws.String(mountTargetID),
	}

	result, err := efsClient.DescribeMountTargets(context.TODO(), input)
	require.NoError(t, err, "Failed to describe mount target %s", mountTargetID)
	require.Len(t, result.MountTargets, 1, "Should return exactly one mount target")

	mt := result.MountTargets[0]

	details := map[string]interface{}{
		"MountTargetId":        aws.ToString(mt.MountTargetId),
		"FileSystemId":         aws.ToString(mt.FileSystemId),
		"SubnetId":             aws.ToString(mt.SubnetId),
		"LifeCycleState":       string(mt.LifeCycleState),
		"IpAddress":            aws.ToString(mt.IpAddress),
		"NetworkInterfaceId":   aws.ToString(mt.NetworkInterfaceId),
		"AvailabilityZoneId":   aws.ToString(mt.AvailabilityZoneId),
		"AvailabilityZoneName": aws.ToString(mt.AvailabilityZoneName),
		"OwnerId":              aws.ToString(mt.OwnerId),
	}

	return details
}

// CompareMountTargetIDs compares two maps of mount target IDs and logs differences
func CompareMountTargetIDs(t *testing.T, before, after map[string]string, stage string) {
	t.Logf("=== Comparing Mount Target IDs at %s ===", stage)

	for key, beforeID := range before {
		if afterID, exists := after[key]; exists {
			if beforeID == afterID {
				t.Logf("✅ %s: UNCHANGED (ID: %s)", key, beforeID)
			} else {
				t.Logf("🔄 %s: CHANGED (was: %s, now: %s)", key, beforeID, afterID)
			}
		} else {
			t.Logf("❌ %s: REMOVED (was: %s)", key, beforeID)
		}
	}

	for key, afterID := range after {
		if _, exists := before[key]; !exists {
			t.Logf("➕ %s: ADDED (ID: %s)", key, afterID)
		}
	}
}

// AssertMountTargetUnchanged asserts that specific mount targets haven't changed
func AssertMountTargetUnchanged(t *testing.T, beforeIDs, afterIDs map[string]string, keys []string) {
	for _, key := range keys {
		beforeID, beforeExists := beforeIDs[key]
		afterID, afterExists := afterIDs[key]

		assert.True(t, beforeExists, "Mount target %s should exist in before state", key)
		assert.True(t, afterExists, "Mount target %s should exist in after state", key)
		assert.Equal(t, beforeID, afterID,
			fmt.Sprintf("Mount target %s should NOT be rebuilt (ID should remain %s)", key, beforeID))
	}
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
func GetRegionFromTerraform(t *testing.T, opts *terraform.Options) string {
	// Try to get region from availability zone names (most reliable)
	// Try both possible output names for compatibility with different examples
	// terraform.OutputMap already extracts the map values for us
	var azNames map[string]string

	// First try mount_target_availability_zone_names (simple example)
	outputs := terraform.OutputAll(t, opts)
	if _, exists := outputs["mount_target_availability_zone_names"]; exists {
		azNames = terraform.OutputMap(t, opts, "mount_target_availability_zone_names")
	} else if _, exists := outputs["mount_target_availability_zones"]; exists {
		// Fallback to mount_target_availability_zones (multi-subnet example)
		azNames = terraform.OutputMap(t, opts, "mount_target_availability_zones")
	}

	for _, az := range azNames {
		if len(az) > 0 {
			// Remove the last character (a, b, c) to get the region
			// For example: us-west-2a -> us-west-2
			region := az[:len(az)-1]
			t.Logf("Detected region %s from availability zone %s", region, az)
			return region
		}
	}

	// Fallback: parse from DNS name (format: fs-xxxxx.efs.REGION.amazonaws.com)
	dnsNames := terraform.OutputMap(t, opts, "mount_target_dns_names")
	for _, dns := range dnsNames {
		if len(dns) > 0 {
			// Parse DNS: fs-xxxxx.efs.REGION.amazonaws.com
			var region string
			efsIdx := -1
			for i := 0; i < len(dns)-4; i++ {
				if dns[i:i+5] == ".efs." {
					efsIdx = i + 5
					break
				}
			}
			if efsIdx > 0 {
				for i := efsIdx; i < len(dns); i++ {
					if dns[i] == '.' {
						region = dns[efsIdx:i]
						break
					}
				}
				if region != "" {
					t.Logf("Detected region %s from DNS name %s", region, dns)
					return region
				}
			}
		}
	}

	// If all else fails, return us-east-2 as default (but log a warning)
	t.Logf("WARNING: Could not detect region from Terraform, using default us-east-2")
	return "us-east-2"
}
