package testimpl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/gruntwork-io/terratest/modules/terraform"
	testTypes "github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposableComplete(t *testing.T, ctx testTypes.TestContext) {
	terraformOptions := ctx.TerratestTerraformOptions()

	addonArn := strings.TrimSpace(terraform.Output(t, terraformOptions, "addon_arn"))
	addonID := strings.TrimSpace(terraform.Output(t, terraformOptions, "addon_id"))
	createdAt := strings.TrimSpace(terraform.Output(t, terraformOptions, "addon_created_at"))
	modifiedAt := strings.TrimSpace(terraform.Output(t, terraformOptions, "addon_modified_at"))
	addonVersion := normalizeTerraformStringOutput(terraform.Output(t, terraformOptions, "addon_version"))
	tagsAll := terraform.OutputMap(t, terraformOptions, "addon_tags_all")

	clusterName, addonName := parseAddonID(t, addonID)
	region := extractRegionFromArn(t, addonArn)

	eksClient := GetAWSEKSClient(t, region)
	addonDetails := describeAddon(t, eksClient, clusterName, addonName)

	t.Run("TestAddonArnMatches", func(t *testing.T) {
		testAddonArnMatches(t, addonArn, clusterName, addonName, region, addonDetails)
	})

	t.Run("TestAddonIdMatches", func(t *testing.T) {
		testAddonIdMatches(t, addonID, clusterName, addonName, addonDetails)
	})

	t.Run("TestAddonTimestamps", func(t *testing.T) {
		testAddonTimestamps(t, createdAt, modifiedAt, addonDetails)
	})

	t.Run("TestAddonTags", func(t *testing.T) {
		testAddonTags(t, tagsAll, addonDetails)
	})

	t.Run("TestAddonVersion", func(t *testing.T) {
		testAddonVersion(t, addonVersion, addonDetails)
	})
}

func GetAWSEKSClient(t *testing.T, region string) *eks.Client {
	loadOptions := []func(*config.LoadOptions) error{}
	if region != "" {
		loadOptions = append(loadOptions, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), loadOptions...)
	require.NoError(t, err, "Failed to load AWS config for EKS client")

	return eks.NewFromConfig(cfg)
}

func describeAddon(t *testing.T, client *eks.Client, clusterName, addonName string) *types.Addon {
	output, err := client.DescribeAddon(context.TODO(), &eks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	})
	require.NoError(t, err, "Failed to describe EKS addon via AWS API")
	require.NotNil(t, output, "DescribeAddon response should not be nil")
	require.NotNil(t, output.Addon, "DescribeAddon response should include addon details")

	return output.Addon
}

func testAddonArnMatches(t *testing.T, terraformArn, clusterName, addonName, region string, addonDetails *types.Addon) {
	require.NotEmpty(t, terraformArn, "Terraform output ARN should not be empty")
	require.NotNil(t, addonDetails.AddonArn, "AWS addon ARN should not be nil")

	awsArn := aws.ToString(addonDetails.AddonArn)
	assert.Equal(t, awsArn, terraformArn, "Terraform output ARN should match AWS addon ARN")
	assert.Truef(t, strings.HasPrefix(terraformArn, "arn:"), "Expected ARN to start with 'arn:' but got %s", terraformArn)
	arnParts := strings.SplitN(terraformArn, ":", 6)
	require.Len(t, arnParts, 6, "Addon ARN should follow ARN format with six segments")
	assert.Equal(t, "eks", arnParts[2], "ARN service segment should be eks")
	assert.Equal(t, region, arnParts[3], "ARN region segment should match target region")
	assert.NotEmpty(t, arnParts[4], "ARN account segment should not be empty")
	resource := arnParts[5]
	assert.Truef(t, strings.HasPrefix(resource, "addon/"+clusterName+"/"+addonName), "Resource segment should start with addon/%s/%s but was %s", clusterName, addonName, resource)
}

func testAddonIdMatches(t *testing.T, terraformID, clusterName, addonName string, addonDetails *types.Addon) {
	require.NotEmpty(t, terraformID, "Terraform output ID should not be empty")
	require.NotNil(t, addonDetails.ClusterName, "AWS addon cluster name should not be nil")
	require.NotNil(t, addonDetails.AddonName, "AWS addon name should not be nil")

	expectedID := clusterName + ":" + addonName
	awsClusterName := aws.ToString(addonDetails.ClusterName)
	awsAddonName := aws.ToString(addonDetails.AddonName)
	awsID := awsClusterName + ":" + awsAddonName

	assert.Equal(t, expectedID, terraformID, "Terraform output ID should follow <cluster>:<addon> format")
	assert.Equal(t, clusterName, awsClusterName, "Cluster name should match between Terraform output and AWS API")
	assert.Equal(t, addonName, awsAddonName, "Addon name should match between Terraform output and AWS API")
	assert.Equal(t, expectedID, awsID, "Combined identifier should align with AWS values")
}

func testAddonTimestamps(t *testing.T, terraformCreatedAt, terraformModifiedAt string, addonDetails *types.Addon) {
	require.NotNil(t, addonDetails.CreatedAt, "AWS addon created timestamp should not be nil")
	require.NotNil(t, addonDetails.ModifiedAt, "AWS addon modified timestamp should not be nil")

	createdFromTerraform := parseTimestamp(t, terraformCreatedAt)
	modifiedFromTerraform := parseTimestamp(t, terraformModifiedAt)
	createdFromAWS := aws.ToTime(addonDetails.CreatedAt).UTC()
	modifiedFromAWS := aws.ToTime(addonDetails.ModifiedAt).UTC()

	assert.WithinDuration(t, createdFromAWS, createdFromTerraform, time.Second, "CreatedAt should match within one second")
	assert.WithinDuration(t, modifiedFromAWS, modifiedFromTerraform, time.Second, "ModifiedAt should match within one second")
	assert.False(t, modifiedFromTerraform.Before(createdFromTerraform), "Terraform ModifiedAt should not be before CreatedAt")
	assert.False(t, modifiedFromAWS.Before(createdFromAWS), "AWS ModifiedAt should not be before CreatedAt")
}

func testAddonTags(t *testing.T, terraformTags map[string]string, addonDetails *types.Addon) {
	require.NotNil(t, terraformTags, "Terraform tags output should not be nil")
	assert.NotEmpty(t, terraformTags, "Terraform tags should not be empty")

	awsTags := flattenTagMap(addonDetails.Tags)
	for key, expected := range terraformTags {
		assert.Equalf(t, expected, awsTags[key], "Tag %s should match between Terraform output and AWS API", key)
	}

	if managedBy, exists := terraformTags["ManagedBy"]; assert.True(t, exists, "ManagedBy tag should be present") {
		assert.Equal(t, "Terraform", managedBy, "ManagedBy tag should default to Terraform")
	}
}

func testAddonVersion(t *testing.T, terraformVersion string, addonDetails *types.Addon) {
	awsVersion := strings.TrimSpace(aws.ToString(addonDetails.AddonVersion))

	if terraformVersion == "" {
		assert.NotEmpty(t, awsVersion, "AWS should report the resolved addon version")
		return
	}

	assert.Equal(t, terraformVersion, awsVersion, "Addon version should match between Terraform output and AWS API")
}

func parseAddonID(t *testing.T, addonID string) (string, string) {
	require.NotEmpty(t, addonID, "Addon ID should not be empty")

	parts := strings.SplitN(addonID, ":", 2)
	require.Len(t, parts, 2, "Addon ID should use the format <cluster-name>:<addon-name>")
	require.NotEmpty(t, parts[0], "Cluster name within addon ID should not be empty")
	require.NotEmpty(t, parts[1], "Addon name within addon ID should not be empty")

	return parts[0], parts[1]
}

func extractRegionFromArn(t *testing.T, arn string) string {
	require.NotEmpty(t, arn, "Addon ARN should not be empty")

	parts := strings.Split(arn, ":")
	require.GreaterOrEqual(t, len(parts), 4, "Addon ARN should contain a region segment")

	region := parts[3]
	require.NotEmpty(t, region, "Region derived from ARN should not be empty")

	return region
}

func parseTimestamp(t *testing.T, value string) time.Time {
	trimmed := strings.TrimSpace(value)
	require.NotEmpty(t, trimmed, "Timestamp output should not be empty")

	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339Nano, trimmed)
	}
	require.NoErrorf(t, err, "Failed to parse timestamp value %q", value)

	return parsed.UTC()
}

func flattenTagMap(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return map[string]string{}
	}

	converted := make(map[string]string, len(tags))
	for key, value := range tags {
		converted[key] = value
	}

	return converted
}

func normalizeTerraformStringOutput(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "null" {
		return ""
	}
	return trimmed
}
