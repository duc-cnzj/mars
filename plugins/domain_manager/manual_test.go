package domain_manager

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var testDomain = "*.mars.local"
var tlsKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA3gRDD0PopDc8NMqcvq4RSBCkORyXIywmnODg9qnTRpuW2lET
tD01fTFhdVOvwjD2eUbAaPi9sJcq3ALBlxuveLIRcZrdWBnz8Wsl9cTI0js6TJQI
s76TQOTy9+V41tGj8jSABIN2kU6J2bX//c4e5Sh6Lm3Bo8b3u5IX/U1K+fbh0D/M
BBpnNhWWsqxYLg6xL77MzZW/7V4kDPDRV/X3/b9lhRdV6J7TeTwDjeMIoxyghfEK
hcjup1u569Y3JM25VZALmF+6RthF1VI06xfh1LK6YZ9/kdyWxGAqhimh2N6AgUgX
lZxrnL5IzxBeisT1oAyppu/eFC/gLwgjWrbqLQIDAQABAoIBAQC6rha9BTreqdsk
AKHf63l4gRW1OAcVehBVpPlELvQJ0NX9aGMLENSjGhx6jQ2iWY1F2Gn9WfoWPGW7
bU3eg1b+Q6xLPA6V/+yQuKnIh9jcfRS+Q1M14C7tSBXMv9TlvI0AIYjrQqc18zYS
K+i+PszXnzttQripv6mavHMTeCRtHMo+rGqVi+MOdjn6mgcNt7wHbwUwpZXgiyVG
l3LSpuwlVupbFmSlgKVG8ZY644BRrrD5tLpT9QjGV7ZhBV1ymlYyq3VqdTw229BL
yFni7W23D1ikZho9JNqLLshTbqCXfIoDjB5WETv4cl1CVVmqVWxTql/qp0jroV2Q
KA4tcgO5AoGBAP/+gGeLz4VpIZSHyci0YoW32WYorYw5KGZfRfJVW7wB/aZcuFeg
fwyRgzFnj5HK4PBRfmHS3igJeQTn7maiaRlf8dbnE0M/GKYI3j3GyEUzPWskn4gQ
ziR4MA2rVtDKuSqcnzXQQy9ZfGFRjBDco8+tVYYpOoqdzbyU5dE8WF//AoGBAN4F
j73OB4SlQ7mLNOjp1QF3feAspcuh/pR7hQMkoyQOslWEVOxB39v6vwzBnGA3EqfG
owKY4ts6PECOX4w85WDbu/4zPf7WzWzwya2YnwenUOh128PxCbpegQXLBijbPlTE
wM6odtsDbOod0JWFZrLv0DLrPuvQpViQZg+uADXTAoGAc4/fRV8vCknAV/3IkKsl
wrmREXYRijiPTU97Ev+HjuLTL5OxwBT65aCWuenHPQh57OLNC7oWgbptAFL3Iyv0
B/lxAhOEdZn5NZLRSNAAvoR4GHMK9XCornv3LWSIp26skljr4m4mtixOYtxeP4pr
BKh58DuSatr78kLBUGhOeN8CgYAXzjraXCP8Oggn9eAndSMMtDY/+imQyv7UBuZ9
Lsl7TUQb3UOJzYpmON2RTZUpz93lNWw3FBOG9BiPx3RBQipKF2Vx3Saxk3CVVMAb
J/ktRehr9G8q9EZZwFZPO7SeXtuxFSOjRPbxhs1/0NCTp6kaWJJXU1f8yvNfqqP2
3G5TVwKBgHa8oOGqYGXg+fVPR89cUYez8sCB9VE7BL/jWxXELbQEHfafdepFtnQL
udRNRUAFltGrjUUKjZlxVLeTvS7Iz8jU5dALIbq8lvayyPvnu+kTi7KrR3UZF+dE
UyFNwdSJdPSTjKqi0uUGZDZNIritIOe2n3TtjAniojrEz0l2fmkp
-----END RSA PRIVATE KEY-----
`
var tlsCrt = `
-----BEGIN CERTIFICATE-----
MIIDjTCCAnWgAwIBAgIUN4vMEdu3kX5rWf+tcda4cnmnSfwwDQYJKoZIhvcNAQEL
BQAwSTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAkNBMRYwFAYDVQQHEw1TYW4gRnJh
bmNpc2NvMRUwEwYDVQQDDAwqLm1hcnMubG9jYWwwHhcNMjIwMjE3MDkzNzAwWhcN
MjMwMjE3MDkzNzAwWjBJMQswCQYDVQQGEwJVUzELMAkGA1UECBMCQ0ExFjAUBgNV
BAcTDVNhbiBGcmFuY2lzY28xFTATBgNVBAMMDCoubWFycy5sb2NhbDCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBAN4EQw9D6KQ3PDTKnL6uEUgQpDkclyMs
Jpzg4Pap00abltpRE7Q9NX0xYXVTr8Iw9nlGwGj4vbCXKtwCwZcbr3iyEXGa3VgZ
8/FrJfXEyNI7OkyUCLO+k0Dk8vfleNbRo/I0gASDdpFOidm1//3OHuUoei5twaPG
97uSF/1NSvn24dA/zAQaZzYVlrKsWC4OsS++zM2Vv+1eJAzw0Vf19/2/ZYUXVeie
03k8A43jCKMcoIXxCoXI7qdbuevWNyTNuVWQC5hfukbYRdVSNOsX4dSyumGff5Hc
lsRgKoYpodjegIFIF5Wca5y+SM8QXorE9aAMqabv3hQv4C8II1q26i0CAwEAAaNt
MGswDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB
/wQCMAAwHQYDVR0OBBYEFOhGtW+dwGGnAb5GtIikfhqNYxBNMBcGA1UdEQQQMA6C
DCoubWFycy5sb2NhbDANBgkqhkiG9w0BAQsFAAOCAQEAWCY/NVnc08xtx32eAgpS
Om1r+SN31/e5L5jdmy4ir8CxFkV0d2bEzjZhbHnvaL7T7NDGgBb68oySJixepvV+
Bx7neoqRU9Sle1kqKG2FHJ4UQTibr127H/71/48w6/idEGekoUjqaFN6bgOhdChh
nlT6DxKgWybyGm5UG0dvphBMVanB6ymHVfTw7Ae8KcyFU1ySRMx7e48sLLhfQs+b
fOf9M9v9F+0dKDM5nIiXcL6dpiTjrqaetEZhPsVlAtl5+AcKi7aEximjPoZ731jF
IwyTV46MgBbGOba8vbUa22GBwZOpapzo9l1YM8h2ZyK/MUIl9LjpFhusZr2bKDmG
fA==
-----END CERTIFICATE-----
`

func TestManualDomainManager_Destroy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + (&ManualDomainManager{}).Name() + " plugin Destroy...").Times(1)
	mm := &ManualDomainManager{}
	mm.Destroy()
}

func TestManualDomainManager_GetCertSecretName(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	mm := &ManualDomainManager{}
	l.EXPECT().Info("[Plugin]: " + mm.Name() + " plugin Initialize...").Times(1)
	mm.Initialize(map[string]any{
		"ns_prefix":       "pfx",
		"tls_crt":         tlsCrt,
		"tls_key":         tlsKey,
		"wildcard_domain": testDomain,
	})
	assert.Equal(t, ManualCertSecretName, mm.GetCertSecretName("", 1))
}

func TestManualDomainManager_GetCerts(t *testing.T) {
	mm := &ManualDomainManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + mm.Name() + " plugin Initialize...").Times(1)
	mm.Initialize(map[string]any{
		"ns_prefix":       "pfx",
		"tls_crt":         tlsCrt,
		"tls_key":         tlsKey,
		"wildcard_domain": testDomain,
	})
	name, key, crt := mm.GetCerts()
	assert.Equal(t, ManualCertSecretName, name)
	assert.Equal(t, tlsKey, key)
	assert.Equal(t, tlsCrt, crt)
}

func TestManualDomainManager_GetClusterIssuer(t *testing.T) {
	mm := &ManualDomainManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + mm.Name() + " plugin Initialize...").Times(1)
	mm.Initialize(map[string]any{
		"ns_prefix":       "pfx",
		"tls_crt":         tlsCrt,
		"tls_key":         tlsKey,
		"wildcard_domain": testDomain,
	})
	assert.Equal(t, "", mm.GetClusterIssuer())
}

func TestManualDomainManager_GetDomain(t *testing.T) {
	mm := &ManualDomainManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + mm.Name() + " plugin Initialize...").Times(1)
	mm.Initialize(map[string]any{
		"ns_prefix":       "pfx",
		"tls_crt":         tlsCrt,
		"tls_key":         tlsKey,
		"wildcard_domain": testDomain,
	})
	preOccupiedLen := 0
	projectName := "pj"
	namespace := "ns"
	assert.Equal(t, Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        -1,
		nsPrefix:     mm.nsPrefix,
		domainSuffix: mm.domainSuffix,
	}.SubStr(), mm.GetDomain(projectName, namespace, preOccupiedLen))
}

func TestManualDomainManager_GetDomainByIndex(t *testing.T) {
	mm := &ManualDomainManager{}
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + mm.Name() + " plugin Initialize...").Times(1)
	mm.Initialize(map[string]any{
		"ns_prefix":       "pfx",
		"tls_crt":         tlsCrt,
		"tls_key":         tlsKey,
		"wildcard_domain": testDomain,
	})
	idx := 0
	preOccupiedLen := 0
	projectName := "pj"
	namespace := "ns"
	assert.Equal(t, Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        idx,
		nsPrefix:     mm.nsPrefix,
		domainSuffix: mm.domainSuffix,
	}.SubStr(), mm.GetDomainByIndex(projectName, namespace, idx, preOccupiedLen))
}

func TestManualDomainManager_Initialize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Info("[Plugin]: " + (&ManualDomainManager{}).Name() + " plugin Initialize...").Times(1)
	mm := &ManualDomainManager{}
	mm.Initialize(map[string]any{
		"ns_prefix":       "pfx",
		"tls_crt":         tlsCrt,
		"tls_key":         tlsKey,
		"wildcard_domain": testDomain,
	})
	assert.Equal(t, testDomain, mm.wildcardDomain)
	assert.Equal(t, "mars.local", mm.domainSuffix)
	assert.Equal(t, "pfx", mm.nsPrefix)
	assert.Equal(t, tlsCrt, mm.tlsCrt)
	assert.Equal(t, tlsKey, mm.tlsKey)
	assert.Error(t, mm.Initialize(map[string]any{
		"ns_prefix":       "",
		"tls_crt":         "",
		"tls_key":         "",
		"wildcard_domain": "",
	}))
	assert.Error(t, mm.Initialize(map[string]any{
		"ns_prefix":       "pfx",
		"tls_crt":         tlsCrt,
		"tls_key":         tlsKey,
		"wildcard_domain": "xxx.com",
	}))
}

func TestManualDomainManager_Name(t *testing.T) {
	assert.Equal(t, "manual_domain_manager", (&ManualDomainManager{}).Name())
}
