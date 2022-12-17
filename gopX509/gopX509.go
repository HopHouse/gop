package gopX509

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"net"
	"time"

	"github.com/hophouse/gop/utils/logger"
)

func RunX509Names(addresses []string) error {

	names := make(map[string]interface{}, 0)

	for _, address := range addresses {

		conf := &tls.Config{
			InsecureSkipVerify: true,
		}

		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", address, conf)
		if err != nil {
			logger.Fprintf(logger.Writer(), "Error on %s : %s\n", address, err)
			continue
		}
		defer conn.Close()

		cert := conn.ConnectionState().PeerCertificates[0]

		newNames, err := ExtractNames(cert)
		if err != nil {
			logger.Fprintln(logger.Writer(), err)
			continue
		}

		for item := range newNames {
			logger.Fprintf(logger.Writer(), "[%s] %s\n", address, item)
			names[item] = nil
		}
	}

	for name := range names {
		logger.Println(name)
	}

	return nil
}

func ExtractNames(cert *x509.Certificate) (map[string]interface{}, error) {

	names := make(map[string]interface{}, 0)

	names[cert.Subject.CommonName] = nil

	for _, ext := range cert.Extensions {
		if ext.Id.Equal(asn1.ObjectIdentifier{2, 5, 29, 17}) {
			values := []asn1.RawValue{}
			_, err := asn1.Unmarshal(ext.Value, &values)
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				continue
			}

			for _, value := range values {
				names[string(value.Bytes)] = nil
			}
		}
	}

	for _, ext := range cert.ExtraExtensions {
		if ext.Id.Equal(asn1.ObjectIdentifier{2, 5, 29, 17}) {
			values := []asn1.RawValue{}
			_, err := asn1.Unmarshal(ext.Value, &values)
			if err != nil {
				logger.Fprintln(logger.Writer(), err)
				continue
			}

			for _, value := range values {
				names[string(value.Bytes)] = nil
			}
		}
	}

	return names, nil
}
