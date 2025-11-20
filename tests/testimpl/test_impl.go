package testimpl

import (
	"context"
	"regexp"
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
	// Get AWS STS client to verify account info
	stsClient := GetAWSSTSClient(t)

	// Get the actual caller identity from AWS
	callerIdentity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	require.NoError(t, err, "Failed to get caller identity from AWS")

	// Get outputs from Terraform
	accountId := terraform.Output(t, ctx.TerratestTerraformOptions(), "account_id")
	arn := terraform.Output(t, ctx.TerratestTerraformOptions(), "arn")
	helloMessage := terraform.Output(t, ctx.TerratestTerraformOptions(), "hello_message")

	t.Run("TestAccountIdMatches", func(t *testing.T) {
		testAccountIdMatches(t, callerIdentity, accountId)
	})

	t.Run("TestArnMatches", func(t *testing.T) {
		testArnMatches(t, callerIdentity, arn)
	})

	t.Run("TestHelloMessage", func(t *testing.T) {
		testHelloMessage(t, helloMessage)
	})
}

func testAccountIdMatches(t *testing.T, callerIdentity *sts.GetCallerIdentityOutput, accountId string) {
	assert.Equal(t, *callerIdentity.Account, accountId, "Account ID from Terraform should match AWS caller identity")
	assert.NotEmpty(t, accountId, "Account ID should not be empty")

	// Verify it's a valid 12-digit account ID
	matched, _ := regexp.MatchString(`^\d{12}$`, accountId)
	assert.True(t, matched, "Account ID should be a 12-digit number")
}

func testArnMatches(t *testing.T, callerIdentity *sts.GetCallerIdentityOutput, arn string) {
	assert.Equal(t, *callerIdentity.Arn, arn, "ARN from Terraform should match AWS caller identity")
	assert.NotEmpty(t, arn, "ARN should not be empty")

	// Verify it's a valid ARN format
	matched, _ := regexp.MatchString(`^arn:aws:`, arn)
	assert.True(t, matched, "ARN should start with 'arn:aws:'")
}

func testHelloMessage(t *testing.T, helloMessage string) {
	assert.NotEmpty(t, helloMessage, "Hello message should not be empty")
	assert.Contains(t, helloMessage, "Hello", "Message should contain 'Hello'")
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
