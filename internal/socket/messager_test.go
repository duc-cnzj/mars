package socket

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/types"

	"github.com/duc-cnzj/mars/internal/contracts"

	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageSender(t *testing.T) {
	assert.Implements(t, (*contracts.DeployMsger)(nil), NewMessageSender(&WsConn{}, "aa", websocket.Type_ProcessPercent))
}

func Test_messager_IsStopped(t *testing.T) {
	sender := &messager{conn: nil, slugName: "", wsType: websocket.Type_ProcessPercent}
	assert.False(t, sender.IsStopped())
	sender.Stop(nil)
	assert.True(t, sender.IsStopped())
}

func Test_messager_SendDeployedResult(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	ps := mock.NewMockPubSub(c)
	sender := NewMessageSender(&WsConn{
		pubSub: ps,
	}, "aa", websocket.Type_ApplyProject)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	sender.SendDeployedResult(websocket.ResultType_Deployed, "aa", nil)
	sender.Stop(nil)
	sender.SendDeployedResult(websocket.ResultType_Deployed, "aa", nil)
}

type matcher struct {
	response *websocket.WsMetadataResponse
}

func (m *matcher) Matches(x any) bool {
	response, ok := x.(*websocket.WsMetadataResponse)
	if !ok {
		return false
	}
	return m.response.String() == response.String()

}

func (m *matcher) String() string {
	return ""
}

type containerLogResponseMatcher struct {
	response *websocket.WsWithContainerMessageResponse
}

func (m *containerLogResponseMatcher) Matches(x any) bool {
	response, ok := x.(*websocket.WsWithContainerMessageResponse)
	if !ok {
		return false
	}
	return m.response.String() == response.String()

}

func (m *containerLogResponseMatcher) String() string {
	return ""
}

func Test_messager_SendEndError(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	ps := mock.NewMockPubSub(c)
	sender := NewMessageSender(&WsConn{
		id:     "1",
		uid:    "2",
		pubSub: ps,
	}, "aa", websocket.Type_ApplyProject)
	ps.EXPECT().ToSelf(&matcher{
		response: &websocket.WsMetadataResponse{
			Metadata: &websocket.Metadata{
				Slug:    "aa",
				Type:    websocket.Type_ApplyProject,
				Result:  ResultError,
				End:     true,
				Uid:     "2",
				Id:      "1",
				Message: "err",
			},
		},
	})
	sender.SendEndError(errors.New("err"))
}

func Test_messager_SendError(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	ps := mock.NewMockPubSub(c)
	sender := NewMessageSender(&WsConn{
		id:     "1",
		uid:    "2",
		pubSub: ps,
	}, "aa", websocket.Type_ApplyProject)
	ps.EXPECT().ToSelf(&matcher{
		response: &websocket.WsMetadataResponse{
			Metadata: &websocket.Metadata{
				Slug:    "aa",
				Type:    websocket.Type_ApplyProject,
				Result:  ResultError,
				End:     false,
				Uid:     "2",
				Id:      "1",
				Message: "err",
			},
		},
	})
	sender.SendError(errors.New("err"))
}

func Test_messager_SendMsg(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	ps := mock.NewMockPubSub(c)
	sender := NewMessageSender(&WsConn{
		id:     "1",
		uid:    "2",
		pubSub: ps,
	}, "aa", websocket.Type_ApplyProject)
	ps.EXPECT().ToSelf(&matcher{
		response: &websocket.WsMetadataResponse{
			Metadata: &websocket.Metadata{
				Slug:    "aa",
				Type:    websocket.Type_ApplyProject,
				Result:  ResultSuccess,
				End:     false,
				Uid:     "2",
				Id:      "1",
				Message: "aaa",
			},
		},
	})
	sender.SendMsg("aaa")
}

func Test_messager_SendProcessPercent(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	ps := mock.NewMockPubSub(c)
	sender := NewMessageSender(&WsConn{
		id:     "1",
		uid:    "2",
		pubSub: ps,
	}, "aa", websocket.Type_ApplyProject)
	ps.EXPECT().ToSelf(&matcher{
		response: &websocket.WsMetadataResponse{
			Metadata: &websocket.Metadata{
				Slug:    "aa",
				Type:    websocket.Type_ProcessPercent,
				Result:  ResultSuccess,
				End:     false,
				Uid:     "2",
				Id:      "1",
				Percent: 10,
			},
		},
	})
	sender.SendProcessPercent(10)
}

func Test_messager_SendProtoMsg(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	ps := mock.NewMockPubSub(c)
	sender := NewMessageSender(&WsConn{
		id:     "1",
		uid:    "2",
		pubSub: ps,
	}, "aa", websocket.Type_ApplyProject)
	pmsg := &websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Slug:    "aa",
			Type:    websocket.Type_ApplyProject,
			Result:  ResultSuccess,
			End:     false,
			Uid:     "2",
			Id:      "1",
			Message: "aaa",
		},
	}
	ps.EXPECT().ToSelf(&matcher{
		response: pmsg,
	})
	sender.SendProtoMsg(pmsg)
}

func Test_messager_SendMsgWithContainerLog(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	ps := mock.NewMockPubSub(c)
	sender := NewMessageSender(&WsConn{
		id:     "1",
		uid:    "2",
		pubSub: ps,
	}, "aa", websocket.Type_ApplyProject)
	ps.EXPECT().ToSelf(&containerLogResponseMatcher{
		response: &websocket.WsWithContainerMessageResponse{
			Metadata: &websocket.Metadata{
				Slug:    "aa",
				Type:    websocket.Type_ApplyProject,
				Result:  ResultLogWithContainers,
				End:     false,
				Uid:     "2",
				Id:      "1",
				Message: "aaa",
			},
			Containers: []*types.Container{
				{
					Namespace: "a",
					Pod:       "b",
					Container: "c",
				},
			},
		},
	})
	sender.SendMsgWithContainerLog("aaa", []*types.Container{
		{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
	})
}
