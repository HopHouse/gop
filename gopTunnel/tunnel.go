package goptunnel

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"strings"

	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/hophouse/gop/utils/logger"
)

type tunnelInterface interface {
	Listen() error
	Accept() error
	Dial() error
	Close() error
	Clone() tunnelInterface
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	RemoteAddr() string
}

type PlainTextTunnel struct {
	Protocol string
	Address  string
	Conn     net.Conn
	Listener net.Listener
}

func (t *PlainTextTunnel) Listen() error {
	var err error
	t.Listener, err = net.Listen(t.Protocol, t.Address)
	return err
}

func (t *PlainTextTunnel) Accept() error {
	var err error
	t.Conn, err = t.Listener.Accept()
	return err
}

func (t *PlainTextTunnel) Dial() error {
	var err error
	t.Conn, err = net.Dial(t.Protocol, t.Address)
	return err
}

func (t *PlainTextTunnel) Close() error {
	return t.Conn.Close()
}

func (t *PlainTextTunnel) Clone() tunnelInterface {
	newTun := PlainTextTunnel{
		Protocol: t.Protocol,
		Address:  t.Address,
		Conn:     t.Conn,
		Listener: t.Listener,
	}

	return &newTun
}

func (t *PlainTextTunnel) Read(b []byte) (n int, err error) {
	/*
		n, err = t.Conn.Read(b)
		if n > 0 {
			logger.Println("[+] Read", b[:n])
		}
		return n, err
	*/
	return t.Conn.Read(b)
}

func (t *PlainTextTunnel) Write(b []byte) (n int, err error) {
	// logger.Println("[+] Write", b)
	return t.Conn.Write(b)
}

func (t *PlainTextTunnel) RemoteAddr() string {
	return t.Conn.RemoteAddr().String()
}

type TlsTunnel struct {
	Protocol string
	Address  string
	Conn     net.Conn
	Listener net.Listener
}

func (t *TlsTunnel) Listen() error {
	var err error
	serverCert, serverKey := gopproxy.GenerateCA()

	caBytes, err := x509.CreateCertificate(rand.Reader, serverCert, serverCert, serverKey.Public(), serverKey)
	if err != nil {
		return err
	}

	// TODO Factorise code
	serverCertPEM := new(bytes.Buffer)
	err = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		logger.Println(err)
	}

	serverPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
	})
	if err != nil {
		logger.Println(err)
	}

	cer, err := tls.X509KeyPair(serverCertPEM.Bytes(), serverPrivKeyPEM.Bytes())
	if err != nil {
		return err
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	// Listen for incoming connections.
	t.Listener, err = tls.Listen(t.Protocol, t.Address, config)
	if err != nil {
		errString := fmt.Sprintf("Error listening : %s", err.Error())
		return errors.New(errString)
	}

	return nil
}

func (t *TlsTunnel) Accept() error {
	var err error
	t.Conn, err = t.Listener.Accept()
	return err
}

func (t *TlsTunnel) Dial() error {
	var err error

	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	t.Conn, err = tls.Dial(t.Protocol, t.Address, config)
	if err != nil {
		return err
	}

	return err
}

func (t *TlsTunnel) Read(b []byte) (n int, err error) {
	return t.Conn.Read(b)
}

func (t *TlsTunnel) Write(b []byte) (n int, err error) {
	return t.Conn.Write(b)
}

func (t *TlsTunnel) Close() error {
	return t.Conn.Close()
}

func (t *TlsTunnel) RemoteAddr() string {
	return t.Conn.RemoteAddr().String()
}

func (t *TlsTunnel) Clone() tunnelInterface {
	newTun := TlsTunnel{
		Protocol: t.Protocol,
		Address:  t.Address,
		Conn:     t.Conn,
		Listener: t.Listener,
	}

	return &newTun
}

type HTTPPlainTextTunnel struct {
	Protocol string
	Address  string
	Conn     net.Conn
	Listener net.Listener
}

func (t *HTTPPlainTextTunnel) Listen() error {
	var err error
	t.Listener, err = net.Listen(t.Protocol, t.Address)
	return err
}

func (t *HTTPPlainTextTunnel) Accept() error {
	var err error
	t.Conn, err = t.Listener.Accept()
	return err
}

func (t *HTTPPlainTextTunnel) Dial() error {
	var err error
	t.Conn, err = net.Dial(t.Protocol, t.Address)
	return err
}

func (t *HTTPPlainTextTunnel) Read(b []byte) (n int, err error) {
	prefix := "User-Agent: "
	bSize := 0

	var content []byte

	scanner := bufio.NewScanner(t.Conn)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			payload := strings.SplitAfter(line, prefix)
			if len(payload) < 2 {
				return 0, errors.New("payload could not be parsed")
			}

			content, err = base64.StdEncoding.DecodeString(payload[1])
			if err != nil {
				logger.Println(err)
				return 0, err
			}

			bSize = len(content)
			for i := 0; i < bSize; i++ {
				b[i] = content[i]
			}

			break
		}
	}

	/*
		c := bufio.NewReader(t.Conn)
		for {

			// read the full message, or return an error
			content, err = c.ReadBytes('\n')
			if err != nil {
				return 0, err
			}

			if len(content) > 0 {
				logger.Println(n)
				logger.Printf("received %x\n", b[:int(n)])

				if strings.HasPrefix(string(content), prefix) {
					payload := strings.SplitAfter(string(content), prefix)
					if len(payload) < 2 {
						return 0, errors.New("payload could not be parsed")
					}

					contentb64, err := base64.StdEncoding.DecodeString(payload[1])
					if err != nil {
						logger.Println(err)
						return 0, err
					}

					bSize = len(contentb64)
					for i := 0; i < bSize; i++ {
						b[i] = content[i]
					}
					break
				}
			}
		}
	*/

	return bSize, nil
}

func (t *HTTPPlainTextTunnel) Write(b []byte) (n int, err error) {
	logger.Println("[+] Write function")
	contenBeginString := "GET / HTTP/1.1\r\nHost: 1.1.1.1\r\nUser-Agent: "
	contentBegin := []byte(contenBeginString)
	contentEnd := []byte("\r\n\r\n")

	contentb64 := base64.StdEncoding.EncodeToString(b)
	logger.Println("[+] b64 string : ", contentb64)

	written := append(contentBegin, []byte(contentb64)...)
	written = append(written, contentEnd...)

	return t.Conn.Write([]byte(written))
}

func (t *HTTPPlainTextTunnel) Close() error {
	return t.Conn.Close()
}

func (t *HTTPPlainTextTunnel) RemoteAddr() string {
	return t.Conn.RemoteAddr().String()
}

func (t *HTTPPlainTextTunnel) Clone() tunnelInterface {
	newTun := HTTPPlainTextTunnel{
		Protocol: t.Protocol,
		Address:  t.Address,
		Conn:     t.Conn,
		Listener: t.Listener,
	}

	return &newTun
}

type UDPPlainTextTunnel struct {
	Protocol string
	Address  string
	Conn     net.Conn
	Listener net.PacketConn
}

func (t *UDPPlainTextTunnel) Listen() error {
	var err error
	t.Listener, err = net.ListenPacket(t.Protocol, t.Address)
	return err
}

func (t *UDPPlainTextTunnel) Accept() error {
	var err error
	return err
}

func (t *UDPPlainTextTunnel) Dial() error {
	var err error
	t.Conn, err = net.Dial(t.Protocol, t.Address)
	return err
}

func (t *UDPPlainTextTunnel) Close() error {
	return t.Conn.Close()
}

func (t *UDPPlainTextTunnel) Clone() tunnelInterface {
	newTun := UDPPlainTextTunnel{
		Protocol: t.Protocol,
		Address:  t.Address,
		Conn:     t.Conn,
		Listener: t.Listener,
	}

	return &newTun
}

func (t *UDPPlainTextTunnel) Read(b []byte) (n int, err error) {
	/*
		n, err = t.Conn.Read(b)
		if n > 0 {
			logger.Println("[+] Read", b[:n])
		}
		return n, err
	*/
	return t.Conn.Read(b)
}

func (t *UDPPlainTextTunnel) Write(b []byte) (n int, err error) {
	// logger.Println("[+] Write", b)
	return t.Conn.Write(b)
}

func (t *UDPPlainTextTunnel) RemoteAddr() string {
	return ""
}
