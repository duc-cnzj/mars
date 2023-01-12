package database

import (
	"fmt"

	"github.com/duc-cnzj/mars/internal/contracts"
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
			//这里错了，实际上是 06-17
			ID: "2022-07-17-changelogs-add-more-columns",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(&models.Changelog{}); err != nil {
					return fmt.Errorf("[%s]: err: %v", "2022-07-17-changelogs-add-more-columns", err)
				}
				return nil
			},
		},
		{
			//这里错了，实际上是 06-17
			ID: "2022-07-17-changelogs-version-tinyint-to-int",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.Migrator().AlterColumn(&models.Changelog{}, "Version"); err != nil {
					return fmt.Errorf("[%s]: err: %v", "2022-07-17-changelogs-version-tinyint-to-int", err)
				}
				return nil
			},
		},
		{
			ID: "2022-07-12-add-action-index-to-events-table",
			Migrate: func(tx *gorm.DB) error {
				if !tx.Migrator().HasIndex(&models.Event{}, "Action") {
					if err := tx.Migrator().CreateIndex(&models.Event{}, "Action"); err != nil {
						return fmt.Errorf("[%s]: err: %v", "2022-07-12-add-action-index-to-events-table", err)
					}
				}
				return nil
			},
		},
		{
			ID: "2022-08-17-add-cache_locks-table",
			Migrate: func(tx *gorm.DB) error {
				tx.AutoMigrate(&models.CacheLock{})
				return nil
			},
		},
		{
			ID: "2022-08-18-rebuild-table-db-cache",
			Migrate: func(tx *gorm.DB) error {
				if tx.Migrator().HasColumn(&models.DBCache{}, "id") {
					tx.Migrator().RenameTable("db_cache", "db_cache_old")
				}
				tx.Migrator().AutoMigrate(&models.DBCache{})
				return nil
			},
		},
		{
			ID: "2022-08-19-drop-table-db_cache_old",
			Migrate: func(tx *gorm.DB) error {
				if tx.Migrator().HasTable("db_cache_old") {
					if err := tx.Migrator().DropTable("db_cache_old"); err != nil {
						return fmt.Errorf("[%s]: err: %v", "2022-08-19-drop-table-db_cache_old", err)
					}
				}
				return nil
			},
		},
		{
			ID: "2022-08-23-events-new-old-table-to-longtext",
			Migrate: func(tx *gorm.DB) error {
				var fields = []string{"Old", "New"}
				for _, field := range fields {
					if err := tx.Migrator().AlterColumn(&models.Event{}, field); err != nil {
						return fmt.Errorf("[%s]: err: %v", "2022-08-23-events-new-old-table-to-longtext", err)
					}
				}
				// sqlite 在 migrate 会丢失索引，所以这里需要重新添加
				if !tx.Migrator().HasIndex(&models.Event{}, "Action") {
					tx.Migrator().CreateIndex(&models.Event{}, "Action")
				}
				return nil
			},
		},
		{
			ID: "2022-08-29-add-upload_type-to-files-table",
			Migrate: func(tx *gorm.DB) error {
				if !tx.Migrator().HasColumn(&models.File{}, "UploadType") {
					if err := tx.Migrator().AddColumn(&models.File{}, "UploadType"); err != nil {
						return fmt.Errorf("[%s]: err: %v", "2022-08-29-add-upload_type-to-files-table", err)
					}
				}
				return nil
			},
		},
		{
			ID: "2022-11-10-projects-extra_values-and-final_extra_values-longtext",
			Migrate: func(tx *gorm.DB) error {
				if tx.Migrator().HasTable(&models.Project{}) {
					types, _ := tx.Migrator().ColumnTypes(&models.Project{})
					for _, columnType := range types {
						if columnType.Name() == "final_extra_values" && columnType.DatabaseTypeName() == "text" {
							if err := tx.Migrator().AlterColumn(&models.Project{}, "FinalExtraValues"); err != nil {
								return err
							}
						}
						if columnType.Name() == "extra_values" && columnType.DatabaseTypeName() == "text" {
							if err := tx.Migrator().AlterColumn(&models.Project{}, "ExtraValues"); err != nil {
								return err
							}
						}
					}
				}

				return nil
			},
		},
		{
			ID: "2022-12-01-create-access_tokens-table",
			Migrate: func(tx *gorm.DB) error {
				if !tx.Migrator().HasTable(&models.AccessToken{}) {
					if err := tx.Migrator().CreateTable(&models.AccessToken{}); err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "2022-12-16-add-deleted_at-index-to-namespaces-table",
			Migrate: func(tx *gorm.DB) error {
				if !tx.Migrator().HasIndex(&models.Namespace{}, "DeletedAt") {
					if err := tx.Migrator().CreateIndex(&models.Namespace{}, "DeletedAt"); err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "2022-12-16-add-idx_namespace_id_deleted_at-to-projects-table",
			Migrate: func(tx *gorm.DB) error {
				if !tx.Migrator().HasIndex(&models.Project{}, "idx_namespace_id_deleted_at") {
					if err := tx.Migrator().CreateIndex(&models.Project{}, "idx_namespace_id_deleted_at"); err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "2022-12-16-del-idx_version_projectid_deleted_at_config_changed-to-changelogs-table",
			Migrate: func(tx *gorm.DB) error {
				if tx.Migrator().HasIndex(&models.Changelog{}, "idx_version_projectid_deleted_at_config_changed") {
					if err := tx.Migrator().DropIndex(&models.Changelog{}, "idx_version_projectid_deleted_at_config_changed"); err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "2022-12-16-add-idx_projectid_config_changed_deleted_at_version-to-changelogs-table",
			Migrate: func(tx *gorm.DB) error {
				if !tx.Migrator().HasIndex(&models.Changelog{}, "idx_projectid_config_changed_deleted_at_version") {
					if err := tx.Migrator().CreateIndex(&models.Changelog{}, "idx_projectid_config_changed_deleted_at_version"); err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			ID: "2023-01-09-add-version-to-projects-table",
			Migrate: func(tx *gorm.DB) error {
				if !tx.Migrator().HasColumn(&models.Project{}, "version") {
					if err := tx.Migrator().AddColumn(&models.Project{}, "Version"); err != nil {
						return err
					}
				}

				return nil
			},
		},
	})

	return gm.Migrate()
}
