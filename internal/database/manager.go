package database

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"gorm.io/gorm"
)

type Manager struct {
	app contracts.ApplicationInterface
	db  *gorm.DB
}

func NewManager(app contracts.ApplicationInterface) *Manager {
	return &Manager{app: app}
}

func (m *Manager) DB() *gorm.DB {
	return m.db
}

func (m *Manager) SetDB(db *gorm.DB) {
	m.db = db
}

func (m *Manager) AutoMigrate(dst ...any) error {
	var migrateV3ToV4 = "migrating: v3 => v4 start."

	// v3 => v4 start.
	if m.db.Migrator().HasColumn(&models.Changelog{}, "gitlab_project_id") {
		// changelogs rename gitlab_project_id => git_project_id
		if err := m.db.Migrator().RenameColumn(&models.Changelog{}, "gitlab_project_id", "git_project_id"); err != nil {
			mlog.Warningf("[%s]: err: %v", migrateV3ToV4, err)
		}
	}

	if m.db.Migrator().HasTable("gitlab_projects") {
		// GitlabProject rename GitProject
		if err := m.db.Migrator().RenameTable("gitlab_projects", &models.GitProject{}); err != nil {
			mlog.Warningf("[%s]: err: %v", migrateV3ToV4, err)
		}
		if m.db.Migrator().HasColumn(&models.GitProject{}, "gitlab_project_id") {
			// gitlab_project_id => git_project_id
			if err := m.db.Migrator().RenameColumn(&models.GitProject{}, "gitlab_project_id", "git_project_id"); err != nil {
				mlog.Warningf("[%s]: err: %v", migrateV3ToV4, err)
			}
		}
	}

	// projects
	// gitlab_project_id git_project_id
	if m.db.Migrator().HasColumn(&models.Project{}, "gitlab_project_id") {
		if err := m.db.Migrator().RenameColumn(&models.Project{}, "gitlab_project_id", "git_project_id"); err != nil {
			mlog.Warningf("[%s]: err: %v", migrateV3ToV4, err)
		}
	}
	// gitlab_branch git_branch
	if m.db.Migrator().HasColumn(&models.Project{}, "gitlab_branch") {
		if err := m.db.Migrator().RenameColumn(&models.Project{}, "gitlab_branch", "git_branch"); err != nil {
			mlog.Warningf("[%s]: err: %v", migrateV3ToV4, err)
		}
	}
	// gitlab_commit git_commit
	if m.db.Migrator().HasColumn(&models.Project{}, "gitlab_commit") {
		if err := m.db.Migrator().RenameColumn(&models.Project{}, "gitlab_commit", "git_commit"); err != nil {
			mlog.Warningf("[%s]: err: %v", migrateV3ToV4, err)
		}
	}
	// v3 => v4 end.

	if err := m.db.AutoMigrate(dst...); err != nil {
		return err
	}

	return nil
}
