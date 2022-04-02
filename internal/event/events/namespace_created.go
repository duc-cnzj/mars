package events

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
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
			err := utils.AddTlsSecret(ns, name, key, crt)
			if err != nil {
				mlog.Error(err)
			}
		}
	}
	return nil
}
