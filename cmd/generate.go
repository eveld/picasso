package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	generator "github.com/eveld/picasso/pkg"
	"github.com/spf13/cobra"
)

var outputPath string
var templatePath string
var variables []string
var csvPath string
var csvOutputVar string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates graphical assets from HCL2 templates",
	Run: func(cmd *cobra.Command, args []string) {
		parameters := readParameters(variables)

		if csvPath != "" {
			csvp, err := parseCSVParameters(csvPath)
			if err != nil {
				log.Fatal(err)
			}

			for _, p := range csvp {
				for k, v := range p {
					parameters[k] = v
				}

				template, err := generator.ParseTemplate(templatePath, parameters)
				if err != nil {
					log.Fatal(err)
				}

				hash := randSeq(4)

				base := "output"
				if csvOutputVar != "" {
					base = filepath.Base(parameters[csvOutputVar])
				}

				// Trim off any file extensions if present.
				file := strings.TrimSuffix(base, filepath.Ext(base))
				filename := fmt.Sprintf("%s-%s", file, hash)

				outputPath := fmt.Sprintf("%s/%s.png", filepath.Dir(outputPath), filename)
				err = generator.Generate(template, version, outputPath)
				if err != nil {
					log.Fatal(err)
				}
			}

		} else {
			template, err := generator.ParseTemplate(templatePath, parameters)
			if err != nil {
				log.Fatal(err)
			}

			err = generator.Generate(template, version, outputPath)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func init() {
	generateCmd.Flags().StringVarP(&templatePath, "template", "t", "", "Use this template to generate the asset")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output the generated asset to this path")
	generateCmd.Flags().StringArrayVarP(&variables, "var", "", nil, "Allows setting variables from the command line, variables are specified as a key and value, e.g --var key=value. Can be specified multiple times")
	generateCmd.Flags().StringVarP(&csvPath, "csv", "", "", "Path to read csv variables from")
	generateCmd.Flags().StringVarP(&csvOutputVar, "csv-var", "", "", "The variable to use when naming files generated from csv variables")

	generateCmd.MarkFlagRequired("template")
	generateCmd.MarkFlagRequired("output")

	rand.Seed(time.Now().UnixNano())
}

func readParameters(variables []string) map[string]string {
	parameters := map[string]string{}
	for _, variable := range variables {
		parts := strings.Split(variable, "=")
		parameters[parts[0]] = parts[1]
	}
	return parameters
}

func parseCSVParameters(file string) ([]map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvr := csv.NewReader(f)
	records, err := csvr.ReadAll()
	if err != nil {
		return nil, err
	}

	headers := []string{}
	csvp := []map[string]string{}
	for i, parameters := range records {
		if i == 0 {
			headers = parameters
		} else {
			par := lineToMap(headers, parameters)
			csvp = append(csvp, par)
		}
	}
	return csvp, nil
}

func lineToMap(headers []string, line []string) map[string]string {
	ret := map[string]string{}
	for i, header := range headers {
		ret[header] = line[i]
	}

	return ret
}
