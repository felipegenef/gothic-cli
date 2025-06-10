/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	gothic_cli "github.com/felipegenef/gothicframework/pkg/cli"
	"github.com/felipegenef/gothicframework/pkg/helpers"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy or remove the application on AWS using AWS SAM.",
	Long: `This command builds and deploys (or removes) the application using AWS SAM.

During deployment, it performs the following steps:
  - Converts template files into Go source files
  - Builds an optimized Docker image tailored for AWS Lambda environments
  - Publishes the image to AWS ECR and uses it as the Lambda runtime
  - Optimizes images found in the 'optimize' folder and uploads them to an S3 bucket
  - Sets up an AWS CloudFront distribution to serve as a gateway for both S3 and Lambda origins
  - Cleans up any existing CloudFront distribution if redeploying

This process ensures your application is efficiently built and deployed to AWS.`,
	RunE: newDeployCommand(gothic_cli.NewCli()),
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("stage", "s", "dev", "Define AWS stage to deploy or delete")
	deployCmd.Flags().StringP("action", "a", "deploy", "Action to be taken, to deploy or delete the api")
}

type DeployCommand struct {
	cli            *gothic_cli.GothicCli
	allowedActions []string
}

func newDeployCommandCli(cli *gothic_cli.GothicCli) DeployCommand {
	return DeployCommand{
		cli:            cli,
		allowedActions: []string{"delete", "deploy"},
	}
}

func newDeployCommand(cli gothic_cli.GothicCli) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		command := newDeployCommandCli(&cli)
		stageFlag, err := cmd.Flags().GetString("stage")
		if err != nil {
			return err
		}
		action, err := cmd.Flags().GetString("action")
		if err != nil {
			return err
		}
		if !command.isValidAction(action) {
			return fmt.Errorf("error: invalid action \"%s\". Allowed values: %v", action, command.allowedActions)

		}

		return command.Deploy(stageFlag, action)
	}
}

func (command *DeployCommand) Deploy(stage string, action string) error {
	if err := command.setup(stage); err != nil {
		return err
	}
	if err := command.cli.Templ.Render(); err != nil {
		return err
	}

	if err := command.cli.FileBasedRouter.Render(command.cli.GetConfig().GoModName); err != nil {
		return err
	}

	if err := command.cli.Tailwind.Build(); err != nil {
		return err
	}

	if err := command.cli.AwsSam.Build(); err != nil {
		return err
	}

	config := command.cli.GetConfig()
	appID, err := command.cli.GetAppId()
	if err != nil {
		return fmt.Errorf("error getting app id: %v", err)
	}

	originBucketName := config.ProjectName + "-" + stage + "-" + appID

	switch action {
	case "deploy":
		if err := command.cli.AwsSam.Deploy(stage, config.ProjectName, config.Deploy.Profile); err != nil {
			return err
		}

		defer command.cli.AWS.AddCloudFrontAssets(originBucketName, config.Deploy.Region, config.Deploy.Profile)
		defer command.cli.AWS.CleanCloudFrontCache(config.ProjectName, stage, config.Deploy.Region, config.Deploy.Profile)
	case "delete":
		if err := command.cli.AWS.RemoveCloudFrontAssets(originBucketName, config.Deploy.Region, config.Deploy.Profile); err != nil {
			return err
		}
		if err := command.cli.AwsSam.DeleteStack(stage, config.ProjectName, config.Deploy.Profile); err != nil {
			return err
		}
	}
	command.cleanup()
	return nil
}

func (command *DeployCommand) setup(stage string) error {
	config := command.cli.GetConfig()

	// Check if the Deploy configuration is present
	if config.Deploy == nil {
		return fmt.Errorf("Deploy configuration missing in gothic-config.json")
	}
	fmt.Println("SELECTED STAGE: " + stage)
	// Select the environment based on the --stage parameter
	var envConfig gothic_cli.EnvVariables = config.Deploy.Stages[stage]
	appID, err := command.cli.GetAppId()

	if err != nil {
		return fmt.Errorf("error getting appId: %v", err)
	}

	// Check if the minimum variables are set
	if envConfig.BucketName == "" || envConfig.LambdaName == "" {
		envConfig.LambdaName = config.ProjectName + "-" + stage + "-" + appID
		envConfig.BucketName = config.ProjectName + "-" + stage + "-" + appID
	}

	var yamlInfo helpers.SamYamlTemplateInfo
	yamlInfo.Timeout = config.Deploy.ServerTimeout
	yamlInfo.MemorySize = config.Deploy.ServerMemory
	yamlInfo.ProjectName = config.ProjectName
	yamlInfo.StageTemplateInfo.Name = stage
	yamlInfo.StageTemplateInfo.BucketName = `BucketName: "` + envConfig.BucketName + `"`
	yamlInfo.StageTemplateInfo.LambdaName = `LambdaName: "` + envConfig.LambdaName + `"`
	yamlInfo.StageTemplateInfo.CertificateArn = ""
	yamlInfo.StageTemplateInfo.HostedZone = ""
	yamlInfo.StageTemplateInfo.CustomDomain = ""
	var env []helpers.EnvValueInfo

	for key, val := range envConfig.ENV {
		var formattedValue string

		// Check the type of the value
		if strValue, ok := val.(string); ok {
			// If the value is a string, add quotes
			formattedValue = fmt.Sprintf("%q", strValue) // %q adds quotes around the string
		} else {
			// For other types, use default formatting
			formattedValue = fmt.Sprintf("%v", val)
		}
		env = append(env, helpers.EnvValueInfo{
			Key:   key,
			Value: formattedValue, // Convert interface{} to string
		})
	}
	yamlInfo.StageTemplateInfo.Env = env

	// Check if a custom domain is needed
	if config.Deploy.CustomDomain {
		if config.Deploy.Region != "us-east-1" && envConfig.CertificateArn == nil {
			return fmt.Errorf("for custom domains, if you set a region other than us-east-1, you must provide a us-east-1 ACM CertificateArn in your environment variables")
		}

		if envConfig.CustomDomain == nil || envConfig.HostedZoneId == nil {
			return fmt.Errorf("environment variables customDomain and hostedZoneId are required when deploy.customDomain is set to true")
		}

		yamlInfo.StageTemplateInfo.CustomDomain = `customDomain: "` + *envConfig.CustomDomain + `"`
		yamlInfo.StageTemplateInfo.HostedZone = `hostedZoneId: "` + *envConfig.HostedZoneId + `"`

		if envConfig.CertificateArn != nil {
			yamlInfo.UsedTemplateName = ".gothicCli/templates/template-custom-domain-with-arn.yaml"
			yamlInfo.StageTemplateInfo.CertificateArn = `certificateArn: "` + *envConfig.CertificateArn + `"`
		} else {
			yamlInfo.UsedTemplateName = ".gothicCli/templates/template-custom-domain.yaml"
		}

	} else {
		yamlInfo.UsedTemplateName = ".gothicCli/templates/template-default.yaml"
	}
	command.cli.Templates.CopyFile(yamlInfo.UsedTemplateName, "template.yaml")
	command.cli.Templates.UpdateFromTemplate("template.yaml", "template.yaml", yamlInfo)
	command.cli.Templates.CopyFile(".gothicCli/templates/Dockerfile-template", "Dockerfile")
	command.cli.Templates.CopyFile(".gothicCli/templates/samconfig-template.toml", "samconfig.toml")
	// Replace the region
	command.cli.Templates.UpdateFromTemplate("samconfig.toml", "samconfig.toml", helpers.SamTomlTemplateInfo{
		StackName: config.ProjectName,
		AwsRegion: config.Deploy.Region,
	})
	return nil
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

func (command *DeployCommand) isValidAction(c string) bool {
	for _, a := range command.allowedActions {
		if a == c {
			return true
		}
	}
	return false
}
