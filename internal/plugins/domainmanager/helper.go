package domainmanager

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// stringArg 从 args 读取 key 对应的字符串参数；key 缺失返回空串与 nil，
// 存在但类型不符返回错误。manual/certmanager/syncsecret 三个插件共用。
// 缺失与空串赋值等价（调用方结构体字段初始即零值），故无需返回 ok 布尔。
func stringArg(args map[string]any, key string) (string, error) {
	v, ok := args[key]
	if !ok {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%s must be string", key)
	}
	return s, nil
}

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
