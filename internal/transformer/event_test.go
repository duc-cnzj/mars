package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/duc-cnzj/mars/v5/internal/util/date"
	"github.com/stretchr/testify/assert"
)

func TestFromEvent_NilInput(t *testing.T) {
	var e *repo.Event
	result := transformer.FromEvent(e)
	assert.Nil(t, result)
}

func TestFromEvent_ValidInput(t *testing.T) {
	now := time.Now()
	e := &repo.Event{
		ID:        1,
		CreatedAt: now,
		UpdatedAt: now,
		Action:    types.EventActionType_Delete,
		Username:  "testUsername",
		Message:   "testMessage",
		Old:       "testOld",
		New:       "testNew",
		Duration:  "testDuration",
		FileID:    nil,
		HasDiff:   true,
		File:      nil,
	}
	result := transformer.FromEvent(e)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, types.EventActionType_Delete, result.Action)
	assert.Equal(t, "testUsername", result.Username)
	assert.Equal(t, "testMessage", result.Message)
	assert.Equal(t, "testOld", result.Old)
	assert.Equal(t, "testNew", result.New)
	assert.Equal(t, "testDuration", result.Duration)
	assert.Equal(t, int32(0), result.FileId)
	assert.Nil(t, result.File)
	assert.Equal(t, true, result.HasDiff)
	assert.Equal(t, date.ToHumanizeDatetimeString(&now), result.EventAt)
	assert.Equal(t, date.ToRFC3339DatetimeString(&now), result.CreatedAt)
	assert.Equal(t, date.ToRFC3339DatetimeString(&now), result.UpdatedAt)
	assert.Empty(t, result.DeletedAt)
}

func TestFromEvent_DeletedEvent(t *testing.T) {
	now := time.Now()
	e := &repo.Event{
		ID:        1,
		Action:    types.EventActionType_Delete,
		Username:  "testUsername",
		Message:   "testMessage",
		Old:       "testOld",
		New:       "testNew",
		Duration:  "testDuration",
		FileID:    nil,
		File:      nil,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &now,
	}
	result := transformer.FromEvent(e)
	assert.NotNil(t, result)
	assert.NotNil(t, result.DeletedAt)
}
