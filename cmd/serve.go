package cmd

import (
	"os"

	gopserve "github.com/hophouse/gop/gopServe"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	directoryServeOption string
	authOption           string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a specific directory through an HTTP server.",
	Long:  "Serve a specific directory through an HTTP server.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if authOption == "" && logFileNameOption == "" {
			utils.NewLoggerStdout()
		} else {
			// If an output directory is specified, check if it exists and then move to it
			if directoryNameOption != "" {
				if _, err := os.Stat(directoryNameOption); os.IsNotExist(err) {
					utils.Log.Fatal("Specified directory by the option '-d' or '--directory', do not exists on the system.")
				}
				os.Chdir(directoryNameOption)
			} else {
				utils.CreateOutputDir(directoryNameOption)
			}

			// Init the logger and let close it later
			utils.NewLogger(LogFile, logFileNameOption)
			defer utils.CloseLogger(LogFile)
		}

		utils.Log.SetFlags(0)
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopserve.RunServeCmd(hostOption, portOption, directoryServeOption, authOption)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "127.0.0.1", "Define the proxy host.")
	serveCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "8000", "Define the proxy port.")
	serveCmd.PersistentFlags().StringVarP(&directoryServeOption, "directory", "d", ".", "Directory to serve.")
	serveCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
}
