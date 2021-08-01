package cmd

import (
	"net"
	"os"

	gopserver "github.com/hophouse/gop/gopServer"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	directoryServeOption   string
	dstUrlOption           string
	interfaceOption        string
	authOption             string
	realmOption            string
	redirectPrefixOption   string
	gophishUrlOption       string
	gophishWhiteListOption []string
)

// serverCmd represents the serve command
var serverCmd = &cobra.Command{
	Use: "server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if authOption == "" {
			utils.NewLoggerStdoutDateTimeFile()
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
}

var serverHTTPCmd = &cobra.Command{
	Use:   "http",
	Short: "Serve a specific directory through an HTTP server.",
	Run: func(cmd *cobra.Command, args []string) {
		gopserver.RunServerHTTPCmd(hostOption, portOption, directoryServeOption, authOption, realmOption)
	},
}

var serverReverseHTTPProxyHTTPCmd = &cobra.Command{
	Use:   "reverse-http-proxy",
	Short: "Act as an HTTP reverse proxy that will redirect the user to the specified destination. A prefix can be specified.",
	Run: func(cmd *cobra.Command, args []string) {
		gopserver.RunRedirectProxyHTTPCmd(hostOption, portOption, dstUrlOption, authOption, realmOption, redirectPrefixOption)
	},
}

var serverReverseHTTPSProxyHTTPCmd = &cobra.Command{
	Use:   "reverse-https-proxy",
	Short: "Act as an HTTP reverse proxy that will redirect the user to the specified destination. A prefix can be specified.",
	Run: func(cmd *cobra.Command, args []string) {
		gopserver.RunRedirectProxyHTTPSCmd(hostOption, portOption, dstUrlOption, authOption, realmOption, redirectPrefixOption)
	},
}

var serverGoPhishProxyHTTPCmd = &cobra.Command{
	Use:   "gophish-proxy",
	Short: "Act as an HTTP server that will do 2 things. It will first redirect the user to a specified location. Then it will log the user information into a specified GoPhish server.",
	Run: func(cmd *cobra.Command, args []string) {
		gopserver.RunGoPhishProxyHTTPCmd(hostOption, portOption, dstUrlOption, gophishUrlOption, authOption, realmOption, gophishWhiteListOption)
	},
}

func init() {
	serverCmd.AddCommand(serverHTTPCmd)
	serverCmd.AddCommand(serverReverseHTTPProxyHTTPCmd)
	serverCmd.AddCommand(serverReverseHTTPSProxyHTTPCmd)
	serverCmd.AddCommand(serverGoPhishProxyHTTPCmd)

	serverCmd.PersistentFlags().StringVarP(&interfaceOption, "interface", "i", "", "Interface to take IP adress.")
	serverCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "0.0.0.0", "Define the proxy host.")
	serverCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "8000", "Define the proxy port.")

	serverHTTPCmd.PersistentFlags().StringVarP(&directoryServeOption, "directory", "d", ".", "Directory to serve.")
	serverHTTPCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverHTTPCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")

	serverReverseHTTPProxyHTTPCmd.PersistentFlags().StringVarP(&dstUrlOption, "destination", "d", "http://127.0.0.1:80", "Destination where traffic will be redirected.")
	serverReverseHTTPProxyHTTPCmd.MarkPersistentFlagRequired("destination")
	serverReverseHTTPProxyHTTPCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverReverseHTTPProxyHTTPCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
	serverReverseHTTPProxyHTTPCmd.PersistentFlags().StringVarP(&redirectPrefixOption, "prefix", "", "", "Prefix used after the domain name/IP address to chose where to retdirect the request.")

	serverReverseHTTPSProxyHTTPCmd.PersistentFlags().StringVarP(&dstUrlOption, "destination", "d", "http://127.0.0.1:80", "Destination where traffic will be redirected.")
	serverReverseHTTPSProxyHTTPCmd.MarkPersistentFlagRequired("destination")
	serverReverseHTTPSProxyHTTPCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverReverseHTTPSProxyHTTPCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
	serverReverseHTTPSProxyHTTPCmd.PersistentFlags().StringVarP(&redirectPrefixOption, "prefix", "", "", "Prefix used after the domain name/IP address to chose where to retdirect the request.")

	serverGoPhishProxyHTTPCmd.Flags().StringVarP(&dstUrlOption, "destination", "d", "", "Destination where traffic will be redirected.")
	serverGoPhishProxyHTTPCmd.MarkFlagRequired("destination")
	serverGoPhishProxyHTTPCmd.Flags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverGoPhishProxyHTTPCmd.Flags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
	serverGoPhishProxyHTTPCmd.Flags().StringVarP(&gophishUrlOption, "gophish", "", "http://localhost/", "Url to the GoPhish server.")
	serverGoPhishProxyHTTPCmd.Flags().StringSliceVarP(&gophishWhiteListOption, "white-list", "", []string{"track"}, "Path where no authentucation will be asked.")
	serverGoPhishProxyHTTPCmd.MarkFlagRequired("gophish")
}
