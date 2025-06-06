package helpers

import (
	"fmt"
	"os"
	"os/exec"
)

type TailwindHelper struct {
}

func NewTailwindHelper() TailwindHelper {
	return TailwindHelper{}
}

func (helper *TailwindHelper) Build() error {
	tailwindCmd := exec.Command("make", "css")
	tailwindCmd.Stdout = os.Stdout
	tailwindCmd.Stdin = os.Stdin
	tailwindCmd.Stderr = os.Stderr

	// Run the command
	err := tailwindCmd.Run()
	if err != nil {
		fmt.Printf("Error generating tailwind css:%v", err)
	}
	return err
}
