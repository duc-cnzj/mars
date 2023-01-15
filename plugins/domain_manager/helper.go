package domain_manager

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

func validateTelsWildcardDomain[T []byte | string](tlsKey T, tlsCrt T, wildcardDomain string) error {
	var (
		err         error
		pair        tls.Certificate
		certificate *x509.Certificate
	)
	pair, err = tls.X509KeyPair([]byte(tlsCrt), []byte(tlsKey))
	if err != nil {
		return err
	}

	certificate, err = x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return err
	}
	var validDomain bool
	for _, dnsName := range certificate.DNSNames {
		if dnsName == wildcardDomain {
			validDomain = true
			break
		}
		continue
	}
	if !validDomain {
		err = fmt.Errorf("域名和证书不匹配, cert dnsNames: %v, 域名: %s", certificate.DNSNames, wildcardDomain)
	}
	return err
}
