package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gkwa/belgianlake/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of belgianlake",
	Long:  `All software has versions. This is belgianlake's`,
	Run: func(cmd *cobra.Command, args []string) {
		buildInfo := version.GetBuildInfo()
		fmt.Println(buildInfo)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
