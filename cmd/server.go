package cmd

import (
	"log"
	"net"
	"path/filepath"
	"sync"

	goprelay "github.com/hophouse/gop/gopRelay"
	gopserver "github.com/hophouse/gop/gopServer"
	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var (
	directoryServeOption    string
	dstUrlOption            string
	interfaceOption         string
	authOption              string
	realmOption             string
	redirectPrefixOption    string
	gophishUrlOption        string
	gophishWhiteListOption  []string
	vhostOption             string
	exfilUrlOption          string
	httpsOption             bool
	relayTargetUrl          []string
	relaySMBServerHostPort  string
	relayHTTPServerHostPort string
	relayHTTPProxyHostPort  string
	// customHTMLFile         string
	// customJSFile           string
)

// serverCmd represents the serve command
var serverCmd = &cobra.Command{
	Use: "server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		rootCmd.PersistentPreRun(cmd, args)

		if !filepath.IsAbs(directoryServeOption) {
			directoryServeOption, err = filepath.Abs(filepath.Join(CurrentDirectory, directoryServeOption))
			if err != nil {
				logger.Fatalln(err)
			}
		}

		directoryServeOption = filepath.Clean(directoryServeOption)

		if interfaceOption != "" {
			ief, err := net.InterfaceByName(interfaceOption)
			if err != nil { // get interface
				logger.Fatalln(err)
			}
			addrs, err := ief.Addrs()
			if err != nil { // get addresses
				logger.Fatalln(err)
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
		fs := gopserver.FilseServer{
			Server: gopserver.Server{
				Host:   hostOption,
				Port:   portOption,
				Scheme: "http",
				Vhost:  vhostOption,
				Auth:   authOption,
				Realm:  realmOption,
			},
			Directory: directoryServeOption,
		}
		if httpsOption {
			fs.Server.Scheme = "https"
			log.Fatal(gopserver.RunServerHTTPSCmd(fs))
		} else {
			log.Fatal(gopserver.RunServerHTTPCmd(fs))
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
		log.Fatal(gopserver.RunRedirectServerHTTPCmd(hostOption, portOption, vhostOption, dstUrlOption, httpsOption, authOption, realmOption))
	},
}

var serverJSExfilHTTPCmd = &cobra.Command{
	Use:   "JSExfil",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		js := &gopserver.JavascriptExfilServer{
			Server: &gopserver.Server{
				Host:   hostOption,
				Port:   portOption,
				Scheme: "http",
				Vhost:  vhostOption,
			},
			ExfilUrl: exfilUrlOption,
			// Box:      packr.New("JSExfil", "../gopServer/JSExfil"),
			InputMu: &sync.Mutex{},
		}
		/*
			 * Disabled following the depreciation of packr v2 and the chose of go:embed
			if customHTMLFile != "" {
				index, err := os.ReadFile(customHTMLFile)
				if err != nil {
					logger.Fatalln(err)
				}

				err = js.Box.AddBytes("index.html", index)
				if err != nil {
					logger.Fatalln(err)
				}
			}
		*/

		/*
			 * Disabled following the depreciation of packr v2 and the chose of go:embed
			if customJSFile != "" {
				customJS, err := os.ReadFile(customJSFile)
				if err != nil {
					logger.Fatalln(err)
				}

				err = js.Box.AddBytes("custom.js", customJS)
				if err != nil {
					logger.Fatalln(err)
				}
			}
		*/
		if httpsOption {
			js.Server.Scheme = "https"
			log.Fatal(gopserver.RunServerHTTPSCmd(js))
		} else {
			log.Fatal(gopserver.RunServerHTTPCmd(js))
		}
	},
}

var serverRelayCmd = &cobra.Command{
	Use: "relay",
	Run: func(cmd *cobra.Command, args []string) {
		// Run local proxy
		go goprelay.RunLocalProxy()

		// Run the relay server
		go goprelay.RunRelayServer(relayTargetUrl)

		// Run the HTTP server
		go goprelay.RunHTTPServer(relayHTTPServerHostPort, goprelay.ProcessIncomingConnChan)

		goprelay.RunSMBServer(relaySMBServerHostPort, goprelay.ProcessIncomingConnChan)

		// Hang forever
		select {}
	},
}

func init() {
	serverCmd.AddCommand(serverHTTPCmd)
	serverCmd.AddCommand(serverReverseHTTPProxyHTTPCmd)
	serverCmd.AddCommand(serverReverseHTTPSProxyHTTPCmd)
	serverCmd.AddCommand(serverGoPhishProxyHTTPCmd)
	serverCmd.AddCommand(serverRedirectHTTPCmd)
	serverCmd.AddCommand(serverJSExfilHTTPCmd)
	serverCmd.AddCommand(serverRelayCmd)

	serverCmd.PersistentFlags().StringVarP(&interfaceOption, "interface", "i", "", "Interface to take IP adress.")
	serverCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "0.0.0.0", "Define the proxy host.")
	serverCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "8000", "Define the proxy port.")

	serverHTTPCmd.PersistentFlags().StringVarP(&directoryServeOption, "directory", "d", ".", "Directory to serve.")
	serverHTTPCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverHTTPCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
	serverHTTPCmd.PersistentFlags().BoolVarP(&httpsOption, "https", "", false, "Define whether or not an SSL/TLS layer is added.")

	serverReverseHTTPProxyHTTPCmd.PersistentFlags().StringVarP(&dstUrlOption, "destination", "d", "http://127.0.0.1:80", "Destination where traffic will be redirected.")
	_ = serverReverseHTTPProxyHTTPCmd.MarkPersistentFlagRequired("destination")
	serverReverseHTTPProxyHTTPCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverReverseHTTPProxyHTTPCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
	serverReverseHTTPProxyHTTPCmd.PersistentFlags().StringVarP(&redirectPrefixOption, "prefix", "", "", "Prefix used after the domain name/IP address to chose where to retdirect the request.")

	serverReverseHTTPSProxyHTTPCmd.PersistentFlags().StringVarP(&dstUrlOption, "destination", "d", "http://127.0.0.1:80", "Destination where traffic will be redirected.")
	_ = serverReverseHTTPSProxyHTTPCmd.MarkPersistentFlagRequired("destination")
	serverReverseHTTPSProxyHTTPCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverReverseHTTPSProxyHTTPCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
	serverReverseHTTPSProxyHTTPCmd.PersistentFlags().StringVarP(&redirectPrefixOption, "prefix", "", "", "Prefix used after the domain name/IP address to chose where to retdirect the request.")

	serverGoPhishProxyHTTPCmd.Flags().StringVarP(&dstUrlOption, "destination", "d", "", "Destination where traffic will be redirected.")
	_ = serverGoPhishProxyHTTPCmd.MarkFlagRequired("destination")
	serverGoPhishProxyHTTPCmd.Flags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverGoPhishProxyHTTPCmd.Flags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")
	serverGoPhishProxyHTTPCmd.Flags().StringVarP(&gophishUrlOption, "gophish", "", "http://localhost/", "Url to the GoPhish server.")
	serverGoPhishProxyHTTPCmd.Flags().StringSliceVarP(&gophishWhiteListOption, "white-list", "", []string{"track"}, "Path where no authentucation will be asked.")
	_ = serverGoPhishProxyHTTPCmd.MarkFlagRequired("gophish")

	serverRedirectHTTPCmd.PersistentFlags().StringVarP(&dstUrlOption, "destination", "d", "http://127.0.0.1:80", "Destination where traffic will be redirected.")
	serverRedirectHTTPCmd.PersistentFlags().StringVarP(&vhostOption, "vhost", "v", "", "Virtual host to use for the server.")
	serverRedirectHTTPCmd.PersistentFlags().BoolVarP(&httpsOption, "https", "", false, "Define whether or not an SSL/TLS layer is added.")
	serverRedirectHTTPCmd.PersistentFlags().StringVarP(&authOption, "auth", "a", "", "Add an authentication option to the server. Could be either \"Basic\" or \"NTLM\".")
	serverRedirectHTTPCmd.PersistentFlags().StringVarP(&realmOption, "realm", "", "", "Realm used for the \"Basic\" authentication.")

	serverJSExfilHTTPCmd.PersistentFlags().StringVarP(&vhostOption, "vhost", "v", "", "Virtual host to use for the server.")
	serverJSExfilHTTPCmd.PersistentFlags().StringVarP(&exfilUrlOption, "exfil-url", "", "", "Exfil URL in a form of http(s)://domain.tld.")
	serverJSExfilHTTPCmd.PersistentFlags().BoolVarP(&httpsOption, "https", "", false, "Define whether or not an SSL/TLS layer is added.")
	// serverJSExfilHTTPCmd.PersistentFlags().StringVarP(&customHTMLFile, "custom-html", "", "", "Define a custom HTML file to use.")
	// serverJSExfilHTTPCmd.PersistentFlags().StringVarP(&customJSFile, "custom-js", "", "", "Define a custom JavaScript file to use.")

	serverRelayCmd.PersistentFlags().StringSliceVarP(&relayTargetUrl, "targets", "t", []string{}, "URLs of the targets.")
	_ = serverRelayCmd.MarkFlagRequired("targets")
	serverRelayCmd.PersistentFlags().StringVarP(&relayHTTPProxyHostPort, "http-proxy", "", "127.0.0.1:4000", "HTTP proxy to listen for navigating through the HTTP proxy (Default: 127.0.0.1:4000).")
	serverRelayCmd.PersistentFlags().StringVarP(&relayHTTPServerHostPort, "http-server", "", "0.0.0.0:80", "HTTP server to listen for traffic to impersonate user (Default: 0.0.0.0:80).")
	serverRelayCmd.PersistentFlags().StringVarP(&relaySMBServerHostPort, "smb-server", "", "0.0.0.0:445", "SMB server to listen for traffic to impersonate user (Default: 0.0.0.0:445).")
}
