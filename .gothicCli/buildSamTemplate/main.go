package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	gothicCliShared "github.com/felipegenef/gothic-cli/.gothicCli"
)

func main() {
	// Define the --stage flag to specify the environment (dev, staging, prod)
	stage := flag.String("stage", "default", "Specify the deployment stage (default, dev, staging, prod)")
	flag.Parse()

	var stageValue string
	if stage != nil {
		stageValue = *stage
	} else {
		stageValue = "" // or a default value
	}

	// Open the configuration file
	file, err := os.Open("gothic-config.json")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Create a variable to store the configuration
	var config gothicCliShared.Config

	// Decode the JSON from the file
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	// Check if the Deploy configuration is present
	if config.Deploy == nil {
		log.Fatalf("Deploy configuration missing in gothic-config.json")
	}
	fmt.Println("SELECTED STAGE: " + stageValue)
	// Select the environment based on the --stage parameter
	var envConfig gothicCliShared.EnvVariables = config.Deploy.Stages[stageValue]

	content, err := os.ReadFile(".gothicCli/app-id.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Convert the content to string
	appID := string(content)

	// Check if the minimum variables are set
	if envConfig.BucketName == "" || envConfig.LambdaName == "" {
		envConfig.LambdaName = config.ProjectName + "-" + stageValue + "-" + appID
		envConfig.BucketName = config.ProjectName + "-" + stageValue + "-" + appID
	}

	// Replace the project name in all files
	filePaths := []string{
		".gothicCli/buildSamTemplate/templates/samconfig-template.toml",
		".gothicCli/buildSamTemplate/templates/template-custom-domain-with-arn.yaml",
		".gothicCli/buildSamTemplate/templates/template-custom-domain.yaml",
		".gothicCli/buildSamTemplate/templates/template-default.yaml",
	}

	for _, filePath := range filePaths {
		if err := replaceInFile("gothic-example", config.ProjectName, filePath); err != nil {
			log.Fatalf("Error replacing project name in file %s: %w", filePath, err)
		}
	}

	// Replace the region
	if err := replaceInFile("regionReplacerString", config.Deploy.Region, ".gothicCli/buildSamTemplate/templates/samconfig-template.toml"); err != nil {
		log.Fatalf("Error replacing region in file %s: %w", ".gothicCli/buildSamTemplate/templates/samconfig-template.toml", err)
	}

	// Check if a custom domain is needed
	if config.Deploy.CustomDomain {
		if config.Deploy.Region != "us-east-1" && envConfig.CertificateArn == nil {
			log.Fatalf("For custom domains, if you set a region other than us-east-1, you must provide a us-east-1 ACM CertificateArn in your environment variables")
		}

		if envConfig.CustomDomain != nil || envConfig.HostedZoneId != nil {
			templateFile := ".gothicCli/buildSamTemplate/templates/template-custom-domain-with-arn.yaml"
			if envConfig.CertificateArn != nil {
				if err := replaceInFile("AcmArnReplacerString", *envConfig.CertificateArn, templateFile); err != nil {
					log.Fatalf("Error replacing certificate ARN in template file: %w", err)
				}
				copyFile(templateFile, "template.yaml")
				replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, stageValue, "template.yaml")
				replaceCustomDomainWithArnValues(envConfig.CustomDomain, envConfig.HostedZoneId, envConfig.CertificateArn, "template.yaml")
				replaceEnvVariables(envConfig.ENV, "template.yaml")
				replaceTimeoutAndMemory(config.Deploy.ServerTimeout, config.Deploy.ServerMemory, "template.yaml")

			} else {
				templateFile := ".gothicCli/buildSamTemplate/templates/template-custom-domain.yaml"
				copyFile(templateFile, "template.yaml")
				replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, stageValue, "template.yaml")
				replaceCustomDomainValues(envConfig.CustomDomain, envConfig.HostedZoneId, "template.yaml")
				replaceEnvVariables(envConfig.ENV, "template.yaml")
				replaceTimeoutAndMemory(config.Deploy.ServerTimeout, config.Deploy.ServerMemory, "template.yaml")
			}
		} else {
			log.Fatalf("Environment variables customDomain and hostedZoneId are required when deploy.customDomain is set to true")
		}
	} else {
		templateFile := ".gothicCli/buildSamTemplate/templates/template-default.yaml"
		copyFile(templateFile, "template.yaml")
		// Replace the environment variables
		replaceEnvVariables(envConfig.ENV, "template.yaml")
		replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, stageValue, "template.yaml")
		replaceTimeoutAndMemory(config.Deploy.ServerTimeout, config.Deploy.ServerMemory, "template.yaml")

	}

	copyFile(".gothicCli/buildSamTemplate/templates/Dockerfile-template", "Dockerfile")
	copyFile(".gothicCli/buildSamTemplate/templates/samconfig-template.toml", "samconfig.toml")
	// Replace the region
	if err := replaceInFile("regionReplacerString", config.Deploy.Region, "samconfig.toml"); err != nil {
		log.Fatalf("Error replacing region in file %s: %w", "samconfig.toml", err)
	}
}

func replaceStageBucketAndLambdaName(lambdaName string, bucketName string, stage string, templateFile string) {
	if err := replaceInFile("lambdaNameReplacerString", `LambdaName: "`+lambdaName+`"`, templateFile); err != nil {
		log.Fatalf("Error adding lambda value to SAM template file")
	}

	if err := replaceInFile("bucketNameReplacerString", `BucketName: "`+bucketName+`"`, templateFile); err != nil {
		log.Fatalf("Error adding bucket value to SAM template file")
	}

	if err := replaceInFile("stageReplacerString", stage, templateFile); err != nil {
		log.Fatalf("Error adding stage value to SAM template file")
	}
}

func replaceEnvVariables(env map[string]interface{}, templateFile string) {
	finalStageMapReplacer := ""
	finalEnvReplacer := ""

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
	if err := replaceInFile("stageMapStringReplacer", finalStageMapReplacer, templateFile); err != nil {
		log.Fatalf("Error adding stage map value to SAM template file: %v", err)
	}

	if err := replaceInFile("EnvStringReplacer", finalEnvReplacer, templateFile); err != nil {
		log.Fatalf("Error adding env value to SAM template file: %v", err)
	}
}

func replaceCustomDomainValues(customDomain *string, hostedZone *string, templateFile string) {
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

	if err := replaceInFile("customDomainReplacerString", `customDomain: "`+customDomainValue+`"`, templateFile); err != nil {
		log.Fatalf("Error adding custom domain value to SAM template file: %v", err)
	}

	if err := replaceInFile("hostedZoneReplacerString", `hostedZoneId: "`+hostedZoneValue+`"`, templateFile); err != nil {
		log.Fatalf("Error adding hosted zone value to SAM template file: %v", err)
	}
}

func replaceCustomDomainWithArnValues(customDomain *string, hostedZone *string, arn *string, templateFile string) {
	// Call the function that replaces custom domain values
	replaceCustomDomainValues(customDomain, hostedZone, templateFile)

	// Check if arn is not nil before dereferencing it
	var arnValue string
	if arn != nil {
		arnValue = *arn
	} else {
		arnValue = "" // or a default value
	}

	// Replace the ARN value in the template file
	if err := replaceInFile("certificateArnReplacerString", `certificateArn: "`+arnValue+`"`, templateFile); err != nil {
		log.Fatalf("Error adding ARN value to SAM template file: %v", err)
	}
}

func replaceTimeoutAndMemory(timeoutValue int, memoryValue int, templateFile string) {
	if err := replaceInFile("timeoutReplacerString", strconv.Itoa(timeoutValue), templateFile); err != nil {
		log.Fatalf("Error adding timeout value to SAM template file")
	}

	if err := replaceInFile("memoryReplacerString", strconv.Itoa(memoryValue), templateFile); err != nil {
		log.Fatalf("Error adding memory value to SAM template file")
	}
}

func copyFile(filePath string, destinationPath string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return os.WriteFile(destinationPath, fileContent, 0644)
}

func replaceInFile(originalString string, replaceString string, filePath string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	replacedFile := []byte(strings.ReplaceAll(string(fileContent), originalString, replaceString))
	return os.WriteFile(filePath, replacedFile, 0644)
}
