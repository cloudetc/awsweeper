package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_CloudWatchLogGroup_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/cloudwatch-log-group"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertCloudWatchLogGroupExists(t, env, id)

	writeConfigID(t, terraformDir, res.CloudWatchLogGroup, id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertCloudWatchLogGroupDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_CloudWatchLogGroup_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}
	t.Skip("Tags not supported yet.")

	env := InitEnv(t)

	terraformDir := "./test-fixtures/cloudwatch-log-group"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertCloudWatchLogGroupExists(t, env, id)

	writeConfigTag(t, terraformDir, res.CloudWatchLogGroup)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertCloudWatchLogGroupDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertCloudWatchLogGroupExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, cloudWatchLogGroupExists(t, env, id))
}

func assertCloudWatchLogGroupDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, cloudWatchLogGroupExists(t, env, id))
}

func cloudWatchLogGroupExists(t *testing.T, env EnvVars, id string) bool {
	opts := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: &id,
	}

	resp, err := env.AWSClient.CloudWatchLogsAPI.DescribeLogGroups(opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.LogGroups) == 0 {
		return false
	}

	return true
}
