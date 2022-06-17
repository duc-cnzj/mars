package database

import (
	"fmt"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/go-gormigrate/gormigrate/v2"
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
	gm := gormigrate.New(m.db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "1970-01-01",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(dst...)
			},
		},
		{
			ID: "2022-01-01",
			Migrate: func(tx *gorm.DB) error {
				var migrateV3ToV4 = "migrating: v3 => v4"

				// v3 => v4 start.
				if tx.Migrator().HasColumn(&models.Changelog{}, "gitlab_project_id") {
					// changelogs rename gitlab_project_id => git_project_id
					if err := tx.Migrator().RenameColumn(&models.Changelog{}, "gitlab_project_id", "git_project_id"); err != nil {
						return fmt.Errorf("[%s]: err: %v", migrateV3ToV4, err)
					}
				}

				if tx.Migrator().HasTable("gitlab_projects") {
					// GitlabProject rename GitProject
					if err := tx.Migrator().RenameTable("gitlab_projects", &models.GitProject{}); err != nil {
						return fmt.Errorf("[%s]: err: %v", migrateV3ToV4, err)
					}
					if tx.Migrator().HasColumn(&models.GitProject{}, "gitlab_project_id") {
						// gitlab_project_id => git_project_id
						if err := tx.Migrator().RenameColumn(&models.GitProject{}, "gitlab_project_id", "git_project_id"); err != nil {
							return fmt.Errorf("[%s]: err: %v", migrateV3ToV4, err)
						}
					}
				}

				// projects
				// gitlab_project_id git_project_id
				if tx.Migrator().HasColumn(&models.Project{}, "gitlab_project_id") {
					if err := tx.Migrator().RenameColumn(&models.Project{}, "gitlab_project_id", "git_project_id"); err != nil {
						return fmt.Errorf("[%s]: err: %v", migrateV3ToV4, err)
					}
				}
				// gitlab_branch git_branch
				if tx.Migrator().HasColumn(&models.Project{}, "gitlab_branch") {
					if err := tx.Migrator().RenameColumn(&models.Project{}, "gitlab_branch", "git_branch"); err != nil {
						return fmt.Errorf("[%s]: err: %v", migrateV3ToV4, err)
					}
				}
				// gitlab_commit git_commit
				if tx.Migrator().HasColumn(&models.Project{}, "gitlab_commit") {
					if err := tx.Migrator().RenameColumn(&models.Project{}, "gitlab_commit", "git_commit"); err != nil {
						return fmt.Errorf("[%s]: err: %v", migrateV3ToV4, err)
					}
				}
				// v3 => v4 end.

				if tx.Migrator().HasTable("commands") {
					if err := tx.Migrator().DropTable("commands"); err != nil {
						return fmt.Errorf("[DropTable 'commands']: err: %v", err)
					}
				}
				return nil
			},
		},
		{
			ID: "2022-05-31-global_config-text-longtext",
			Migrate: func(tx *gorm.DB) error {
				if tx.Migrator().HasTable(&models.GitProject{}) {
					types, _ := tx.Migrator().ColumnTypes(&models.GitProject{})
					for _, columnType := range types {
						if columnType.Name() == "global_config" && columnType.DatabaseTypeName() == "text" {
							if err := tx.Migrator().AlterColumn(&models.GitProject{}, "GlobalConfig"); err != nil {
								return err
							}
							break
						}
					}
				}
				if tx.Migrator().HasTable(&models.Changelog{}) {
					types, _ := tx.Migrator().ColumnTypes(&models.Changelog{})
					for _, columnType := range types {
						if columnType.Name() == "manifest" && columnType.DatabaseTypeName() == "text" {
							if err := tx.Migrator().AlterColumn(&models.Changelog{}, "Manifest"); err != nil {
								return err
							}
							break
						}
					}
				}
				return nil
			},
		},
		{
			ID: "2022-05-31-manifest",
			Migrate: func(tx *gorm.DB) error {
				if !tx.Migrator().HasColumn(&models.Project{}, "Manifest") {
					if err := tx.Migrator().AddColumn(&models.Project{}, "Manifest"); err != nil {
						return err
					}
				}
				var projects []models.Project
				tx.Find(&projects)
				for _, project := range projects {
					if project.Manifest == "" {
						var changelog models.Changelog
						if tx.Where("`project_id` = ?", project.ID).Last(&changelog).Error == nil {
							tx.Model(&project).UpdateColumn("manifest", changelog.Manifest)
						}
					}
				}

				return nil
			},
		},
		{
			ID: "2022-07-17-changelogs-add-more-columns",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(&models.Changelog{}); err != nil {
					return fmt.Errorf("[%s]: err: %v", "2022-07-17-changelogs-add-more-columns", err)
				}
				return nil
			},
		},
		{
			ID: "2022-07-17-changelogs-version-tinyint-to-int",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.Migrator().AlterColumn(&models.Changelog{}, "Version"); err != nil {
					return fmt.Errorf("[%s]: err: %v", "2022-07-17-changelogs-version-tinyint-to-int", err)
				}
				return nil
			},
		},
	})

	if err := gm.Migrate(); err != nil {
		mlog.Error(err)
		return err
	}

	return nil
}
