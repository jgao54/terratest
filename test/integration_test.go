// Integration tests that test cross-package functionality in AWS.
package test

import (
	"testing"

	"github.com/gruntwork-io/terraform-test"
	"github.com/gruntwork-io/terraform-test/aws"
	"github.com/gruntwork-io/terraform-test/log"
	"github.com/gruntwork-io/terraform-test/terraform"
	"github.com/gruntwork-io/terraform-test/util"
)

func TestUploadKeyPair(t *testing.T) {
	// Assign randomly generated values
	region := aws.GetRandomRegion()
	id := util.UniqueId()

	// Create the keypair
	keyPair, err := util.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Errorf("Failed to generate keypair: %s\n", err.Error())
	}

	// Create key in EC2
	t.Logf("Creating EC2 Keypair %s in %s...", id, region)
	defer aws.DeleteEC2KeyPair(region, id)
	aws.CreateEC2KeyPair(region, id, keyPair.PublicKey)
}

func TestTerraformApplyMainFunction(t *testing.T) {
	rand, err := main.CreateRandomResourceCollection()
	defer main.DestroyRandomResourceCollection(rand)
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	main.TerraformApply("Integration Test - TestTerraformApplyMainFunction", "resources/minimal-example", vars, false)
}

func TestTerraformApplyAndDestroyOnMinimalExample(t *testing.T) {
	logger := log.NewLogger("TestTerraformApplyAndDestroy")

	// CONSTANTS
	terraformTemplatePath := "resources/minimal-example"

	// SETUP
	region := aws.GetRandomRegion()
	id := util.UniqueId()
	logger.Printf("Random values selected. Region = %s, Id = %s\n", region, id)

	// Create and upload the keypair
	keyPair, err := util.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Errorf("Failed to generate keypair: %s\n", err.Error())
	}
	logger.Println("Generated keypair. Printing out private key...")
	logger.Printf("%s", keyPair.PrivateKey)

	logger.Println("Creating EC2 Keypair...")
	defer aws.DeleteEC2KeyPair(region, id)
	aws.CreateEC2KeyPair(region, id, keyPair.PublicKey)

	// Set Terraform to use Remote State
	err = aws.AssertS3BucketExists(TF_REMOTE_STATE_S3_BUCKET_REGION, TF_REMOTE_STATE_S3_BUCKET_NAME)
	if err != nil {
		t.Fatal("Terraform Remote State S3 Bucket does not exist.")
	}

	terraform.ConfigureRemoteState(terraformTemplatePath, TF_REMOTE_STATE_S3_BUCKET_NAME, id + "/main.tf", TF_REMOTE_STATE_S3_BUCKET_REGION, logger)

	// TEST
	// Apply the Terraform template
	vars := make(map[string]string)
	vars["aws_region"] = region
	vars["ec2_key_name"] = id
	vars["ec2_instance_name"] = id
	vars["ec2_image"] = aws.GetUbuntuAmi(region)

	logger.Println("Running terraform apply...")
	defer terraform.Destroy("resources/minimal-example", vars, logger)
	err = terraform.Apply(terraformTemplatePath, vars, logger)
	if err != nil {
		t.Fatalf("Failed to terraform apply: %s", err.Error())
	}
}

func TestTerraformApplyWithRetryOnMinimalExample(t *testing.T) {
	logger := log.NewLogger("TestTerraformApplyAndDestroy")

	// CONSTANTS
	terraformTemplatePath := "resources/minimal-example"

	// SETUP
	region := aws.GetRandomRegion()
	id := util.UniqueId()
	logger.Printf("Random values selected. Region = %s, Id = %s\n", region, id)

	// Create and upload the keypair
	keyPair, err := util.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Errorf("Failed to generate keypair: %s\n", err.Error())
	}
	logger.Println("Generated keypair. Printing out private key...")
	logger.Printf("%s", keyPair.PrivateKey)

	logger.Println("Creating EC2 Keypair...")
	defer aws.DeleteEC2KeyPair(region, id)
	aws.CreateEC2KeyPair(region, id, keyPair.PublicKey)

	// Set Terraform to use Remote State
	err = aws.AssertS3BucketExists(TF_REMOTE_STATE_S3_BUCKET_REGION, TF_REMOTE_STATE_S3_BUCKET_NAME)
	if err != nil {
		t.Fatal("Terraform Remote State S3 Bucket does not exist.")
	}

	terraform.ConfigureRemoteState(terraformTemplatePath, TF_REMOTE_STATE_S3_BUCKET_NAME, id + "/main.tf", TF_REMOTE_STATE_S3_BUCKET_REGION, logger)

	// TEST
	// Apply the Terraform template
	vars := make(map[string]string)
	vars["aws_region"] = region
	vars["ec2_key_name"] = id
	vars["ec2_instance_name"] = id
	vars["ec2_image"] = aws.GetUbuntuAmi(region)

	logger.Println("Running terraform apply...")
	defer terraform.Destroy("resources/minimal-example", vars, logger)
	_, err = terraform.ApplyAndGetOutputWithRetry(terraformTemplatePath, vars, logger)
	if err != nil {
		t.Fatalf("Failed to terraform apply: %s", err.Error())
	}
}