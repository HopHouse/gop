package cmd

import (
	"net"
	"os"

	gopserve "github.com/hophouse/gop/gopServe"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	directoryServeOption string
	interfaceOption      string
	authOption           string
	realmOption          string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a specific directory through an HTTP server.",
	Long:  "Serve a specific directory through an HTTP server.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if authOption == "" {
			utils.NewLoggerStdout()
		} else {
			// If an output directory is specified, check if it exists and then move to it
			if directoryNameOption != "" {
				if _, err := os.Stat(directoryNameOption); os.IsNotExist(err) {
					utils.Log.Fatal("Specified directory by the option '-d' or '--directory', do not exists on the system.")
				}
				os.Chdir(directoryNameOption)
			} else {
				utils.CreateOutputDir(directoryNameOption, cmd.Name())
			}

			// Init the logger and let close it later
			utils.NewLogger(LogFile, logFileNameOption)
			defer utils.CloseLogger(LogFile)
		}

		utils.Log.SetFlags(0)

		if interfaceOption != "" {
			ief, err := net.InterfaceByName(interfaceOption)
			if err != nil { // get interface
				return
			}
			addrs, err := ief.Addrs()
			if err != nil { // get addresses
				return
			}
			for _, addr := range addrs { // get ipv4 address
				ipv4Addr := addr.(*net.IPNet).IP.To4()
				if ipv4Addr != nil {
					break
				}
				if ipv4Addr != nil {
					hostOption = ipv4Addr.String()
				}
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopserve.RunServeCmd(hostOption, portOption, directoryServeOption, authOption, realmOption)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().StringVarP(&interfaceOption, "interface", "i", "", "Interface to take IP adress.")
	serveCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "0.0.0.0", "Define the proxy host.")
	serveCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "8000", "Define the proxy port.")
	serveCmd.PersistentFlags().StringVarP(&directoryServeOption, "directory", "d", ".", "Directory to serve.")
	serveCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serveCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
}
