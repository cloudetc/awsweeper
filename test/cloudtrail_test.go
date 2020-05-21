package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudtrail"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	res "github.com/cloudetc/awsweeper/resource"
)

func TestAcc_CloudTrail_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/cloudtrail"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertCloudTrailExists(t, env, id)

	writeConfigID(t, terraformDir, res.CloudTrail, id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertCloudTrailDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_CloudTrail_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/cloudtrail"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertCloudTrailExists(t, env, id)

	writeConfigTag(t, terraformDir, res.CloudTrail)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertCloudTrailDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertCloudTrailExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, cloudTrailExists(t, env, id))
}

func assertCloudTrailDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, cloudTrailExists(t, env, id))
}

func cloudTrailExists(t *testing.T, env EnvVars, id string) bool {
	opts := &cloudtrail.DescribeTrailsInput{
		TrailNameList: []*string{&id},
	}
	resp, err := env.AWSClient.DescribeTrails(opts)
	if err != nil {
		t.Fatal()
	}

	if len(resp.TrailList) == 0 {
		return false
	}

	return true
}
