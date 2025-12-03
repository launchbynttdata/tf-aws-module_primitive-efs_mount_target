package testimpl

import (
	"context"
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
    // Get outputs from Terraform
    mountTargetIDs := terraform.OutputMap(t, ctx.TerratestTerraformOptions(), "aws_efs_mount_target_id")
    mountTargetDNSNames := terraform.OutputMap(t, ctx.TerratestTerraformOptions(), "aws_efs_mount_target_dns_name")
    mountTargetNetworkInterfaceIDs := terraform.OutputMap(t, ctx.TerratestTerraformOptions(), "aws_efs_mount_target_network_interface_id")

    // Initialize AWS clients
    awsConfig := GetAWSConfig(t)
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
            t.Logf("Validating mount target %s in subnet %s", mountTargetID, subnetID)

            // Describe the mount target using AWS API
            input := &efs.DescribeMountTargetsInput{
                MountTargetId: aws.String(mountTargetID),
            }

            result, err := efsClient.DescribeMountTargets(context.TODO(), input)
            require.NoError(t, err, "Failed to describe mount target %s for subnet %s", mountTargetID, subnetID)
            require.NotNil(t, result, "DescribeMountTargets result should not be nil")
            require.Len(t, result.MountTargets, 1, "Should return exactly one mount target")

            mountTarget := result.MountTargets[0]

            // Verify mount target properties
            assert.Equal(t, mountTargetID, *mountTarget.MountTargetId, "Mount target ID should match")
            assert.Equal(t, subnetID, *mountTarget.SubnetId, "Subnet ID should match - expected %s, got %s", subnetID, *mountTarget.SubnetId)
            assert.NotEmpty(t, *mountTarget.FileSystemId, "File system ID should not be empty")
            assert.NotEmpty(t, *mountTarget.IpAddress, "IP address should not be empty")
            assert.NotEmpty(t, *mountTarget.NetworkInterfaceId, "Network interface ID should not be empty")
            assert.NotEmpty(t, *mountTarget.AvailabilityZoneId, "Availability zone ID should not be empty")
            assert.NotEmpty(t, *mountTarget.AvailabilityZoneName, "Availability zone name should not be empty")

            // Verify lifecycle state
            assert.Equal(t, "available", string(mountTarget.LifeCycleState),
                "Mount target %s should be in 'available' state", mountTargetID)
        }
    })

    t.Run("TestNetworkInterfaceExistsInAWS", func(t *testing.T) {
        for subnetID, networkInterfaceID := range mountTargetNetworkInterfaceIDs {
            t.Logf("Validating network interface %s in subnet %s", networkInterfaceID, subnetID)

            // Describe the network interface using EC2 API
            input := &ec2.DescribeNetworkInterfacesInput{
                NetworkInterfaceIds: []string{networkInterfaceID},
            }

            result, err := ec2Client.DescribeNetworkInterfaces(context.TODO(), input)
            require.NoError(t, err, "Failed to describe network interface %s for subnet %s", networkInterfaceID, subnetID)
            require.NotNil(t, result, "DescribeNetworkInterfaces result should not be nil")
            require.Len(t, result.NetworkInterfaces, 1, "Should return exactly one network interface")

            netInterface := result.NetworkInterfaces[0]

            // Verify network interface properties
            assert.Equal(t, networkInterfaceID, *netInterface.NetworkInterfaceId, "Network interface ID should match")
            assert.Equal(t, subnetID, *netInterface.SubnetId, "Subnet ID should match for network interface - expected %s, got %s", subnetID, *netInterface.SubnetId)
            assert.NotEmpty(t, *netInterface.PrivateIpAddress, "Network interface should have a private IP address")
            assert.Equal(t, "in-use", string(netInterface.Status),
                "Network interface should be in 'in-use' state")

            // Verify security groups are attached
            require.NotEmpty(t, netInterface.Groups, "Network interface should have at least one security group attached")
            for _, group := range netInterface.Groups {
                assert.NotEmpty(t, *group.GroupId, "Security group ID should not be empty")
                assert.NotEmpty(t, *group.GroupName, "Security group name should not be empty")
            }
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

func GetAWSSTSClient(t *testing.T) *sts.Client {
    awsSTSClient := sts.NewFromConfig(GetAWSConfig(t))
    return awsSTSClient
}

func GetAWSConfig(t *testing.T) (cfg aws.Config) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	require.NoErrorf(t, err, "unable to load SDK config, %v", err)
	return cfg
}
