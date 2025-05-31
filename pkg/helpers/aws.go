package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type AwsHelper struct {
}

func NewAwsHelper() AwsHelper {
	return AwsHelper{}
}

func (a *AwsHelper) AddCloudFrontAssets(originBucketName string, region string, awsProfile string) error {

	// Construct the S3 bucket name
	bucketPublicFolderName := "s3://" + originBucketName + "/public"

	// Check the action and execute the corresponding command

	addFilesCmd := exec.Command("aws", "s3", "cp", "public", bucketPublicFolderName, "--recursive", "--region", region, "--profile", awsProfile)
	addFilesCmd.Stdout = os.Stdout
	addFilesCmd.Stdin = os.Stdin
	addFilesCmd.Stderr = os.Stderr

	// Run the command
	err := addFilesCmd.Run()
	if err != nil {
		fmt.Printf("Error adding CloudFront assets: %v", err)
		return err
	}
	fmt.Println("S3 Files added successfully.")
	return nil

}

func (a *AwsHelper) RemoveCloudFrontAssets(originBucketName string, region string, awsProfile string) error {

	// Construct the S3 bucket name
	bucketPublicFolderName := "s3://" + originBucketName + "/public"

	removeFilesCmd := exec.Command("aws", "s3", "rm", bucketPublicFolderName, "--recursive", "--region", region, "--profile", awsProfile)
	removeFilesCmd.Stdout = os.Stdout
	removeFilesCmd.Stdin = os.Stdin
	removeFilesCmd.Stderr = os.Stderr

	// Run the command
	err := removeFilesCmd.Run()
	if err != nil {
		fmt.Printf("Error removing CloudFront Assets: %v", err)
		return err
	}
	fmt.Println("S3 Files deleted successfully.")

	return nil
}

func (a *AwsHelper) CleanCloudFrontCache(stackName string, stage string, region string, awsProfile string) error {

	// Execute the command to get the CloudFront distribution ID
	getDistributionIdCMD := exec.Command("aws", "cloudformation", "describe-stacks", "--stack-name", stackName+"-"+stage, "--query", "Stacks[0].Outputs[?OutputKey=='CloudFrontId'].OutputValue", "--output", "text", "--region", region, "--profile", awsProfile)

	// Capture the output of the command
	var out bytes.Buffer
	getDistributionIdCMD.Stdout = &out
	err := getDistributionIdCMD.Run()
	if err != nil {
		fmt.Printf("Error getting CloudFront Id: %v", err)
		return err
	}

	// The result of the command will be the CloudFront distribution ID
	distributionId := strings.TrimSpace(out.String()) // Remove any extra spaces
	if distributionId == "" {
		fmt.Printf("CloudFront ID not found")
		return fmt.Errorf("CloudFront ID not found")
	}

	// Now, use the distribution ID in the command to create the invalidation
	cleanCachesCmd := exec.Command("aws", "cloudfront", "create-invalidation", "--distribution-id", distributionId, "--paths", "/*", "--region", region, "--profile", awsProfile)

	// Execute the cache cleanup command
	cleanCacheErr := cleanCachesCmd.Run()

	if cleanCacheErr != nil {
		fmt.Printf("Error cleaning up deploy files: %v", cleanCacheErr)
		return cleanCacheErr
	}

	// Print the distribution ID and confirm the cache cleanup
	fmt.Printf("Successfully reset CloudFront cache for distribution: %s\n", distributionId)
	return nil
}
