package cli_utils

import (
	"fmt"
	"os"
	"text/template"
)

type InitCMDTemplateInfo struct {
	ProjectName            string
	GoModName              string
	TailWindFileName       string
	MainBinaryFileName     string
	MainServerPackageName  string
	MainServerFunctionName string
}

func ReplaceOnFile(templateFilePath string, outputFilePath string, templateStruct InitCMDTemplateInfo) error {
	templateFileData, err := os.ReadFile(templateFilePath)
	if err != nil {
		return err
	}
	data := template.Must(template.New(templateFilePath).Parse(string(templateFileData)))
	// Cria ou abre o arquivo de saída
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = data.Execute(outFile, templateStruct)
	if err != nil {
		return fmt.Errorf("error replacing go module name to file %s: %w", outputFilePath, err)
	}

	return nil
}
