// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// AccessTokensColumns holds the columns for the "access_tokens" table.
	AccessTokensColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "updated_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "deleted_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "token", Type: field.TypeString, Unique: true, Size: 100},
		{Name: "usage", Type: field.TypeString, Size: 50},
		{Name: "email", Type: field.TypeString, Size: 255, Default: ""},
		{Name: "expired_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "last_used_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "user_info", Type: field.TypeJSON, Nullable: true},
	}
	// AccessTokensTable holds the schema information for the "access_tokens" table.
	AccessTokensTable = &schema.Table{
		Name:       "access_tokens",
		Columns:    AccessTokensColumns,
		PrimaryKey: []*schema.Column{AccessTokensColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "accesstoken_email",
				Unique:  false,
				Columns: []*schema.Column{AccessTokensColumns[6]},
			},
		},
	}
	// CacheLocksColumns holds the columns for the "cache_locks" table.
	CacheLocksColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "key", Type: field.TypeString, Unique: true},
		{Name: "owner", Type: field.TypeString},
		{Name: "expired_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
	}
	// CacheLocksTable holds the schema information for the "cache_locks" table.
	CacheLocksTable = &schema.Table{
		Name:       "cache_locks",
		Columns:    CacheLocksColumns,
		PrimaryKey: []*schema.Column{CacheLocksColumns[0]},
	}
	// ChangelogsColumns holds the columns for the "changelogs" table.
	ChangelogsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "updated_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "deleted_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "version", Type: field.TypeInt, Default: 1},
		{Name: "username", Type: field.TypeString, Size: 100},
		{Name: "config", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"mysql": "longtext"}},
		{Name: "git_branch", Type: field.TypeString, Nullable: true},
		{Name: "git_commit", Type: field.TypeString, Nullable: true},
		{Name: "docker_image", Type: field.TypeJSON, Nullable: true},
		{Name: "env_values", Type: field.TypeJSON, Nullable: true},
		{Name: "extra_values", Type: field.TypeJSON, Nullable: true},
		{Name: "final_extra_values", Type: field.TypeJSON, Nullable: true},
		{Name: "git_commit_web_url", Type: field.TypeString, Nullable: true},
		{Name: "git_commit_title", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "git_commit_author", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "git_commit_date", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "config_changed", Type: field.TypeBool, Default: false},
		{Name: "project_id", Type: field.TypeInt, Nullable: true},
	}
	// ChangelogsTable holds the schema information for the "changelogs" table.
	ChangelogsTable = &schema.Table{
		Name:       "changelogs",
		Columns:    ChangelogsColumns,
		PrimaryKey: []*schema.Column{ChangelogsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "changelogs_projects_changelogs",
				Columns:    []*schema.Column{ChangelogsColumns[18]},
				RefColumns: []*schema.Column{ProjectsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "changelog_project_id_config_changed_deleted_at_version",
				Unique:  false,
				Columns: []*schema.Column{ChangelogsColumns[18], ChangelogsColumns[17], ChangelogsColumns[3], ChangelogsColumns[4]},
			},
		},
	}
	// DbCacheColumns holds the columns for the "db_cache" table.
	DbCacheColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "key", Type: field.TypeString, Unique: true},
		{Name: "value", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"mysql": "longtext"}},
		{Name: "expired_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
	}
	// DbCacheTable holds the schema information for the "db_cache" table.
	DbCacheTable = &schema.Table{
		Name:       "db_cache",
		Columns:    DbCacheColumns,
		PrimaryKey: []*schema.Column{DbCacheColumns[0]},
	}
	// EventsColumns holds the columns for the "events" table.
	EventsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "updated_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "deleted_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "action", Type: field.TypeInt32, Default: 0},
		{Name: "username", Type: field.TypeString, Size: 255, Default: ""},
		{Name: "message", Type: field.TypeString, Size: 255, Default: ""},
		{Name: "old", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"mysql": "longtext"}},
		{Name: "new", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"mysql": "longtext"}},
		{Name: "has_diff", Type: field.TypeBool, Default: false},
		{Name: "duration", Type: field.TypeString, Default: ""},
		{Name: "file_id", Type: field.TypeInt, Nullable: true},
	}
	// EventsTable holds the schema information for the "events" table.
	EventsTable = &schema.Table{
		Name:       "events",
		Columns:    EventsColumns,
		PrimaryKey: []*schema.Column{EventsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "events_files_events",
				Columns:    []*schema.Column{EventsColumns[11]},
				RefColumns: []*schema.Column{FilesColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "event_action",
				Unique:  false,
				Columns: []*schema.Column{EventsColumns[4]},
			},
			{
				Name:    "event_username_created_at",
				Unique:  false,
				Columns: []*schema.Column{EventsColumns[5], EventsColumns[1]},
			},
		},
	}
	// FavoritesColumns holds the columns for the "favorites" table.
	FavoritesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "email", Type: field.TypeString},
		{Name: "namespace_id", Type: field.TypeInt, Nullable: true},
	}
	// FavoritesTable holds the schema information for the "favorites" table.
	FavoritesTable = &schema.Table{
		Name:       "favorites",
		Columns:    FavoritesColumns,
		PrimaryKey: []*schema.Column{FavoritesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "favorites_namespaces_favorites",
				Columns:    []*schema.Column{FavoritesColumns[2]},
				RefColumns: []*schema.Column{NamespacesColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
	}
	// FilesColumns holds the columns for the "files" table.
	FilesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "updated_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "deleted_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "upload_type", Type: field.TypeString, Size: 100, Default: "local"},
		{Name: "path", Type: field.TypeString, Size: 255},
		{Name: "size", Type: field.TypeUint64, Default: 0, SchemaType: map[string]string{"mysql": "int"}},
		{Name: "username", Type: field.TypeString, Size: 255, Default: ""},
		{Name: "namespace", Type: field.TypeString, Size: 100, Default: ""},
		{Name: "pod", Type: field.TypeString, Size: 100, Default: ""},
		{Name: "container", Type: field.TypeString, Size: 100, Default: ""},
		{Name: "container_path", Type: field.TypeString, Size: 255, Default: ""},
	}
	// FilesTable holds the schema information for the "files" table.
	FilesTable = &schema.Table{
		Name:       "files",
		Columns:    FilesColumns,
		PrimaryKey: []*schema.Column{FilesColumns[0]},
	}
	// MembersColumns holds the columns for the "members" table.
	MembersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "updated_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "deleted_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "email", Type: field.TypeString, Size: 50},
		{Name: "namespace_id", Type: field.TypeInt, Nullable: true},
	}
	// MembersTable holds the schema information for the "members" table.
	MembersTable = &schema.Table{
		Name:       "members",
		Columns:    MembersColumns,
		PrimaryKey: []*schema.Column{MembersColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "members_namespaces_members",
				Columns:    []*schema.Column{MembersColumns[5]},
				RefColumns: []*schema.Column{NamespacesColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "member_email",
				Unique:  false,
				Columns: []*schema.Column{MembersColumns[4]},
			},
		},
	}
	// NamespacesColumns holds the columns for the "namespaces" table.
	NamespacesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "updated_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "deleted_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "name", Type: field.TypeString, Size: 100, Collation: "utf8mb4_general_ci"},
		{Name: "image_pull_secrets", Type: field.TypeJSON},
		{Name: "private", Type: field.TypeBool, Default: false},
		{Name: "creator_email", Type: field.TypeString, Size: 50},
		{Name: "description", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"mysql": "text"}},
	}
	// NamespacesTable holds the schema information for the "namespaces" table.
	NamespacesTable = &schema.Table{
		Name:       "namespaces",
		Columns:    NamespacesColumns,
		PrimaryKey: []*schema.Column{NamespacesColumns[0]},
	}
	// ProjectsColumns holds the columns for the "projects" table.
	ProjectsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "updated_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "deleted_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "name", Type: field.TypeString, Size: 100, Default: ""},
		{Name: "git_project_id", Type: field.TypeInt, Nullable: true},
		{Name: "git_branch", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "git_commit", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "config", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"mysql": "longtext"}},
		{Name: "creator", Type: field.TypeString},
		{Name: "override_values", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"mysql": "longtext"}},
		{Name: "docker_image", Type: field.TypeJSON, Nullable: true},
		{Name: "pod_selectors", Type: field.TypeJSON, Nullable: true},
		{Name: "atomic", Type: field.TypeBool, Default: false},
		{Name: "deploy_status", Type: field.TypeInt32, Default: 0},
		{Name: "env_values", Type: field.TypeJSON, Nullable: true},
		{Name: "extra_values", Type: field.TypeJSON, Nullable: true},
		{Name: "final_extra_values", Type: field.TypeJSON, Nullable: true},
		{Name: "version", Type: field.TypeInt, Default: 1},
		{Name: "config_type", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "manifest", Type: field.TypeJSON, Nullable: true},
		{Name: "git_commit_web_url", Type: field.TypeString, Size: 255, Default: ""},
		{Name: "git_commit_title", Type: field.TypeString, Size: 255, Default: ""},
		{Name: "git_commit_author", Type: field.TypeString, Size: 255, Default: ""},
		{Name: "git_commit_date", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "namespace_id", Type: field.TypeInt, Nullable: true},
		{Name: "repo_id", Type: field.TypeInt, Nullable: true},
	}
	// ProjectsTable holds the schema information for the "projects" table.
	ProjectsTable = &schema.Table{
		Name:       "projects",
		Columns:    ProjectsColumns,
		PrimaryKey: []*schema.Column{ProjectsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "projects_namespaces_projects",
				Columns:    []*schema.Column{ProjectsColumns[25]},
				RefColumns: []*schema.Column{NamespacesColumns[0]},
				OnDelete:   schema.SetNull,
			},
			{
				Symbol:     "projects_repos_projects",
				Columns:    []*schema.Column{ProjectsColumns[26]},
				RefColumns: []*schema.Column{ReposColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "project_git_project_id",
				Unique:  false,
				Columns: []*schema.Column{ProjectsColumns[5]},
			},
		},
	}
	// ReposColumns holds the columns for the "repos" table.
	ReposColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "updated_at", Type: field.TypeTime, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "deleted_at", Type: field.TypeTime, Nullable: true, SchemaType: map[string]string{"mysql": "datetime"}},
		{Name: "name", Type: field.TypeString, Size: 255, Collation: "utf8mb4_general_ci"},
		{Name: "default_branch", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "git_project_name", Type: field.TypeString, Nullable: true},
		{Name: "git_project_id", Type: field.TypeInt32, Nullable: true},
		{Name: "enabled", Type: field.TypeBool, Default: false},
		{Name: "need_git_repo", Type: field.TypeBool, Default: false},
		{Name: "mars_config", Type: field.TypeJSON, Nullable: true},
		{Name: "description", Type: field.TypeString, Default: ""},
	}
	// ReposTable holds the schema information for the "repos" table.
	ReposTable = &schema.Table{
		Name:       "repos",
		Columns:    ReposColumns,
		PrimaryKey: []*schema.Column{ReposColumns[0]},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		AccessTokensTable,
		CacheLocksTable,
		ChangelogsTable,
		DbCacheTable,
		EventsTable,
		FavoritesTable,
		FilesTable,
		MembersTable,
		NamespacesTable,
		ProjectsTable,
		ReposTable,
	}
)

func init() {
	ChangelogsTable.ForeignKeys[0].RefTable = ProjectsTable
	DbCacheTable.Annotation = &entsql.Annotation{
		Table: "db_cache",
	}
	EventsTable.ForeignKeys[0].RefTable = FilesTable
	FavoritesTable.ForeignKeys[0].RefTable = NamespacesTable
	MembersTable.ForeignKeys[0].RefTable = NamespacesTable
	ProjectsTable.ForeignKeys[0].RefTable = NamespacesTable
	ProjectsTable.ForeignKeys[1].RefTable = ReposTable
}
