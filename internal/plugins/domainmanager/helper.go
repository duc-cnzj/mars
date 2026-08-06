package domainmanager

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// validateTLSWildcardDomain 校验 tlsKey/tlsCrt 组成的证书里，DNSNames 是否包含 wildcardDomain。
func validateTLSWildcardDomain[T []byte | string](tlsKey T, tlsCrt T, wildcardDomain string) error {
	var (
		err         error
		pair        tls.Certificate
		certificate *x509.Certificate
	)
	pair, err = tls.X509KeyPair([]byte(tlsCrt), []byte(tlsKey))
	if err != nil {
		return err
	}

	// tls.X509KeyPair 成功即证明证书已可解析，此处的 ParseCertificate 恒不失败，错误可安全忽略。
	certificate, _ = x509.ParseCertificate(pair.Certificate[0])
	var validDomain bool
	for _, dnsName := range certificate.DNSNames {
		if dnsName == wildcardDomain {
			validDomain = true
			break
		}
	}
	if !validDomain {
		err = fmt.Errorf("域名和证书不匹配, cert dnsNames: %v, 域名: %s", certificate.DNSNames, wildcardDomain)
	}
	return err
}
