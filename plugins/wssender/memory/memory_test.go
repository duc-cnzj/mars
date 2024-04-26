package memory

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v4/plugins/wssender"

	"github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func Test_memoryPubSub_Close(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns:   map[string]*Conn{},
		roomIDs: map[int64]map[int64]map[string][]labels.Selector{},
		idRooms: map[string]map[int64]struct{}{},
	}
	ps := &memoryPubSub{
		manager: ms,
		uid:     "1",
		id:      "1",
	}
	assert.Nil(t, ps.Close())
	ms.Add(ps.uid, ps.id)
	assert.Len(t, ms.conns, 1)
	assert.Nil(t, ps.Close())
	assert.Len(t, ms.conns, 0)
}

func Test_memoryPubSub_ID(t *testing.T) {
	t.Parallel()
	ps := &memoryPubSub{id: "a"}
	assert.Equal(t, "a", ps.ID())
}

func Test_memoryPubSub_Info(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns:   map[string]*Conn{},
		roomIDs: map[int64]map[int64]map[string][]labels.Selector{},
		idRooms: map[string]map[int64]struct{}{},
	}
	ps := &memoryPubSub{manager: ms}
	assert.Equal(t, ps.manager.conns, ps.Info())
}

func Test_memoryPubSub_Uid(t *testing.T) {
	t.Parallel()
	ps := &memoryPubSub{uid: "uid"}
	assert.Equal(t, "uid", ps.Uid())
}

func Test_memorySender_Add(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns:   map[string]*Conn{},
		roomIDs: map[int64]map[int64]map[string][]labels.Selector{},
		idRooms: map[string]map[int64]struct{}{},
	}
	ps := &memoryPubSub{
		manager: ms,
		uid:     "1",
		id:      "2",
	}
	assert.Nil(t, ps.Close())
	ms.Add(ps.uid, ps.id)
	assert.NotNil(t, ms.conns["2"])
	assert.Equal(t, wssender.MessageChSize, cap(ms.conns["2"].ch))
	assert.Equal(t, "1", ms.conns["2"].uid)
	assert.Equal(t, "2", ms.conns["2"].id)

	assert.Len(t, ms.conns, 1)
	ms.Add("", "")
	assert.Len(t, ms.conns, 1)
}

func Test_memorySender_Delete(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns:   map[string]*Conn{"a": nil},
		roomIDs: map[int64]map[int64]map[string][]labels.Selector{},
		idRooms: map[string]map[int64]struct{}{},
	}
	ms.Delete("xxx", "a")
	assert.Len(t, ms.conns, 0)
}

func Test_memorySender_Destroy(t *testing.T) {
	t.Parallel()
	ms := &memorySender{}
	assert.Nil(t, ms.Destroy())
}

func Test_memorySender_Initialize(t *testing.T) {
	t.Parallel()
	ms := &memorySender{}
	assert.Nil(t, ms.Initialize(map[string]any{}))
	assert.NotNil(t, ms.conns)
	assert.NotNil(t, ms.roomIDs)
	assert.NotNil(t, ms.idRooms)
}

func Test_memoryPubSub_Join(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	ms := &memorySender{
		conns:   map[string]*Conn{},
		roomIDs: map[int64]map[int64]map[string][]labels.Selector{},
		idRooms: map[string]map[int64]struct{}{},
	}
	ps := &memoryPubSub{
		manager: ms,
		uid:     "1",
		id:      "2",
	}
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	ns := &models.Namespace{
		Name: "ns",
	}
	db.Create(ns)
	pmodel := &models.Project{
		Name:         "app",
		PodSelectors: "name=app,age=17|name=bpp,age=18",
		NamespaceId:  ns.ID,
	}
	assert.Nil(t, db.Create(pmodel).Error)
	assert.Nil(t, ps.Join(int64(pmodel.ID)))
	assert.Nil(t, ps.Join(int64(pmodel.ID)))

	var selectors []labels.Selector
	for _, s := range pmodel.GetPodSelectors() {
		parse, err := labels.Parse(s)
		assert.Nil(t, err)
		selectors = append(selectors, parse)
	}
	assert.Equal(t, ms.roomIDs[int64(pmodel.NamespaceId)][int64(pmodel.ID)]["2"], selectors)
	_, ok := ms.idRooms["2"][int64(pmodel.NamespaceId)]
	assert.True(t, ok)
}

func Test_memoryPubSub_Leave(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns: map[string]*Conn{},
		roomIDs: map[int64]map[int64]map[string][]labels.Selector{
			1: {
				2: map[string][]labels.Selector{
					"2": nil,
				},
			},
		},
		idRooms: map[string]map[int64]struct{}{
			"2": {
				1: {},
			},
		},
	}
	ps := &memoryPubSub{
		manager: ms,
		uid:     "1",
		id:      "2",
	}
	assert.Nil(t, ps.Leave(1, 2))
	assert.Len(t, ms.roomIDs[1][2], 0)
	assert.Len(t, ms.idRooms["2"], 0)
}

func Test_memoryPubSub_Publish(t *testing.T) {
	t.Parallel()
	parse, _ := labels.Parse("a=b")
	parse2, _ := labels.Parse("a=b,age=18")
	ms := &memorySender{
		conns: map[string]*Conn{},
		roomIDs: map[int64]map[int64]map[string][]labels.Selector{
			1: {
				2: map[string][]labels.Selector{
					"2": {parse, parse2},
				},
			},
		},
		idRooms: map[string]map[int64]struct{}{
			"2": {
				1: {},
			},
		},
	}
	ps := &memoryPubSub{
		manager: ms,
		uid:     "1",
		id:      "2",
	}
	ms.Add("1", "2")
	ps.Publish(1, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "ns",
			Labels: labels.Set{
				"a": "b",
			},
		},
	})
	ps.Publish(1, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "ns",
			Labels: labels.Set{
				"a": "c",
			},
		},
	})
	ps.Publish(1, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "ns",
			Labels: labels.Set{
				"age": "18",
			},
		},
	})
	ps.Publish(1, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "ns",
			Labels: labels.Set{
				"a":   "b",
				"age": "18",
			},
		},
	})
	assert.Len(t, ms.conns["2"].ch, 2)
}

func Test_memoryPubSub_Run(t *testing.T) {
	t.Parallel()
	ps := &memoryPubSub{}
	ctx, fn := context.WithCancel(context.TODO())
	defer fn()
	assert.Nil(t, ps.Run(ctx))
}

func Test_memoryPubSub_Subscribe(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns: map[string]*Conn{},
	}
	ps := &memoryPubSub{
		manager: ms,
		id:      "2",
	}
	ms.Add("1", "2")
	dn := func() <-chan []byte {
		return ms.conns["2"].ch
	}
	assert.Equal(t, ps.Subscribe(), dn())
}

func Test_memoryPubSub_ToAll(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns: map[string]*Conn{},
	}
	ps := &memoryPubSub{
		manager: ms,
	}
	ms.Add("xx", "1")
	ms.Add("xxx", "2")
	ps.ToAll(&websocket.WsMetadataResponse{})
	assert.Len(t, ms.conns["1"].ch, 1)
	assert.Len(t, ms.conns["2"].ch, 1)
}

func Test_memoryPubSub_ToOthers(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns: map[string]*Conn{},
	}
	ps := &memoryPubSub{
		id:      "1",
		uid:     "1",
		manager: ms,
	}
	ms.Add("1", "1")
	ms.Add("2", "2")
	ps.ToOthers(&websocket.WsMetadataResponse{})
	assert.Len(t, ms.conns["1"].ch, 0)
	assert.Len(t, ms.conns["2"].ch, 1)
}

func Test_memoryPubSub_ToSelf(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns: map[string]*Conn{},
	}
	ps := &memoryPubSub{
		id:      "1",
		uid:     "1",
		manager: ms,
	}
	ms.Add("1", "1")
	ms.Add("2", "2")
	ps.ToSelf(&websocket.WsMetadataResponse{})
	assert.Len(t, ms.conns["1"].ch, 1)
	assert.Len(t, ms.conns["2"].ch, 0)
}

func Test_memorySender_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, memorySenderName, (&memorySender{}).Name())
}

func Test_memorySender_New(t *testing.T) {
	t.Parallel()
	ms := &memorySender{
		conns: map[string]*Conn{},
	}
	sub := ms.New("a", "b")
	assert.NotNil(t, sub)
	assert.Equal(t, 1, len(ms.conns))
	_, ok := ms.conns["b"]
	assert.True(t, ok)
}
