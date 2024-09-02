package transformer

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/stretchr/testify/assert"
)

func TestFromMemberNilInput(t *testing.T) {
	var member *repo.Member
	result := FromMember(member)
	assert.Nil(t, result)
}

func TestFromMemberValidInput(t *testing.T) {
	member := &repo.Member{
		ID:    1,
		Email: "test@example.com",
	}
	result := FromMember(member)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, "test@example.com", result.Email)
}
