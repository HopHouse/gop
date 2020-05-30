# GoPentest

## Visit
The list of URL that need to be visites can be passed to stdin if the option `--stdin` is set or `-i/--input-file`.
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