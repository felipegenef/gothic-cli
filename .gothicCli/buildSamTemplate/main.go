package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	gothicCliShared "github.com/felipegenef/gothic-cli/.gothicCli"
)

func main() {
	// Define o flag --stage para especificar o ambiente (dev, staging, prod)
	stage := flag.String("stage", "default", "Specify the deployment stage (default, dev, staging, prod)")
	flag.Parse()

	// Abre o arquivo de configuração
	file, err := os.Open("gothic-config.json")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Cria uma variável para armazenar a configuração
	var config gothicCliShared.Config

	// Decodifica o JSON do arquivo
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	// Verifica se a configuração de Deploy está presente
	if config.Deploy == nil {
		log.Fatalf("Deploy configuration missing on gothic-config.json")
	}

	// Seleciona o ambiente com base no parâmetro --stage
	var envConfig *gothicCliShared.EnvVariables

	switch *stage {
	case "default":
		envConfig = &config.Deploy.Stages.Default
	case "dev":
		envConfig = &config.Deploy.Stages.Dev
	case "staging":
		envConfig = &config.Deploy.Stages.Staging
	case "prod":
		envConfig = &config.Deploy.Stages.Prod

	default:
		log.Fatalf("Invalid stage: %s. Must be one of: dev, staging, prod", *stage)
	}

	// Verifica se as variáveis mínimas estão definidas
	if envConfig.BucketName == "" || envConfig.LambdaName == "" {
		log.Fatalf("Both BucketName and LambdaName must be set for stage: %s", *stage)
	}

	// Substitui o nome do projeto em todos os arquivos
	filePaths := []string{
		".gothicCli/buildSamTemplate/samconfig.toml",
		".gothicCli/buildSamTemplate/templates/template-custom-domain-with-arn.yaml",
		".gothicCli/buildSamTemplate/templates/template-custom-domain.yaml",
		".gothicCli/buildSamTemplate/templates/template-default.yaml",
	}

	for _, filePath := range filePaths {
		if err := replaceOnFile("gothic-example", config.ProjectName, filePath); err != nil {
			log.Fatalf("error replacing project name to file %s: %w", filePath, err)
		}
	}

	// Substitui a região
	if err := replaceOnFile("us-east-1", config.Deploy.Region, ".gothicCli/buildSamTemplate/samconfig.toml"); err != nil {
		log.Fatalf("error replacing region in file %s: %w", ".gothicCli/buildSamTemplate/samconfig.toml", err)
	}

	// Verifica se um domínio customizado é necessário
	if config.Deploy.CustomDomain {
		if config.Deploy.Region != "us-east-1" && envConfig.CertificateArn == nil {
			log.Fatalf("For custom domains, if you set a different region than us-east-1, you should provide a us-east-1 ACM CertificateArn on your Environment variables")
		}

		if envConfig.CustomDomain != nil || envConfig.HostedZoneId != nil {
			templateFile := ".gothicCli/buildSamTemplate/templates/template-custom-domain-with-arn.yaml"
			if envConfig.CertificateArn != nil {
				if err := replaceOnFile("AcmArnReplacerString", *envConfig.CertificateArn, templateFile); err != nil {
					log.Fatalf("error replacing certificate ARN in template file: %w", err)
				}
				copyFile(templateFile, "template.yaml")
				replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, *stage, "template.yaml")
				replaceCustomDomainWithArnValues(envConfig.CustomDomain, envConfig.HostedZoneId, envConfig.CertificateArn, "template.yaml")
				replaceEnvVariables(envConfig.ENV, "template.yaml")
				replaceTimeoutAndMemory(config.Deploy.ServerTimeout, config.Deploy.ServerMemory, "template.yaml")

			} else {
				templateFile := ".gothicCli/buildSamTemplate/templates/template-custom-domain.yaml"
				copyFile(templateFile, "template.yaml")
				replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, *stage, "template.yaml")
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
		// Substitui as variáveis de ambiente
		replaceEnvVariables(envConfig.ENV, "template.yaml")
		replaceStageBucketAndLambdaName(envConfig.LambdaName, envConfig.BucketName, *stage, "template.yaml")
		replaceTimeoutAndMemory(config.Deploy.ServerTimeout, config.Deploy.ServerMemory, "template.yaml")

	}

	copyFile(".gothicCli/buildSamTemplate/samconfig.toml", "samconfig.toml")
}

func replaceStageBucketAndLambdaName(lambdaName string, bucketName string, stage string, templateFile string) {
	if err := replaceOnFile("lambdaNameReplacerString", `LambdaName: "`+lambdaName+`"`, templateFile); err != nil {
		log.Fatalf("error adding lambda value to sam template file")
	}

	if err := replaceOnFile("bucketNameReplacerString", `BucketName: "`+bucketName+`"`, templateFile); err != nil {
		log.Fatalf("error adding bucket value to sam template file")
	}

	if err := replaceOnFile("stageReplacerString", stage, templateFile); err != nil {
		log.Fatalf("error adding stage value to sam template file")
	}
}

func replaceEnvVariables(env map[string]interface{}, templateFile string) {
	finalStageMapReplacer := ""
	finalEnvReplacer := ""

	for key, value := range env {
		finalStageMapReplacer += "      " + key + ": " + fmt.Sprintf("%v", value) + "\n"
		finalEnvReplacer += "          " + key + ": !FindInMap [StagesMap, !Ref Stage, " + key + "]\n"
	}

	// Substitui no arquivo com o conteúdo do mapa
	if err := replaceOnFile("stageMapStringReplacer", finalStageMapReplacer, templateFile); err != nil {
		log.Fatalf("error adding stage map value to sam template file: %v", err)
	}

	if err := replaceOnFile("EnvStringReplacer", finalEnvReplacer, templateFile); err != nil {
		log.Fatalf("error adding env value to sam template file: %v", err)
	}
}

func replaceCustomDomainValues(customDomain *string, hostedZone *string, templateFile string) {
	// Verifica se customDomain não é nil antes de desreferenciá-lo
	var customDomainValue string
	if customDomain != nil {
		customDomainValue = *customDomain
	} else {
		customDomainValue = "" // ou um valor padrão
	}

	// Verifica se hostedZone não é nil antes de desreferenciá-lo
	var hostedZoneValue string
	if hostedZone != nil {
		hostedZoneValue = *hostedZone
	} else {
		hostedZoneValue = "" // ou um valor padrão
	}

	if err := replaceOnFile("customDomainReplacerString", `customDomain: "`+customDomainValue+`"`, templateFile); err != nil {
		log.Fatalf("error adding custom domain value to sam template file: %v", err)
	}

	if err := replaceOnFile("hostedZoneReplacerString", `hostedZoneId: "`+hostedZoneValue+`"`, templateFile); err != nil {
		log.Fatalf("error adding hosted zone value to sam template file: %v", err)
	}
}

func replaceCustomDomainWithArnValues(customDomain *string, hostedZone *string, arn *string, templateFile string) {
	// Chama a função que substitui valores de domínio customizado
	replaceCustomDomainValues(customDomain, hostedZone, templateFile)

	// Verifica se arn não é nil antes de desreferenciá-lo
	var arnValue string
	if arn != nil {
		arnValue = *arn
	} else {
		arnValue = "" // ou um valor padrão
	}

	// Substitui o valor do ARN no arquivo de template
	if err := replaceOnFile("certificateArnReplacerString", `certificateArn: "`+arnValue+`"`, templateFile); err != nil {
		log.Fatalf("error adding arn value to sam template file: %v", err)
	}
}

func replaceTimeoutAndMemory(timeoutValue int, memoryValue int, templateFile string) {
	if err := replaceOnFile("timeoutReplacerString", string(timeoutValue), templateFile); err != nil {
		log.Fatalf("error adding timeout value to sam template file")
	}

	if err := replaceOnFile("memoryReplacerString", string(memoryValue), templateFile); err != nil {
		log.Fatalf("error adding memory value to sam template file")
	}
}

func copyFile(filePath string, destinyPath string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return os.WriteFile(destinyPath, fileContent, 0644)
}

func replaceOnFile(originalString string, replaceString string, filePath string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	replacedFile := []byte(strings.ReplaceAll(string(fileContent), originalString, replaceString))
	return os.WriteFile(filePath, replacedFile, 0644)
}
