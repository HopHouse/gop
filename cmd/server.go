package cmd

import (
	"io/ioutil"
	"net"
	"sync"

	"github.com/gobuffalo/packr/v2"
	gopserver "github.com/hophouse/gop/gopServer"
	"github.com/hophouse/gop/utils/logger"
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
	vhostOption            string
	exfilUrlOption         string
	httpsOption            bool
	customHTMLFile         string
)

// serverCmd represents the serve command
var serverCmd = &cobra.Command{
	Use: "server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)

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
		if httpsOption {
			gopserver.RunServerHTTPSCmd(hostOption, portOption, directoryServeOption, authOption, realmOption)
		} else {
			gopserver.RunServerHTTPCmd(hostOption, portOption, directoryServeOption, authOption, realmOption)
		}
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

var serverRedirectHTTPCmd = &cobra.Command{
	Use:   "redirect",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		gopserver.RunRedirectServerHTTPCmd(hostOption, portOption, vhostOption, dstUrlOption, httpsOption)
	},
}

var serverJSExfilHTTPCmd = &cobra.Command{
	Use:   "JSExfil",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		js := &gopserver.JavascriptExfilStruct{
			Host:     hostOption,
			Port:     portOption,
			Vhost:    vhostOption,
			ExfilUrl: exfilUrlOption,
			Scheme:   "http",
			Box:      packr.New("JSExfil", "../gopServer/JSExfil"),
			InputMu:  &sync.Mutex{},
		}

		if customHTMLFile != "" {
			index, err := ioutil.ReadFile(customHTMLFile)
			if err != nil {
				logger.Fatalln(err)
			}

			err = js.Box.AddBytes("index.html", index)
			if err != nil {
				logger.Fatalln(err)
			}
		}

		if httpsOption {
			js.Scheme = "https"
			js.RunJavascriptExfilHTTPSServerCmd()
		} else {
			js.RunJavascriptExfilHTTPServerCmd()
		}
	},
}

func init() {
	serverCmd.AddCommand(serverHTTPCmd)
	serverCmd.AddCommand(serverReverseHTTPProxyHTTPCmd)
	serverCmd.AddCommand(serverReverseHTTPSProxyHTTPCmd)
	serverCmd.AddCommand(serverGoPhishProxyHTTPCmd)
	serverCmd.AddCommand(serverRedirectHTTPCmd)
	serverCmd.AddCommand(serverJSExfilHTTPCmd)

	serverCmd.PersistentFlags().StringVarP(&interfaceOption, "interface", "i", "", "Interface to take IP adress.")
	serverCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "0.0.0.0", "Define the proxy host.")
	serverCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "8000", "Define the proxy port.")

	serverHTTPCmd.PersistentFlags().StringVarP(&directoryServeOption, "directory", "d", ".", "Directory to serve.")
	serverHTTPCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverHTTPCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
	serverHTTPCmd.PersistentFlags().BoolVarP(&httpsOption, "https", "", false, "Define whether or not an SSL/TLS layer is added.")

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

	serverRedirectHTTPCmd.PersistentFlags().StringVarP(&dstUrlOption, "destination", "d", "http://127.0.0.1:80", "Destination where traffic will be redirected.")
	serverRedirectHTTPCmd.PersistentFlags().StringVarP(&vhostOption, "vhost", "v", "", "Virtual host to use for the server.")
	serverRedirectHTTPCmd.PersistentFlags().BoolVarP(&httpsOption, "https", "", false, "Define whether or not an SSL/TLS layer is added.")

	serverJSExfilHTTPCmd.PersistentFlags().StringVarP(&vhostOption, "vhost", "v", "", "Virtual host to use for the server.")
	serverJSExfilHTTPCmd.PersistentFlags().StringVarP(&exfilUrlOption, "exfil-url", "", "", "Exfil URL in a form of http(s)://domain.tld.")
	serverJSExfilHTTPCmd.PersistentFlags().BoolVarP(&httpsOption, "https", "", false, "Define whether or not an SSL/TLS layer is added.")
	serverJSExfilHTTPCmd.PersistentFlags().StringVarP(&customHTMLFile, "custom-html", "", "", "Define a custom HTML file to use.")
}
