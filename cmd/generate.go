package cmd

import (
	"log"
	"strings"

	generator "github.com/eveld/picasso/pkg"
	"github.com/spf13/cobra"
)

var outputPath string
var templatePath string
var variables []string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates graphical assets from HCL2 templates",
	Run: func(cmd *cobra.Command, args []string) {
		parameters := map[string]string{}
		for _, variable := range variables {
			parts := strings.Split(variable, "=")
			parameters[parts[0]] = parts[1]
		}

		template, err := generator.ParseTemplate(templatePath, parameters)
		if err != nil {
			log.Fatal(err)
		}

		err = generator.Generate(template, outputPath)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	generateCmd.Flags().StringVarP(&templatePath, "template", "t", "", "Use this template to generate the asset")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output the generated asset to this path")
	generateCmd.Flags().StringSliceVarP(&variables, "var", "", nil, "Allows setting variables from the command line, variables are specified as a key and value, e.g --var key=value. Can be specified multiple times")

	generateCmd.MarkFlagRequired("template")
	generateCmd.MarkFlagRequired("output")
}
