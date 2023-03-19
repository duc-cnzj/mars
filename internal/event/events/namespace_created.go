package events

import (
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/utils/tls"
	v1 "k8s.io/api/core/v1"
)

const EventNamespaceCreated contracts.Event = "namespace_created"

type NamespaceCreatedData struct {
	NsModel  *models.Namespace
	NsK8sObj *v1.Namespace
}

func init() {
	Register(EventNamespaceCreated, HandleInjectTlsSecret)
}

func HandleInjectTlsSecret(data any, e contracts.Event) error {
	if createdData, ok := data.(NamespaceCreatedData); ok {
		name, key, crt := plugins.GetDomainManager().GetCerts()
		if name != "" && key != "" && crt != "" {
			ns := createdData.NsK8sObj.Name
			err := tls.AddTlsSecret(ns, name, key, crt)
			if err != nil {
				mlog.Error(err)
			}
		}
	}
	return nil
}
