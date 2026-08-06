package domainmanager

import (
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
)

// maxDomainLength lets encrypt 对 subdomain 的长度限制为 64，所有域名生成共用此上限。
const maxDomainLength = 64

// Subdomain 承载一个子域名的各组成部分，SubStr 按长度约束依次降级拼接。
type Subdomain struct {
	maxLen       int
	projectName  string
	namespace    string
	index        int
	nsPrefix     string
	domainSuffix string
}

// SubStr 在不超过 maxLen 的前提下，按 完整版 → 去掉 ns 前缀的中等版 → 哈希截断的简单版 依次降级。
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

// HasIndex 是否带 index 后缀（index 为 -1 表示不带）。
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
	nname := strings.TrimPrefix(s.namespace, s.nsPrefix)
	if s.HasIndex() {
		return fmt.Sprintf("%s-%s-%d.%s", s.projectName, nname, s.index, s.domainSuffix)
	}

	return fmt.Sprintf("%s-%s.%s", s.projectName, nname, s.domainSuffix)
}

// SimpleSubdomain 简单版本, 用项目名+namespace 的哈希截断保证长度合法
func (s Subdomain) SimpleSubdomain() string {
	leftLen := s.maxLen - len(s.domainSuffix) - 1
	if leftLen <= 0 {
		panic(fmt.Errorf("substr error: max len: %d, left len: %d, domainSuffix: %s, project: %s, ns: %s, index: %d", s.maxLen, leftLen, s.domainSuffix, s.projectName, s.namespace, s.index))
	}
	var str = fmt.Sprintf("%s-%s", s.projectName, s.namespace)
	if s.HasIndex() {
		str = fmt.Sprintf("%s-%s-%d", s.projectName, s.namespace, s.index)
	}
	ss := substr(hasher.Hash(str), leftLen)

	return fmt.Sprintf("%s.%s", ss, s.domainSuffix)
}

// substr 截取字符串前 length 个字符，长度不足时原样返回。
func substr(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[0:length]
}
