package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of picasso",
	Long:  `All software has versions. This is picasso's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Picasso %s\n", version)
	},
}
