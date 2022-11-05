# GOP - GOPentest
![Go](https://github.com/HopHouse/gop/workflows/Go/badge.svg)

## Requirements
 - `Google Chrome` is needed by the screenshot and crawler commands.

## Installation
```powershell
go install github.com/hophouse/gop@latest
```

## Commands
```
GOP provides a toolbox to do pentest tasks.

Usage:
  gop [command]

Available Commands:
  crawler        Crawler command to crawl recursively or not a domain or a website.
  help           Help about any command
  host           Resolve hostname to get the IP address.
  osint          OSINT module.
  proxy          Set up a proxy to use
  scan           Scan commands to scan any kind of assets.
  schedule       Schedule a command to be executed at a precise time. If an option is not defined, the value of the current date will be taken.
  screenshot     Take screenshots of the supplied URLs. The program will take the stdin if no input file is passed as argument.
  search         Search for files on disk that matches a specific patterne. Regex or partial filename can be passed to the script.
  serve          Serve a specific directory through an HTTP server.
  shell          Set up a shell either reverse or bind.
  static-crawler The static crawler command will visit the supplied Website found link, script and style sheet inside it.
  visit          Visit supplied URLs.

Flags:
  -h, --help                      help for gop
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.

Use "gop [command] --help" for more information about a command.
```


### Proxy
If the proxy option is set, please use set the following option "Suppress Burp error messages in browser". If it is not done, all the response will have the status code `200 OK`. The option can be enabled in `Proxy -> Options -> Miscellaneous`.

#### CertificateAutority - CA
Generate a private key.
```powershell
openssl genrsa -out ca.key 4096
```

Generate the CA.
```powershell
openssl req -new -x509 -sha256 -key ca.key -out ca.crt -days 3650
```

## Serve
This command will serve file through a HTTP Web server. A few authentication methods can be added with the option `--auth`. The HTTP `Basic` and `NTLM` authentication are included.

### Add a new server

```Golang
type NewServer struct {
	Server Server
}

func (s NewServer) GetCertSubject() string {
	return s.Server.GetCertSubject()
}

func (s NewServer) GetServer(r *mux.Router, n *negroni.Negroni) (http.Server, error) {
	addr := fmt.Sprintf("%s:%s", s.Server.Host, s.Server.Port)
	logger.Printf("[+] HTTP server : %s://%s\n", s.Server.Scheme, addr)

	n.UseHandler(r)

	server := http.Server{
		Addr:    addr,
		Handler: n,
	}

	return server, nil
}

func (s NewServer) CreateRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	})

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	return r
}

func (s NewServer) CreateMiddleware() *negroni.Negroni {
	return s.Server.CreateMiddleware()
}
```