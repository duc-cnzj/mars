package domain_resolver

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
)

var (
	name            = "domain_resolver_default"
	maxDomainLength = 64
)

func init() {
	dr := &DefaultDomainResolver{}
	plugins.RegisterPlugin(dr.Name(), dr)
}

// DefaultDomainResolver 因为 lets encrypt 对 subdomain 长度要求为 64，所以需要处理。
type DefaultDomainResolver struct {
	nsPrefix string
}

func (d *DefaultDomainResolver) Name() string {
	return name
}

func (d *DefaultDomainResolver) Initialize(args map[string]interface{}) error {
	if p, ok := args["ns_prefix"]; ok {
		d.nsPrefix = p.(string)
	}
	mlog.Info("[Plugin]: " + d.Name() + " plugin Initialize...")
	return nil
}

func (d *DefaultDomainResolver) Destroy() error {
	mlog.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
	return nil
}

func (d *DefaultDomainResolver) GetDomainByIndex(domainSuffix, projectName, namespace string, index, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        index,
		nsPrefix:     d.nsPrefix,
		domainSuffix: domainSuffix,
	}.SubStr()
}

func (d *DefaultDomainResolver) GetDomain(domainSuffix, projectName, namespace string, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        -1,
		nsPrefix:     d.nsPrefix,
		domainSuffix: domainSuffix,
	}.SubStr()
}

type Subdomain struct {
	maxLen       int
	projectName  string
	namespace    string
	index        int
	nsPrefix     string
	domainSuffix string
}

func (s Subdomain) SubStr() string {
	if s.maxLen == 0 {
		return s.CompleteSubdomain()
	}

	if len(s.CompleteSubdomain()) <= s.maxLen {
		return s.CompleteSubdomain()
	}

	if len(s.MediumSubdomain()) <= s.maxLen {
		return s.MediumSubdomain()
	}

	return s.SimpleSubdomain()
}

func (s Subdomain) HasIndex() bool {
	return s.index != -1
}

// CompleteSubdomain 获取完整的名称 mars-devops-test-default.test.com
func (s Subdomain) CompleteSubdomain() string {
	if s.HasIndex() {
		return fmt.Sprintf("%s-%s-%d.%s", s.projectName, s.namespace, s.index, s.domainSuffix)
	}

	return fmt.Sprintf("%s-%s.%s", s.projectName, s.namespace, s.domainSuffix)
}

// MediumSubdomain 中等版本, 去掉了 ns "devops-" 前缀
func (s Subdomain) MediumSubdomain() string {
	nname := strings.TrimLeft(s.namespace, s.nsPrefix)
	if s.HasIndex() {
		return fmt.Sprintf("%s-%s-%d.%s", s.projectName, nname, s.index, s.domainSuffix)
	}

	return fmt.Sprintf("%s-%s.%s", s.projectName, nname, s.domainSuffix)
}

// SimpleSubdomain 简单版本
func (s Subdomain) SimpleSubdomain() string {
	leftLen := s.maxLen - len(s.domainSuffix) - 1
	if leftLen <= 0 {
		panic(errors.New(fmt.Sprintf("substr error: max len: %d, left len: %d, domainSuffix: %s, project: %s, ns: %s, index: %d", s.maxLen, leftLen, s.domainSuffix, s.projectName, s.namespace, s.index)))
	}
	var str = fmt.Sprintf("%s-%s", s.projectName, s.namespace)
	if s.HasIndex() {
		str = fmt.Sprintf("%s-%s-%d", s.projectName, s.namespace, s.index)
	}
	ss := substr(hash(str), leftLen)

	return fmt.Sprintf("%s.%s", ss, s.domainSuffix)
}

func substr(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[0:length]
}

func hash(data string) string {
	h := md5.New()
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}
