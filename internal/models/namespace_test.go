package models

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/utils/date"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNamespace_GetImagePullSecrets(t *testing.T) {
	m := Namespace{
		ID:               1,
		Name:             "test",
		ImagePullSecrets: `a,b`,
	}
	assert.Len(t, m.GetImagePullSecrets(), 2)
	for _, secret := range m.GetImagePullSecrets() {
		t.Log(secret.Name)
		if !(secret.Name == "a" || secret.Name == "b") {
			assert.True(t, false)
		}
	}
}

func TestNamespace_ImagePullSecretsArray(t *testing.T) {
	m := Namespace{
		ID:               1,
		Name:             "test",
		ImagePullSecrets: `a,b`,
	}
	assert.Equal(t, []string{"a", "b"}, m.ImagePullSecretsArray())
}

func TestNamespace_ProtoTransform(t *testing.T) {
	m := &Namespace{
		ID:               1,
		Name:             "ns",
		ImagePullSecrets: "a,b",
		CreatedAt:        time.Now().Add(15 * time.Minute),
		UpdatedAt:        time.Now().Add(30 * time.Minute),
		DeletedAt: gorm.DeletedAt{
			Time:  time.Now().Add(-10 * time.Second),
			Valid: true,
		},
		Projects: nil,
	}
	assert.Equal(t, &types.NamespaceModel{
		Id:               int64(m.ID),
		Name:             m.Name,
		ImagePullSecrets: []*types.ImagePullSecret{{Name: "a"}, {Name: "b"}},
		Projects:         nil,
		CreatedAt:        date.ToRFC3339DatetimeString(&m.CreatedAt),
		UpdatedAt:        date.ToRFC3339DatetimeString(&m.UpdatedAt),
		DeletedAt:        date.ToRFC3339DatetimeString(&m.DeletedAt.Time),
	}, m.ProtoTransform())
}
