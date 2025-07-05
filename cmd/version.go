package cmd

import (
	"github.com/spf13/cobra"

	"github.com/gkwa/belgianlake/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of belgianlake",
	Long:  `All software has versions. This is belgianlake's`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := LoggerFrom(cmd.Context())
		buildInfo := version.GetBuildInfo()
		logger.Info("Version information", "buildInfo", buildInfo.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
