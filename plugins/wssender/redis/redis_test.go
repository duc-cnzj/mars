package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/duc-cnzj/mars/v4/plugins/wssender"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	rdb  *redis.Client
	skip = false
)

var (
	host string = os.Getenv("REDIS_HOST")
	port string = os.Getenv("REDIS_PORT")
	pwd  string = os.Getenv("REDIS_PWD")
	db   string = os.Getenv("REDIS_DB")
	addr string
)

func TestMain(t *testing.M) {
	setDefault := func(key *string, value string) {
		if *key == "" {
			*key = value
		}
	}
	setDefault(&host, "127.0.0.1")
	setDefault(&port, "6379")
	setDefault(&pwd, "")
	setDefault(&db, "1")
	rdbNum := 1
	atoi, err := strconv.Atoi(db)
	if err == nil {
		rdbNum = atoi
	}
	addr = fmt.Sprintf("%v:%v", host, port)
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       rdbNum,
	})
	if err := rdb.Ping(context.TODO()).Err(); err != nil {
		skip = true
	}
	code := t.Run()
	rdb.Close()
	os.Exit(code)
}

func NewRdb(db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})
}

func Test_getRedisProjectEventRoom(t *testing.T) {
	if skip {
		t.Skip()
	}
	t.Parallel()
	assert.Equal(t, "project-pod-events:1", getRedisProjectEventRoom(1))
}

func Test_podEventManagers_Join(t *testing.T) {
	if skip {
		t.Skip()
	}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	ps := &podEventManagers{
		id:           "id",
		uid:          "uid",
		rds:          rdb,
		pubSub:       rdb.Subscribe(context.TODO()),
		channels:     map[string]struct{}{},
		pidSelectors: map[int64][]labels.Selector{},
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
	assert.Error(t, ps.Join(int64(999999)))

	var selectors []labels.Selector
	for _, s := range pmodel.GetPodSelectors() {
		parse, err := labels.Parse(s)
		assert.Nil(t, err)
		selectors = append(selectors, parse)
	}

	_, ok := ps.channels[getRedisProjectEventRoom(ns.ID)]
	assert.True(t, ok)
	ls := ps.pidSelectors[int64(pmodel.ID)]
	assert.Equal(t, selectors, ls)
	assert.Nil(t, ps.pubSub.Close())
}

func Test_podEventManagers_Leave(t *testing.T) {
	if skip {
		t.Skip()
	}
	pem := &podEventManagers{
		id:     "id",
		uid:    "uid",
		rds:    rdb,
		pubSub: rdb.Subscribe(context.TODO()),
		channels: map[string]struct{}{
			getRedisProjectEventRoom(1): {},
		},
		pidSelectors: map[int64][]labels.Selector{
			1: nil,
		},
	}
	pem.Leave(1, 1)
	assert.Len(t, pem.channels, 0)
	assert.Len(t, pem.pidSelectors, 0)
	pem.pubSub.Close()
}

func Test_podEventManagers_Publish(t *testing.T) {
	if skip {
		t.Skip()
	}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	rs := &redisSender{
		rds: rdb,
	}
	cancel, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()
	wg := sync.WaitGroup{}
	wg.Add(2)

	ps := rs.New("uid", "id")
	go func() {
		defer wg.Done()
		ps.Run(cancel)
	}()

	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	ns := &models.Namespace{
		Name: "ns",
	}
	db.Create(ns)
	pmodel := &models.Project{
		Name:         "app",
		PodSelectors: "name=app",
		NamespaceId:  ns.ID,
	}
	assert.Nil(t, db.Create(pmodel).Error)
	ps.Join(int64(pmodel.ID))

	ch := ps.Subscribe()
	go func() {
		defer wg.Done()
		<-ch
		cancelFunc()
		assert.True(t, true)
	}()

	ps.Publish(int64(ns.ID), &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"name": "app",
			},
		},
	})
	wg.Wait()
	ps.Close()
}

func Test_podEventManagers_Run(t *testing.T) {
	if skip {
		t.Skip()
	}
	pem := &podEventManagers{
		pubSub: rdb.Subscribe(context.TODO()),
	}
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Equal(t, "context canceled", pem.Run(cancel).Error())

	pem2 := &podEventManagers{
		pubSub: rdb.Subscribe(context.TODO()),
	}
	cancel2, cancelFunc2 := context.WithCancel(context.TODO())
	defer cancelFunc2()
	assert.Nil(t, pem2.pubSub.Close())
	assert.Equal(t, "podEventManagers ch closed", pem2.Run(cancel2).Error())
}

func Test_rdsPubSub_Subscribe(t *testing.T) {
	if skip {
		t.Skip()
	}
	cancel, cancelFunc := context.WithCancel(context.TODO())

	r := rdsPubSub{
		wsPubSub: rdb.Subscribe(context.TODO()),
		done:     cancel,
		doneFunc: cancelFunc,
	}
	r.Subscribe()
	channel := r.wsPubSub.Channel()
	assert.Nil(t, r.wsPubSub.Close())
	_, ok := <-channel
	assert.False(t, ok)
	_, ok = <-r.done.Done()
	assert.False(t, ok)
}

func Test_rdsPubSub_Close(t *testing.T) {
	if skip {
		t.Skip()
	}
	cancel, cancelFunc := context.WithCancel(context.TODO())

	r := rdsPubSub{
		done:     cancel,
		doneFunc: cancelFunc,
	}
	assert.Nil(t, r.Close())
	_, ok := <-r.done.Done()
	assert.False(t, ok)
}

func Test_rdsPubSub_ID(t *testing.T) {
	if skip {
		t.Skip()
	}
	assert.Equal(t, "id", (&rdsPubSub{id: "id"}).ID())
}

func Test_rdsPubSub_Info(t *testing.T) {
	if skip {
		t.Skip()
	}
	assert.Equal(t, "<unknown>", (&rdsPubSub{}).Info())
}

func Test_rdsPubSub_ToAll(t *testing.T) {
	if skip {
		t.Skip()
	}
	rs := &redisSender{rds: NewRdb(10)}
	defer rs.rds.Close()
	ps1 := rs.New("bbb", "b-1")
	ps2 := rs.New("bbb", "b-2")

	ch1 := ps1.Subscribe()
	ch2 := ps2.Subscribe()

	msg := &websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Id:  "b-1",
			Uid: "bbb",
			To:  plugins.ToAll,
		},
	}

	marshal := wssender.TransformToResponse(msg)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.Nil(t, ps1.ToAll(msg))
	}()
	wg.Wait()
	res := <-ch1
	assert.True(t, true)
	assert.Equal(t, marshal, res)
	res2 := <-ch2
	assert.True(t, true)
	assert.Equal(t, marshal, res2)
	ps1.Close()
	ps2.Close()
}

func Test_rdsPubSub_ToOthers(t *testing.T) {
	if skip {
		t.Skip()
	}
	rs := &redisSender{rds: NewRdb(2)}
	defer rs.rds.Close()
	ps1 := rs.New("aaa", "a-1")
	ps2 := rs.New("aaa", "a-2")
	ch := ps2.Subscribe()

	msg := &websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			To: plugins.ToOthers,
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.Nil(t, ps1.ToOthers(msg))
	}()
	wg.Wait()
	res := <-ch
	assert.True(t, true)
	assert.Equal(t, wssender.TransformToResponse(msg), res)
	ps1.Close()
	ps2.Close()
}

func Test_rdsPubSub_ToSelf(t *testing.T) {
	if skip {
		t.Skip()
	}
	rs := &redisSender{rds: NewRdb(3)}
	defer rs.rds.Close()
	ps1 := rs.New("aaa", "a-1")
	ch := ps1.Subscribe()

	msg := &websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			To: plugins.ToSelf,
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.Nil(t, ps1.ToSelf(msg))
	}()
	wg.Wait()
	res2 := <-ch
	assert.True(t, true)
	assert.Equal(t, wssender.TransformToResponse(msg), res2)
	ps1.Close()
}

func Test_rdsPubSub_Uid(t *testing.T) {
	if skip {
		t.Skip()
	}
	t.Parallel()
	ps := &rdsPubSub{
		uid: "uid",
	}
	assert.Equal(t, "uid", ps.Uid())
}

func Test_redisSender_Destroy(t *testing.T) {
	if skip {
		t.Skip()
	}
	t.Parallel()
	rs := &redisSender{}
	err := rs.Initialize(map[string]any{
		"addr":     addr,
		"password": pwd,
		"db":       2,
	})
	assert.Nil(t, err)
	assert.Nil(t, rs.Destroy())
	assert.Error(t, rs.rds.Close())
}

func Test_redisSender_Initialize(t *testing.T) {
	if skip {
		t.Skip()
	}
	t.Parallel()
	rs := &redisSender{}
	err := rs.Initialize(map[string]any{
		"addr":     addr,
		"password": pwd,
		"db":       2,
	})
	assert.Nil(t, err)
	assert.Nil(t, rs.rds.Close())

	rs2 := &redisSender{}
	err = rs2.Initialize(map[string]any{
		"addr":     addr,
		"password": pwd,
	})
	assert.Error(t, err)

	rs3 := &redisSender{}
	err = rs3.Initialize(map[string]any{
		"addr":     "xxx:6379",
		"password": pwd,
		"db":       2,
	})
	assert.Error(t, err)
}

func Test_redisSender_Name(t *testing.T) {
	if skip {
		t.Skip()
	}
	t.Parallel()
	assert.Equal(t, redisSenderName, (&redisSender{}).Name())
}

func Test_redisSender_New(t *testing.T) {
	if skip {
		t.Skip()
	}
	t.Parallel()
	ps := (&redisSender{
		rds: rdb,
	}).New("uid", "id")
	pp := ps.(*rdsPubSub)
	assert.NotNil(t, pp.done)
	assert.NotNil(t, pp.doneFunc)
	assert.NotNil(t, pp.ch)
	assert.Same(t, rdb, pp.rds)
	assert.Equal(t, "uid", pp.uid)
	assert.Equal(t, "id", pp.id)
	assert.NotNil(t, pp.wsPubSub)
	assert.NotNil(t, pp.ProjectPodEventPublisher)
	assert.NotNil(t, pp.ProjectPodEventSubscriber)
	assert.Nil(t, pp.wsPubSub.Close())
	assert.Nil(t, pp.ProjectPodEventSubscriber.(*podEventManagers).pubSub.Close())
}
