package gopX509

import (
	"crypto/tls"
	"encoding/asn1"

	"github.com/hophouse/gop/utils/logger"
)

func RunX509Names(address string) error {

	names := make(map[string]interface{}, 0)

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", address, conf)
	if err != nil {
		return err
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]

	names[cert.Subject.CommonName] = nil

	for _, ext := range cert.Extensions {
		if ext.Id.Equal(asn1.ObjectIdentifier{2, 5, 29, 17}) {
			values := []asn1.RawValue{}
			_, err := asn1.Unmarshal(ext.Value, &values)
			if err != nil {
				break
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
				break
			}

			for _, value := range values {
				names[string(value.Bytes)] = nil
			}
		}
	}

	for name, _ := range names {
		logger.Println(name)
	}

	return nil
}
