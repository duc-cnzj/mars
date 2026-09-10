package transformer_test

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/stretchr/testify/assert"
)

func TestFromMember_NilInput(t *testing.T) {
	var member *biz.Member
	result := transformer.FromMember(member)
	assert.Nil(t, result)
}

func TestFromMember_ValidInput(t *testing.T) {
	member := &biz.Member{
		ID:    1,
		Email: "test@example.com",
	}
	result := transformer.FromMember(member)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, "test@example.com", result.Email)
}
