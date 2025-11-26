package testimpl

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

	       t.Run("TestMountTargetIDs", func(t *testing.T) {
		       for subnet, id := range mountTargetIDs {
			       assert.NotEmpty(t, id, "Mount target ID for subnet %s should not be empty", subnet)
		       }
	       })

	       t.Run("TestMountTargetDNSNames", func(t *testing.T) {
		       for subnet, dns := range mountTargetDNSNames {
			       assert.NotEmpty(t, dns, "Mount target DNS name for subnet %s should not be empty", subnet)
			       assert.Contains(t, dns, ".efs.", "DNS name for subnet %s should contain '.efs.'", subnet)
		       }
	       })

	       t.Run("TestMountTargetNetworkInterfaceIDs", func(t *testing.T) {
		       for subnet, ni := range mountTargetNetworkInterfaceIDs {
			       assert.NotEmpty(t, ni, "Network interface ID for subnet %s should not be empty", subnet)
		       }
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
