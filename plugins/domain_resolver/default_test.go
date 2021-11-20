package domain_resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Substr(t *testing.T) {
	sd := Subdomain{
		maxLen:       12,
		projectName:  "app",
		namespace:    "devops-test",
		index:        -1,
		nsPrefix:     "devops-",
		domainSuffix: "test.com",
	}
	assert.Equal(t, "app-devops-test.test.com", sd.CompleteSubdomain())
	assert.Equal(t, "app-test.test.com", sd.MediumSubdomain())
	assert.Equal(t, sd.SimpleSubdomain(), sd.SubStr())

	sd.maxLen = 17
	assert.Equal(t, sd.MediumSubdomain(), sd.SubStr())
	assert.True(t, len(sd.SubStr()) <= sd.maxLen)
	sd.maxLen = 9999
	assert.Equal(t, sd.CompleteSubdomain(), sd.SubStr())
	assert.True(t, len(sd.SubStr()) <= sd.maxLen)

	sd2 := Subdomain{
		maxLen:       12,
		projectName:  "app",
		namespace:    "devops-test",
		index:        1,
		nsPrefix:     "devops-",
		domainSuffix: "test.com",
	}

	assert.Equal(t, "app-devops-test-1.test.com", sd2.CompleteSubdomain())
	assert.Equal(t, "app-test-1.test.com", sd2.MediumSubdomain())
	assert.Equal(t, sd2.SimpleSubdomain(), sd2.SubStr())
	assert.Len(t, sd2.SubStr(), sd2.maxLen)

	sd3 := Subdomain{
		maxLen:       1,
		projectName:  "app",
		namespace:    "devops-test",
		index:        1,
		nsPrefix:     "devops-",
		domainSuffix: "test.com",
	}
	assert.Panics(t, func() {
		sd3.SubStr()
	})
}
