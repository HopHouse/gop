# GOP - GOPentest
![Go](https://github.com/HopHouse/gop/workflows/Go/badge.svg)

## Requirements 
 - `Google Chrome` is needed by the screenshot and crawler commands.

## Installation
```powershell
go get github.com/hophouse/gop
```

## Documentation
[https://pkg.go.dev/github.com/hophouse/gop](https://pkg.go.dev/github.com/hophouse/gop)

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

### Crawler
```
Crawler command to crawl recursively or not a domain or a website.

Usage:
  gop crawler [flags]

Flags:
  -t, --concurrency int   Thread used to take screenshot. (default 10)
      --delay int         Use this delay in seconds between each requests.      
  -h, --help              help for crawler
  -p, --proxy string      Use the specified proxy.
  -r, --recursive         Crawl the website recursively.
  -s, --screenshot        Take a screenshot of each internal resource found.    
  -u, --url string        URL to test.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")   
      --output-directory string   Use the following directory to output results.
```

### Host
```
Crawler command to crawl recursively or not a domain or a website.

Usage:
  gop crawler [flags]

Flags:
  -t, --concurrency int   Thread used to take screenshot. (default 10)
      --delay int         Use this delay in seconds between each requests.      
  -h, --help              help for crawler
  -p, --proxy string      Use the specified proxy.
  -r, --recursive         Crawl the website recursively.
  -s, --screenshot        Take a screenshot of each internal resource found.    
  -u, --url string        URL to test.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")   
      --output-directory string   Use the following directory to output results.
```

### Osint
```
OSINT module.

Usage:
  gop osint [command]

Available Commands:
  emailgen    Generate email based on input data. It will create all the possible variations based on the allowed delimiters.

Flags:
  -h, --help   help for osint

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.

Use "gop osint [command] --help" for more information about a command.
```

#### emailgen
```
Generate email based on input data. It will create all the possible variations based on the allowed delimiters.

Usage:
  gop osint emailgen [flags]

Flags:
      --delimiters strings   Delimiters to construct the mail address. (default [.,-,_,#,$,%,&,*,+,/,=,!,?,^,',`,{,|,},~])
  -d, --domain string        Domain used into the email address.
  -f, --firstname string     First name.
  -h, --help                 help for emailgen
  -s, --surname string       Surname.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.
```

### Proxy
```
Set up a proxy to use

Usage:
  gop proxy [flags]

Flags:
  -H, --Host string   Define the proxy host. (default "127.0.0.1")
  -P, --Port string   Define the proxy port. (default "8080")
  -h, --help          help for proxy
  -i, --intercept     Intercept traffic.
  -v, --verbose       Display more information about packets.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")   
      --output-directory string   Use the following directory to output results.
```

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

### Scan
```
Scan commands to scan any kind of assets.

Usage:
  gop scan [command]

Available Commands:
  file        Search for files on disk that matches a specific patterne. Regex or partial filename can be passed to the script.
  network     Port scan the network. Only valid IP address must be passed as input.

Flags:
  -h, --help   help for scan

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.

Use "gop scan [command] --help" for more information about a command.
```     

#### File
```
Search for files on disk that matches a specific patterne. Regex or partial filename can be passed to the script.

Usage:
  gop scan file [flags]

Flags:
      --blacklist-extensions strings   Extension that will be blacklisted. (default [exe,ttf,dll,svg,go,pyi,html,css,js,yar,md,.lnk,tex,settingcontent-ms,template,lnk,nasl,sys,nse,lib])
      --blacklist-location strings     Locations were the script will not look. (default [C:\Windows,C:\Users\Public])
  -t, --concurrency int                Number of threads used. (default 10)
  -h, --help                           help for file
      --only-files                     Only display found items that are files.
  -p, --path strings                   Locations were to look the script have to look.
  -s, --search strings                 Specify a file will all the pattern that need to be checked. (default [(?i)identifiants,(?i)password,(?i)mot de passe,(?i)motdepasse,(?i)compte(s)?,kdb(x)?,(?i)secret,key[0-9].db$,(?i)backup,.ntds$,SYSTEM$,SAM$,id_rsa.*])
      --whitelist-extensions strings   Extension that will be whithelisted. If specified the black list option is taken in consideration by the program. Exemple : msg, squlite, zip, backup

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.
```

#### Network
```
Port scan the network. Only valid IP address must be passed as input.

Usage:
  gop scan network [flags]

Flags:
  -t, --concurrency int     Number of threads used to take to scan. (default 5000)
  -h, --help                help for network
  -i, --input-file string   Input file with the IP addresses to scan. If no file is passed, then the stdin will be taken.
      --open                Display only open ports.
  -o, --output string       Display result format as :
                                - text
                                - grep
                                - short
                            Option text will display a human-readable outpit.
                            Option grep will displays the following format for each host :
                                ip,protcol,port,status.
                            Option short will activate the flag --open and will display value for each open port as :    
                                ip:port (default "text")
  -p, --port string         Ports to scan. Can be either :
                                - X,Y,Z
                                - X-Y
                                - X-Y,Z
                            Family of ports can also be passed :
                                - http
                                - ssh
                                - mail
                                - ...   
                            Options can be combined, example :
                                - 22,http,445,8080-8088
      --tcp                 Scan with the TCP protocol.
      --udp                 [WIP] Scan with the UDP protocol.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results..
```

### Schedule      
```
Schedule a command to be executed at a precise time. If an option is not defined, the value of the current date will be taken.

Usage:
  gop schedule [flags]

Flags:
  -c, --command string     Command to execute.
      --day int            Days of the month. (default -1)
  -h, --help               help for schedule
      --hour int           Hour of the day. (default -1)
  -m, --minute int         Minute of the hour. (default -1)
      --month int          Month of the year. (default -1)
      --plus-days int      Add this number of days from the execution to execute the command.
      --plus-hours int     Add this number of hours from the execution to execute the command.
      --plus-minutes int   Add this number of minutes from the execution to execute the command.
      --plus-seconds int   Add this number of seconds from the execution to execute the command.
  -s, --second int         Second of the minute. (default -1)
      --year int           Year were the command will be launched. (default -1)

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.
```          

## Screenshot    
The screenshot command is using `chromecdp` to take screenshot. 

```
Take screenshots of the supplied URLs. The program will take the stdin if no input file is passed as argument.

Usage:
  gop screenshot [flags]

Flags:
  -t, --concurrency int     Thread used to take screenshot. (default 5)
      --delay int           Use this delay in seconds between requests. (default 1)
  -h, --help                help for screenshot
  -i, --input-file string   Use the specified cookie.
  -p, --proxy string        Use this proxy to visit the pages.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")      
      --output-directory string   Use the following directory to output results. 
```          

## Serve
This command will serve file through a HTTP Web server. A few authentication methods can be added with the option `--auth`. The HTTP `Basic` and `NTLM` authentication are included.
```
Serve a specific directory through an HTTP server.

Usage:
  gop serve [flags]

Flags:
  -H, --Host string        Define the proxy host. (default "127.0.0.1")
  -P, --Port string        Define the proxy port. (default "8000")
  -a, --auth string        Add an authentication option to the server. Could be either "Basic" or "NTLM".
  -d, --directory string   Directory to serve. (default ".")
  -h, --help               help for serve
      --realm string       Realm used for the "Basic" authentication.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.
```

## Shell         
```
Set up a shell either reverse or bind.

Usage:
  gop shell [flags]

Flags:
  -H, --Host string   Define the proxy host. (default "127.0.0.1")
  -m, --Mode string   Define the mode where the shell is runned : blind or reverse. (default "reverse")
  -P, --Port string   Define the proxy port. (default "8000")
  -h, --help          help for shell

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.
```          

## Static-crawler
```
The static crawler command will visit the supplied Website found link, script and style sheet inside it.

Usage:
  gop static-crawler [flags]

Flags:
  -t, --concurrency int   Thread used to take screenshot. (default 10)
  -c, --cookie string     Use the specified cookie.
      --delay int         Use this delay in seconds between each requests.      
  -h, --help              help for static-crawler
  -p, --proxy string      Use the specified proxy.
  -r, --recursive         Crawl the website recursively.
      --report            Generate a report.
  -s, --screenshot        Take a screenshot on each visited link.
  -u, --url string        URL to test.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")   
      --output-directory string   Use the following directory to output results.
```          

## Visit
The list of URL that needs to be visited can be passed to stdin if the option `--stdin` is set or `-i/--input-file`.
```
Visit supplied URLs.

Usage:
  gop visit [flags]

Flags:
      --burp                Set the proxy directly to the default address and port of burp.
  -h, --help                help for visit
  -i, --input-file string   Use the specified cookie.
  -p, --proxy string        Use this proxy to visit the pages.

Global Flags:
      --logfile string            Set a custom log file. (default "logs.txt")
      --output-directory string   Use the following directory to output results.
```          

## Tee
	- [ ] Tee command to load also load the command into a file
	
## CI/CD
.gitlab-co.yml file.
```
image: golang:1.10

stages:
  - build

before_script:
  - go get -u github.com/golang/dep/cmd/dep
  - mkdir -p $GOPATH/src/github.com/hophouse/gop
  - cd $GOPATH/src/github.com/hophouse/gop
  - ln -s $CI_PROJECT_DIR
  - cd $CI_PROJECT_NAME
  - dep ensure
  - GOOS=windows GOARCH=amd64 go install
  - GOOS=linux GOARCH=amd64 go install

after_script:
  - mv workdir release

build:
  stage: build
  artifacts:
    paths:
      - release
    name: artifact:build
    when: on_success
    expire_in: 1 week
  script:
    - make
```
