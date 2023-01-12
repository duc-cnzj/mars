package database

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestManager_DB(t *testing.T) {
	manager := NewManager(nil)
	db := gorm.DB{}
	manager.SetDB(&db)
	assert.Same(t, &db, manager.DB())
}

func TestManager_SetDB(t *testing.T) {
	manager := NewManager(nil)
	db := gorm.DB{}
	manager.SetDB(&db)
	assert.Same(t, &db, manager.db)
}

func TestNewManager(t *testing.T) {
	assert.Implements(t, (*contracts.DBManager)(nil), NewManager(nil))
}

type File struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Path     string `json:"path" gorm:"size:255;not null;comment:文件全路径"`
	Size     uint64 `json:"size" gorm:"not null;default:0;comment:文件大小"`
	Username string `json:"username" gorm:"size:255;not null;default:'';comment:用户名称"`

	Namespace     string `json:"namespace" gorm:"size:100;not null;default:'';"`
	Pod           string `json:"pod" gorm:"size:100;not null;default:'';"`
	Container     string `json:"container" gorm:"size:100;not null;default:'';"`
	ContainerPath string `json:"container_path" gorm:"size:255;not null;default:'';comment:容器中的文件路径"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Changelog struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Version       uint8  `json:"version" gorm:"not null;default:1;"`
	Username      string `json:"username" gorm:"size:100;not null;comment:'修改人'"`
	Manifest      string `json:"manifest" gorm:"type:text;"`
	Config        string `json:"config" gorm:"type:text;commit:用户提交的配置"`
	ConfigChanged bool   `json:"config_changed"`

	ProjectID       int `json:"project_id" gorm:"not null;default:0;"`
	GitlabProjectID int `json:"gitlab_project_id" gorm:"not null;default:0;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type GitlabProject struct {
	ID int `json:"id" gorm:"primaryKey;"`

	DefaultBranch   string `json:"default_branch" gorm:"type:varchar(255);not null;default:'';"`
	Name            string `json:"name" gorm:"type:varchar(255);not null;default:'';"`
	GitlabProjectId int    `json:"gitlab_project_id" gorm:"not null;type:integer;"`
	Enabled         bool   `json:"enabled" gorm:"not null;default:false;"`
	GlobalEnabled   bool   `json:"global_enabled" gorm:"not null;default:false;"`
	GlobalConfig    string `json:"global_config" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Project struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Name             string `json:"name" gorm:"size:100;not null;comment:'项目名'"`
	GitlabProjectId  int    `json:"gitlab_project_id" gorm:"not null;type:integer;"`
	GitlabBranch     string `json:"gitlab_branch" gorm:"not null;size:255;"`
	GitlabCommit     string `json:"gitlab_commit" gorm:"not null;size:255;"`
	Config           string `json:"config"`
	OverrideValues   string `json:"override_values"`
	DockerImage      string `json:"docker_image" gorm:"not null;size:255;default:''"`
	PodSelectors     string `json:"pod_selectors" gorm:"type:text;nullable;"`
	NamespaceId      int    `json:"namespace_id"`
	Atomic           bool   `json:"atomic"`
	ExtraValues      string `json:"extra_values" gorm:"type:text;nullable;comment:'用户表单传入的额外值'"`
	FinalExtraValues string `json:"final_extra_values" gorm:"type:text;nullable;comment:'用户表单传入的额外值 + 系统默认的额外值'"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Commands struct {
	ID int `json:"id" gorm:"primaryKey;"`
}

type Event struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Action   uint8  `json:"action" gorm:"type:tinyint;not null;default:0;"`
	Username string `json:"username" gorm:"size:255;not null;default:'';comment:用户名称"`
	Message  string `json:"message" gorm:"size:255;not null;default:'';"`

	Old      string `json:"old" gorm:"type:text;"`
	New      string `json:"new" gorm:"type:text;"`
	Duration string `json:"duration" gorm:"not null;default:''"`

	FileID *int `json:"file_id" gorm:"nullable;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Namespace struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Name             string `json:"name" gorm:"size:100;not null;comment:项目空间名"`
	ImagePullSecrets string `json:"image_pull_secrets" gorm:"size:255;not null;default:'';comment:项目空间拉取镜像的secrets，数组"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Projects []Project
}

func TestManager_AutoMigrate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.Exec("PRAGMA foreign_keys = ON", nil)
	s, _ := db.DB()
	defer s.Close()

	ma := &Manager{db: db}
	assert.Nil(t, db.AutoMigrate(&Changelog{}, &GitlabProject{}, &Project{}, &Commands{}, &Event{}, &File{}, &Namespace{}))
	assert.True(t, db.Migrator().HasColumn(&Changelog{}, "gitlab_project_id"))
	assert.True(t, db.Migrator().HasTable("gitlab_projects"))
	assert.True(t, db.Migrator().HasColumn("gitlab_projects", "gitlab_project_id"))
	assert.True(t, db.Migrator().HasColumn(&Project{}, "gitlab_project_id"))
	assert.True(t, db.Migrator().HasColumn(&Project{}, "gitlab_branch"))
	assert.True(t, db.Migrator().HasColumn(&Project{}, "gitlab_commit"))
	assert.True(t, db.Migrator().HasTable("commands"))
	assert.False(t, db.Migrator().HasColumn(&Project{}, "manifest"))
	assert.False(t, db.Migrator().HasColumn("files", "upload_type"))
	types, err := db.Migrator().ColumnTypes(&GitlabProject{})
	assert.Nil(t, err)
	for _, columnType := range types {
		if columnType.Name() == "global_config" {
			assert.Equal(t, "text", columnType.DatabaseTypeName())
			break
		}
	}
	etypes, err := db.Migrator().ColumnTypes(&Event{})
	assert.Nil(t, err)
	for _, columnType := range etypes {
		if columnType.Name() == "old" || columnType.Name() == "new" {
			assert.Equal(t, "text", columnType.DatabaseTypeName())
		}
	}
	ptypes, err := db.Migrator().ColumnTypes(&Project{})
	assert.Nil(t, err)
	for _, columnType := range ptypes {
		if columnType.Name() == "extra_values" {
			assert.Equal(t, "text", columnType.DatabaseTypeName())
		}
		if columnType.Name() == "final_extra_values" {
			assert.Equal(t, "text", columnType.DatabaseTypeName())
		}
	}
	assert.False(t, db.Migrator().HasTable("cache_locks"))
	assert.False(t, db.Migrator().HasIndex(&models.Event{}, "Action"))
	if db.Migrator().HasTable(&models.AccessToken{}) {
		assert.Nil(t, db.Migrator().DropTable(&models.AccessToken{}))
	}

	assert.False(t, db.Migrator().HasIndex(&models.Namespace{}, "DeletedAt"))
	assert.False(t, db.Migrator().HasIndex(&models.Project{}, "idx_namespace_id_deleted_at"))
	assert.False(t, db.Migrator().HasIndex(&models.Changelog{}, "idx_projectid_config_changed_deleted_at_version"))
	assert.False(t, db.Migrator().HasColumn("projects", "version"))

	assert.Nil(t, ma.AutoMigrate())

	assert.True(t, db.Migrator().HasColumn("projects", "version"))
	assert.False(t, db.Migrator().HasIndex(&models.Changelog{}, "idx_version_projectid_deleted_at_config_changed"))
	assert.True(t, db.Migrator().HasIndex(&models.Changelog{}, "idx_projectid_config_changed_deleted_at_version"))
	assert.True(t, db.Migrator().HasColumn("files", "upload_type"))
	assert.False(t, db.Migrator().HasColumn(&Changelog{}, "gitlab_project_id"))
	assert.False(t, db.Migrator().HasTable("gitlab_projects"))
	assert.False(t, db.Migrator().HasColumn("git_projects", "gitlab_project_id"))
	assert.False(t, db.Migrator().HasColumn(&Project{}, "gitlab_project_id"))
	assert.False(t, db.Migrator().HasColumn(&Project{}, "gitlab_branch"))
	assert.False(t, db.Migrator().HasColumn(&Project{}, "gitlab_commit"))
	assert.False(t, db.Migrator().HasTable("commands"))
	assert.True(t, db.Migrator().HasColumn(&Project{}, "manifest"))
	assert.True(t, db.Migrator().HasIndex(&models.Event{}, "Action"))
	assert.True(t, db.Migrator().HasTable("cache_locks"))
	assert.False(t, db.Migrator().HasColumn(&models.DBCache{}, "id"))
	assert.True(t, db.Migrator().HasTable(&models.AccessToken{}))
	assert.True(t, db.Migrator().HasIndex(&models.Namespace{}, "DeletedAt"))
	assert.True(t, db.Migrator().HasIndex(&models.Project{}, "idx_namespace_id_deleted_at"))

	types, err = db.Migrator().ColumnTypes("git_projects")
	assert.Nil(t, err)
	for _, columnType := range types {
		if columnType.Name() == "global_config" {
			assert.Equal(t, "longtext", columnType.DatabaseTypeName())
			break
		}
	}
	etypes, err = db.Migrator().ColumnTypes("events")
	assert.Nil(t, err)
	for _, columnType := range etypes {
		if columnType.Name() == "old" || columnType.Name() == "new" {
			assert.Equal(t, "longtext", columnType.DatabaseTypeName())
			break
		}
	}

	ptypes, err = db.Migrator().ColumnTypes(&Project{})
	assert.Nil(t, err)
	for _, columnType := range ptypes {
		if columnType.Name() == "extra_values" {
			assert.Equal(t, "longtext", columnType.DatabaseTypeName())
		}
		if columnType.Name() == "final_extra_values" {
			assert.Equal(t, "longtext", columnType.DatabaseTypeName())
		}
	}
}

func TestManager_AutoMigrate2(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.Exec("PRAGMA foreign_keys = ON", nil)
	s, _ := db.DB()
	defer s.Close()

	ma := &Manager{db: db}
	assert.Nil(t, db.AutoMigrate(&models.Project{}, &models.Namespace{}, &models.Changelog{}, &models.GitProject{}, &models.Event{}))

	p := &models.Project{
		Name:      "app",
		Manifest:  "",
		Namespace: models.Namespace{Name: "test"},
	}
	assert.Nil(t, db.Create(p).Error)
	assert.Nil(t, db.Create(&models.Changelog{
		Version:   1,
		Username:  "abc",
		Manifest:  "xxx",
		ProjectID: p.ID,
		GitProject: models.GitProject{
			Name: "xx",
		},
	}).Error)
	assert.Nil(t, db.Create(&models.Changelog{
		Version:   2,
		Username:  "duc",
		Manifest:  "yyy",
		ProjectID: p.ID,
		GitProject: models.GitProject{
			Name: "yy",
		},
	}).Error)

	assert.Nil(t, ma.AutoMigrate())
	var pp models.Project
	db.First(&pp, p.ID)
	assert.Equal(t, "yyy", pp.Manifest)
}
