package repo

import (
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type DomainRepo interface {
	GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string
}

type domainRepo struct {
	logger mlog.Logger
	dm     application.DomainManager
}

func (d *domainRepo) GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string {
	return d.dm.GetDomainByIndex(projectName, namespace, index, preOccupiedLen)
}

var _ DomainRepo = (*domainRepo)(nil)

func NewDomainRepo(logger mlog.Logger, pl application.PluginManger) DomainRepo {
	return &domainRepo{logger: logger, dm: pl.Domain()}
}
