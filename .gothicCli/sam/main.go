package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"

	gothicCliShared "github.com/felipegenef/gothic-cli/.gothicCli"
)

func main() {

	// Define command-line flags
	action := flag.String("action", "deploy", "Specify the action: deploy, delete or build")
	stage := flag.String("stage", "default", "Specify the deployment stage (default, dev, staging, prod)")
	flag.Parse()

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

	var stageValue string
	if stage != nil {
		stageValue = *stage
	} else {
		stageValue = "default" // or a default value
	}

	switch *action {
	case "deploy":

		samDeployCMD := exec.Command("sam", "deploy", "--stack-name", config.ProjectName+"-"+stageValue, "--parameter-overrides", "Stage="+stageValue, "--profile", config.Deploy.Profile)
		samDeployCMD.Stdout = os.Stdout
		samDeployCMD.Stdin = os.Stdin
		samDeployCMD.Stderr = os.Stderr

		// Run the command
		if err := samDeployCMD.Run(); err != nil {
			log.Fatalf("Error deploying app:%v", err)
		}

	case "delete":
		samDeleteCMD := exec.Command("sam", "delete", "--stack-name", config.ProjectName+"-"+stageValue, "--profile", config.Deploy.Profile)
		samDeleteCMD.Stdout = os.Stdout
		samDeleteCMD.Stdin = os.Stdin
		samDeleteCMD.Stderr = os.Stderr

		// Run the command
		if err := samDeleteCMD.Run(); err != nil {
			log.Fatalf("Error deleting app:%v", err)
		}
	case "build":
		samBuildCMD := exec.Command("sam", "build")
		samBuildCMD.Stdout = os.Stdout
		samBuildCMD.Stdin = os.Stdin
		samBuildCMD.Stderr = os.Stderr

		// Run the command
		if err := samBuildCMD.Run(); err != nil {
			log.Fatalf("Error building app:%v", err)
		}

	default:
		log.Fatalf("Invalid action specified: %s. Use 'deploy', 'delete' or 'build'.", *action)
	}
}
