/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	gothic_cli "github.com/felipegenef/gothic-cli/pkg/cli"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: newDeployComand(gothic_cli.NewCli()),
}

func newDeployComand(cli gothic_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		comand := newDeployCommandCli(&cli)
		stageFlag, err := cmd.Flags().GetString("stage")
		if err != nil {
			return err
		}
		action, err := cmd.Flags().GetString("action")
		if err != nil {
			return err
		}
		if !isValidAction(action) {
			return fmt.Errorf("error: invalid action \"%s\". Allowed values: %v", action, allowedActions)

		}
		comand.Deploy(stageFlag, action)
		return nil
	}
}

var allowedActions = []string{"delete", "deploy"}

func isValidAction(c string) bool {
	for _, a := range allowedActions {
		if a == c {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("stage", "s", "dev", "Define AWS stage to deploy or delete")
	deployCmd.Flags().StringP("action", "a", "deploy", "Action to be taken, to deploy or delete the api")
}

type DeployCommand struct {
	cli *gothic_cli.GothicCli
}

func newDeployCommandCli(cli *gothic_cli.GothicCli) DeployCommand {
	return DeployCommand{
		cli: cli,
	}
}

func (command *DeployCommand) cdnAddOrRemoveAssets(stage string, action string) {

	config := command.cli.GetConfig()

	// Read the app ID from the file
	content, err := os.ReadFile(".gothicCli/app-id.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Convert the content to string
	appID := string(content)

	// Construct the S3 bucket name
	bucketPublicFolderName := "s3://" + config.ProjectName + "-" + stage + "-" + appID + "/public"

	// Check the action and execute the corresponding command
	switch action {
	case "add":
		addFilesCmd := exec.Command("aws", "s3", "cp", "public", bucketPublicFolderName, "--recursive", "--region", config.Deploy.Region, "--profile", config.Deploy.Profile)
		addFilesCmd.Stdout = os.Stdout
		addFilesCmd.Stdin = os.Stdin
		addFilesCmd.Stderr = os.Stderr

		// Run the command
		if err := addFilesCmd.Run(); err != nil {
			log.Fatalf("Error executing add command: %v", err)
		}
		fmt.Println("S3 Files added successfully.")

	case "delete":
		removeFilesCmd := exec.Command("aws", "s3", "rm", bucketPublicFolderName, "--recursive", "--region", config.Deploy.Region, "--profile", config.Deploy.Profile)
		removeFilesCmd.Stdout = os.Stdout
		removeFilesCmd.Stdin = os.Stdin
		removeFilesCmd.Stderr = os.Stderr

		// Run the command
		if err := removeFilesCmd.Run(); err != nil {
			log.Fatalf("Error executing delete command: %v", err)
		}
		fmt.Println("S3 Files deleted successfully.")

	default:
		log.Fatalf("Invalid action specified: %s. Use 'add' or 'delete'.", action)
	}
}

func (command *DeployCommand) setup(stage string, action string) {
	config := command.cli.GetConfig()

	// Check if the Deploy configuration is present
	if config.Deploy == nil {
		log.Fatalf("Deploy configuration missing in gothic-config.json")
	}
	fmt.Println("SELECTED STAGE: " + stage)
	// Select the environment based on the --stage parameter
	var envConfig gothic_cli.EnvVariables = config.Deploy.Stages[stage]
	// TODO: move this ID to the config file
	content, err := os.ReadFile(".gothicCli/app-id.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Convert the content to string
	appID := string(content)

	// Check if the minimum variables are set
	if envConfig.BucketName == "" || envConfig.LambdaName == "" {
		envConfig.LambdaName = config.ProjectName + "-" + stage + "-" + appID
		envConfig.BucketName = config.ProjectName + "-" + stage + "-" + appID
	}

	// Replace the project name in all files
	filePaths := []string{
		".gothicCli/templates/samconfig-template.toml",
		".gothicCli/templates/template-custom-domain-with-arn.yaml",
		".gothicCli/templates/template-custom-domain.yaml",
		".gothicCli/templates/template-default.yaml",
	}
	// TODO use native template replace on cli.Template struct methods
	for _, filePath := range filePaths {
		if err := command.replaceInFile("gothic-example", config.ProjectName, filePath); err != nil {
			log.Fatalf("Error replacing project name in file %s: %w", filePath, err)
		}
	}

	// Check if a custom domain is needed
	if config.Deploy.CustomDomain {
		if config.Deploy.Region != "us-east-1" && envConfig.CertificateArn == nil {
			log.Fatalf("For custom domains, if you set a region other than us-east-1, you must provide a us-east-1 ACM CertificateArn in your environment variables")
		}

		if envConfig.CustomDomain != nil || envConfig.HostedZoneId != nil {
			templateFile := ".gothicCli/templates/template-custom-domain-with-arn.yaml"
			if envConfig.CertificateArn != nil {
				if err := command.replaceInFile("AcmArnReplacerString", *envConfig.CertificateArn, templateFile); err != nil {
					log.Fatalf("Error replacing certificate ARN in template file: %w", err)
				}

				// TODO use native template replace on cli.Template struct methods
				command.copyFile(templateFile, "template.yaml")
				command.replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, stage, "template.yaml")
				command.replaceCustomDomainWithArnValues(envConfig.CustomDomain, envConfig.HostedZoneId, envConfig.CertificateArn, "template.yaml")
				command.replaceEnvVariables(envConfig.ENV, "template.yaml")
				command.replaceTimeoutAndMemory(config.Deploy.ServerTimeout, config.Deploy.ServerMemory, "template.yaml")

			} else {
				// TODO use native template replace on cli.Template struct methods
				templateFile := ".gothicCli/templates/template-custom-domain.yaml"

				command.copyFile(templateFile, "template.yaml")
				command.replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, stage, "template.yaml")
				command.replaceCustomDomainValues(envConfig.CustomDomain, envConfig.HostedZoneId, "template.yaml")
				command.replaceEnvVariables(envConfig.ENV, "template.yaml")
				command.replaceTimeoutAndMemory(config.Deploy.ServerTimeout, config.Deploy.ServerMemory, "template.yaml")
			}
		} else {
			log.Fatalf("Environment variables customDomain and hostedZoneId are required when deploy.customDomain is set to true")
		}
	} else {
		// TODO use native template replace on cli.Template struct methods
		templateFile := ".gothicCli/templates/template-default.yaml"

		command.copyFile(templateFile, "template.yaml")
		// Replace the environment variables
		command.replaceEnvVariables(envConfig.ENV, "template.yaml")
		command.replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, stage, "template.yaml")
		command.replaceTimeoutAndMemory(config.Deploy.ServerTimeout, config.Deploy.ServerMemory, "template.yaml")

	}
	// TODO use native template replace on cli.Template struct methods
	command.copyFile(".gothicCli/templates/Dockerfile-template", "Dockerfile")
	command.copyFile(".gothicCli/templates/samconfig-template.toml", "samconfig.toml")
	// Replace the region
	if err := command.replaceInFile("regionReplacerString", config.Deploy.Region, "samconfig.toml"); err != nil {
		log.Fatalf("Error replacing region in file %s: %w", "samconfig.toml", err)
	}

}

func (command *DeployCommand) cleanupCache(config gothic_cli.Config, stage string) {
	// Execute the command to get the CloudFront distribution ID
	getDistributionIdCMD := exec.Command("aws", "cloudformation", "describe-stacks", "--stack-name", config.ProjectName+"-"+stage, "--query", "Stacks[0].Outputs[?OutputKey=='CloudFrontId'].OutputValue", "--output", "text", "--region", config.Deploy.Region, "--profile", config.Deploy.Profile)

	// Capture the output of the command
	var out bytes.Buffer
	getDistributionIdCMD.Stdout = &out
	if err := getDistributionIdCMD.Run(); err != nil {
		log.Fatalf("Error getting CloudFront Id: %v", err)
	}

	// The result of the command will be the CloudFront distribution ID
	distributionId := strings.TrimSpace(out.String()) // Remove any extra spaces
	if distributionId == "" {
		log.Fatal("CloudFront ID not found")
	}

	// Now, use the distribution ID in the command to create the invalidation
	cleanCachesCmd := exec.Command("aws", "cloudfront", "create-invalidation", "--distribution-id", distributionId, "--paths", "/*", "--region", config.Deploy.Region, "--profile", config.Deploy.Profile)

	// Execute the cache cleanup command
	if err := cleanCachesCmd.Run(); err != nil {
		log.Fatalf("Error cleaning up deploy files: %v", err)
	}

	// Print the distribution ID and confirm the cache cleanup
	fmt.Printf("Successfully reset CloudFront cache for distribution: %s\n", distributionId)
}

func (command *DeployCommand) Deploy(stage string, action string) {
	command.setup(stage, action)
	// TODO deal with error
	command.cli.Templ.Render()
	// TODO deal with error
	command.cli.Tailwind.Build()
	// TODO deal with error
	command.cli.AwsSam.Build()
	// Create a variable to store the configuration
	config := command.cli.GetConfig()

	content, err := os.ReadFile(".gothicCli/app-id.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Convert the content to string
	appID := string(content)

	originBucketName := config.ProjectName + "-" + stage + "-" + appID

	switch action {
	case "deploy":
		// TODO deal with error
		command.cli.AwsSam.Deploy(stage, config.ProjectName, config.Deploy.Profile)

		defer command.cli.AWS.AddCloudFrontAssets(originBucketName, config.Deploy.Region, config.Deploy.Profile)
		defer command.cli.AWS.CleanCloudFrontCache(config.ProjectName, stage, config.Deploy.Region, config.Deploy.Profile)
	case "delete":
		// TODO deal with error
		command.cli.AWS.RemoveCloudFrontAssets(originBucketName, config.Deploy.Region, config.Deploy.Profile)
		// TODO deal with error
		command.cli.AwsSam.DeleteStack(stage, config.ProjectName, config.Deploy.Profile)
	}
	command.cleanup()
}

func (command *DeployCommand) replaceCustomDomainWithArnValues(customDomain *string, hostedZone *string, arn *string, templateFile string) {
	// TODO use native template replace on cli.Template struct methods
	// Call the function that replaces custom domain values
	command.replaceCustomDomainValues(customDomain, hostedZone, templateFile)

	// Check if arn is not nil before dereferencing it
	var arnValue string
	if arn != nil {
		arnValue = *arn
	} else {
		arnValue = "" // or a default value
	}
	// TODO use native template replace on cli.Template struct methods
	// Replace the ARN value in the template file
	if err := command.replaceInFile("certificateArnReplacerString", `certificateArn: "`+arnValue+`"`, templateFile); err != nil {
		log.Fatalf("Error adding ARN value to SAM template file: %v", err)
	}
}

func (command *DeployCommand) replaceTimeoutAndMemory(timeoutValue int, memoryValue int, templateFile string) {
	// TODO use native template replace on cli.Template struct methods
	if err := command.replaceInFile("timeoutReplacerString", strconv.Itoa(timeoutValue), templateFile); err != nil {
		log.Fatalf("Error adding timeout value to SAM template file")
	}
	// TODO use native template replace on cli.Template struct methods
	if err := command.replaceInFile("memoryReplacerString", strconv.Itoa(memoryValue), templateFile); err != nil {
		log.Fatalf("Error adding memory value to SAM template file")
	}
}

func (command *DeployCommand) copyFile(filePath string, destinationPath string) error {
	fileContent, err := os.ReadFile(filePath)

	if err != nil {
		return err
	}

	return os.WriteFile(destinationPath, fileContent, 0644)
}

func (command *DeployCommand) replaceInFile(originalString string, replaceString string, filePath string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	replacedFile := []byte(strings.ReplaceAll(string(fileContent), originalString, replaceString))
	return os.WriteFile(filePath, replacedFile, 0644)
}

func (command *DeployCommand) replaceCustomDomainValues(customDomain *string, hostedZone *string, templateFile string) {
	// Check if customDomain is not nil before dereferencing it
	var customDomainValue string
	if customDomain != nil {
		customDomainValue = *customDomain
	} else {
		customDomainValue = "" // or a default value
	}

	// Check if hostedZone is not nil before dereferencing it
	var hostedZoneValue string
	if hostedZone != nil {
		hostedZoneValue = *hostedZone
	} else {
		hostedZoneValue = "" // or a default value
	}
	// TODO use native template replace on cli.Template struct methods
	if err := command.replaceInFile("customDomainReplacerString", `customDomain: "`+customDomainValue+`"`, templateFile); err != nil {
		log.Fatalf("Error adding custom domain value to SAM template file: %v", err)
	}
	// TODO use native template replace on cli.Template struct methods
	if err := command.replaceInFile("hostedZoneReplacerString", `hostedZoneId: "`+hostedZoneValue+`"`, templateFile); err != nil {
		log.Fatalf("Error adding hosted zone value to SAM template file: %v", err)
	}
}

func (command *DeployCommand) replaceEnvVariables(env map[string]interface{}, templateFile string) {
	finalStageMapReplacer := ""
	finalEnvReplacer := ""
	// TODO use native template replace on cli.Template struct methods
	for key, value := range env {
		var formattedValue string

		// Check the type of the value
		if strValue, ok := value.(string); ok {
			// If the value is a string, add quotes
			formattedValue = fmt.Sprintf("%q", strValue) // %q adds quotes around the string
		} else {
			// For other types, use default formatting
			formattedValue = fmt.Sprintf("%v", value)
		}

		finalStageMapReplacer += "      " + key + ": " + formattedValue + "\n"
		finalEnvReplacer += "          " + key + ": !FindInMap [StagesMap, !Ref Stage, " + key + "]\n"
	}

	// Replace in the file with the map content
	if err := command.replaceInFile("stageMapStringReplacer", finalStageMapReplacer, templateFile); err != nil {
		log.Fatalf("Error adding stage map value to SAM template file: %v", err)
	}

	if err := command.replaceInFile("EnvStringReplacer", finalEnvReplacer, templateFile); err != nil {
		log.Fatalf("Error adding env value to SAM template file: %v", err)
	}
}

func (command *DeployCommand) replaceStageBucketAndLambdaName(lambdaName string, bucketName string, stage string, templateFile string) {
	// TODO use native template replace on cli.Template struct methods
	if err := command.replaceInFile("lambdaNameReplacerString", `LambdaName: "`+lambdaName+`"`, templateFile); err != nil {
		log.Fatalf("Error adding lambda value to SAM template file")
	}

	if err := command.replaceInFile("bucketNameReplacerString", `BucketName: "`+bucketName+`"`, templateFile); err != nil {
		log.Fatalf("Error adding bucket value to SAM template file")
	}

	if err := command.replaceInFile("stageReplacerString", stage, templateFile); err != nil {
		log.Fatalf("Error adding stage value to SAM template file")
	}
}

func (command *DeployCommand) cleanup() {
	// Define the paths of the files you want to delete
	filesToDelete := []string{
		"Dockerfile",
		"template.yaml",
		"samconfig.toml",
	}

	// Iterate over each file and attempt to delete it
	for _, filePath := range filesToDelete {
		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("Error cleaning up deploy file file %s: %v\n", filePath, err)
			continue
		}
	}
}
