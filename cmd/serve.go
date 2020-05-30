package cmd

import (

    "github.com/hophouse/gop/gopServe"
    "github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	directoryServeOption string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a specific directory through an HTTP server.",
	Long: "Serve a specific directory through an HTTP server.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.NewLoggerStdout()
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopserve.RunServeCmd(hostOption, portOption, directoryServeOption)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "127.0.0.1", "Define the proxy host.")
	serveCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "8000", "Define the proxy port.")
	serveCmd.PersistentFlags().StringVarP(&directoryServeOption, "directory", "d", ".", "Directory to serve.")
}
