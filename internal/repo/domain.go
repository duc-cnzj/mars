package repo

import (
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type DomainRepo interface {
	GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string
}

var _ DomainRepo = (*domainRepo)(nil)

type domainRepo struct {
	logger mlog.Logger
	dm     application.DomainManager
}

func NewDomainRepo(logger mlog.Logger, pl application.PluginManger) DomainRepo {
	return &domainRepo{logger: logger, dm: pl.Domain()}
}

func (d *domainRepo) GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string {
	return d.dm.GetDomainByIndex(projectName, namespace, index, preOccupiedLen)
}
