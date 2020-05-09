package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"

	res "github.com/cloudetc/awsweeper/resource"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_ECSCluster_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/ecs-cluster"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")

	assertEcsClusterExists(t, env, id)

	writeConfigID(t, terraformDir, res.EcsCluster, id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n", "-debug")
	require.NoError(t, err)

	assertEcsClusterDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_ECSCluster_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/ecs-cluster"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")

	assertEcsClusterExists(t, env, id)

	writeConfigTag(t, terraformDir, res.EcsCluster)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertEcsClusterDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertEcsClusterExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, ecsClusterExists(t, env, id))
}

func assertEcsClusterDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, ecsClusterExists(t, env, id))
}

func ecsClusterExists(t *testing.T, env EnvVars, id string) bool {
	opts := &ecs.DescribeClustersInput{
		Clusters: []*string{&id},
	}

	resp, err := env.AWSClient.DescribeClusters(opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Clusters) == 0 {
		return false
	}

	if *resp.Clusters[0].Status == "INACTIVE" {
		return false
	}

	return true
}
