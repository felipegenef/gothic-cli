package helpers

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type AwsSamHelper struct {
}

func NewAwsSamHelper() AwsSamHelper {
	return AwsSamHelper{}
}

func (t *AwsSamHelper) Build() error {
	samBuildCMD := exec.Command("sam", "build")
	samBuildCMD.Stdout = os.Stdout
	samBuildCMD.Stdin = os.Stdin
	samBuildCMD.Stderr = os.Stderr

	// Run the command
	err := samBuildCMD.Run()
	if err != nil {
		fmt.Printf("Error building AWS Sam app:%v", err)
	}
	return err
}

func (t *AwsSamHelper) Deploy(stage string, stackName string, awsProfile string) error {
	samDeployCMD := exec.Command("sam", "deploy", "--stack-name", stackName+"-"+stage, "--parameter-overrides", "Stage="+stage, "--profile", awsProfile)
	samDeployCMD.Stdout = os.Stdout
	samDeployCMD.Stdin = os.Stdin
	samDeployCMD.Stderr = os.Stderr

	// Run the command
	if err := samDeployCMD.Run(); err != nil {
		log.Fatalf("Error deploying app:%v", err)
	}

	return nil
}

func (t *AwsSamHelper) DeleteStack(stage string, stackName string, awsProfile string) error {

	samDeleteCMD := exec.Command("sam", "delete", "--stack-name", stackName+"-"+stage, "--profile", awsProfile)
	samDeleteCMD.Stdout = os.Stdout
	samDeleteCMD.Stdin = os.Stdin
	samDeleteCMD.Stderr = os.Stderr

	// Run the command
	err := samDeleteCMD.Run()
	if err != nil {
		log.Fatalf("Error deleting app:%v", err)
	}
	return err
}
