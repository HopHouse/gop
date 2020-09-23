# GOP - GOPentest
![Go](https://github.com/HopHouse/gop/workflows/Go/badge.svg)

```
GOP provide a help performing some pentest tasks.

Usage:
  gop [command]]

Available Commands:
  crawler     Crawler command to crawl passively or actively a domain or a website.
  help        Help about any command
  host        Resolve hostname to get the IP address.
  proxy       Set up a proxy to use
  screenshot  Take screenshots of the supplied URLs. The program will take the stdin if no input file is passed as argument.
  serve       Serve a specific directory through an HTTP server.
  shell       Set up a shell either reverse or bind.
  visit       Visit supplied URLs.

Flags:
  -h, --help                      help for gop
  -l, --logfile string            Set a custom log file. (default "logs.txt")
  -D, --output-directory string   Use the following directory to output results.
  -u, --url string                URL to test.

Use "gop [command] --help" for more information about a command.
```

## Visit
The list of URL that needs to be visited can be passed to stdin if the option `--stdin` is set or `-i/--input-file`.
### Proxy
If the proxy option is set, please use set the following option "Suppress Burp error messages in browser". If it is not done, all the response will have the status code `200 OK`. The option can be enabled in `Proxy -> Options -> Miscellaneous`.

## Proxy
### CertificateAutority - CA
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
