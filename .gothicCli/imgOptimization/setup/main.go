package main

import (
	"log"
	"os/exec"
)

func main() {
	getResizeCMD := exec.Command("go", "get", "github.com/nfnt/resize")
	getWebpCMD := exec.Command("go", "get", "golang.org/x/image")
	// Make sure needed packages have been added to go.mod
	if err := getResizeCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
	if err := getWebpCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
	downloadResizeCMD := exec.Command("go", "mod", "download", "github.com/nfnt/resize")
	downloadWebpCMD := exec.Command("go", "mod", "download", "golang.org/x/image")
	// Make sure needed packages have been downloaded
	if err := downloadResizeCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
	if err := downloadWebpCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
}
