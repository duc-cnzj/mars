package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	gorm2 "github.com/duc-cnzj/mars/v4/examples/v5migrate/gorm"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/migrate"
	"github.com/duc-cnzj/mars/v4/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v4/internal/ent/repo"
	_ "github.com/duc-cnzj/mars/v4/internal/ent/runtime"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/releaseutil"
)

type DB struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type Config struct {
	V4 DB `json:"v4"`
	V5 DB `json:"v5"`
}

var cfg Config = Config{
	V4: DB{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		Username: "root",
		Password: "",
		Database: "marsv4-prod",
	},
	V5: DB{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		Username: "root",
		Password: "",
		Database: "marsv5-migrate",
	},
}

func main() {
	v4DB, err := connGorm(cfg.V4)
	logFatal(err)

	v5DB, err := InitEnt(cfg.V5)
	logFatal(err)
	err = v5DB.Schema.Create(
		context.TODO(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
	logFatal(err)

	log.Println("开始迁移数据")
	migrateAccessToken(v4DB, v5DB)
	migrateNamespace(v4DB, v5DB)
	migrateGitProjectToRepo(v4DB, v5DB)
	migrateProject(v4DB, v5DB)
}

func migrateProject(gdb *gorm.DB, edb *ent.Client) {
	var projects []gorm2.Project
	gdb.Model(&gorm2.Project{}).Preload("Namespace").Find(&projects)
	log.Println("projects len: ", len(projects))
	if edb.Project.Query().CountX(context.TODO()) != 0 {
		log.Println("[skip]: migrateProject")
		return
	}

	err := WithTx(context.TODO(), edb, func(edb *ent.Tx) error {
		for _, project := range projects {
			first, err := edb.Namespace.Query().Where(namespace.Name(project.Namespace.Name)).First(context.TODO())
			if err != nil {
				return err
			}
			var envValues map[string]string
			if err := json.Unmarshal([]byte(project.EnvValues), &envValues); err != nil {
				log.Println("EnvValues err:", err)
				return err
			}
			var ev []*types.KeyValue
			for k, v := range envValues {
				ev = append(ev, &types.KeyValue{
					Key:   k,
					Value: v,
				})
			}
			var extraValues []*websocket_pb.ExtraValue
			if project.ExtraValues != "" {
				if err := json.Unmarshal([]byte(project.ExtraValues), &extraValues); err != nil {
					log.Println("ExtraValues err:", err, project.ExtraValues)
					return err
				}
			}
			//if project.FinalExtraValues != "" {
			//	var finExtraValues []*websocket_pb.ExtraValue
			//	if err := json.Unmarshal([]byte(project.FinalExtraValues), &finExtraValues); err != nil {
			//		log.Println("finExtraValues err:", err, project.FinalExtraValues)
			//		return err
			//	}
			//}
			repo, err := edb.Repo.Query().Where(repo.GitProjectID(int32(project.GitProjectId))).First(context.TODO())
			if err != nil {
				return err
			}
			values, err := finalExtraValues(repo.MarsConfig, extraValues)
			if err != nil {
				log.Println(repo.ID, project.ID, err)
				return err
			}
			if _, err = edb.Project.Create().
				SetName(project.Name).
				SetGitProjectID(project.GitProjectId).
				SetGitBranch(project.GitBranch).
				SetGitCommit(project.GitCommit).
				SetConfig(project.Config).
				SetOverrideValues(project.OverrideValues).
				SetDockerImage(strings.Split(project.DockerImage, " ")).
				SetPodSelectors(strings.Split(project.PodSelectors, "|")).
				SetNamespaceID(first.ID).
				SetAtomic(project.Atomic).
				SetDeployStatus(types.Deploy(project.DeployStatus)).
				SetEnvValues(ev).
				SetExtraValues(extraValues).
				SetFinalExtraValues(values).
				SetRepoID(repo.ID).
				SetVersion(project.Version).
				SetConfigType(project.ConfigType).
				SetManifest(SplitManifests(project.Manifest)).
				SetGitCommitWebURL(project.GitCommitWebUrl).
				SetGitCommitTitle(project.GitCommitTitle).
				SetGitCommitAuthor(project.GitCommitAuthor).
				SetGitCommitDate(*project.GitCommitDate).
				SetCreator("1025434218@qq.com").
				Save(context.TODO()); err != nil {
				return err
			}

		}
		return nil
	})
	logFatal(err)
}

func finalExtraValues(cfg *mars.Config, ExtraValues []*websocket_pb.ExtraValue) ([]*websocket_pb.ExtraValue, error) {
	if cfg == nil || ExtraValues == nil {
		return nil, nil
	}
	var validValuesMap = make(map[string]any)
	var useDefaultMap = make(map[string]bool)

	var configElementsMap = make(map[string]*mars.Element)
	//indent, _ := json.MarshalIndent(cfg.Elements, " ", "    ")
	//log.Println(string(indent))
	//log.Println("before sort")
	sort.Slice(cfg.Elements, func(x, y int) bool {
		return cfg.Elements[x].Order < cfg.Elements[y].Order
	})
	//indent2, _ := json.MarshalIndent(cfg.Elements, " ", "    ")
	//log.Println("after sort")
	//log.Println(string(indent2))
	//if len(ExtraValues) > 2 {
	//	os.Exit(1)
	//}
	for _, element := range cfg.Elements {
		configElementsMap[element.Path] = element
		defaultValue, e := typedValue(element, element.Default)
		if e != nil {
			return nil, e
		}
		validValuesMap[element.Path] = defaultValue
		useDefaultMap[element.Path] = true
	}

	// validate
	for _, value := range ExtraValues {
		if element, ok := configElementsMap[value.Path]; ok {
			useDefaultMap[value.Path] = false
			typeValue, err := typedValue(element, value.Value)
			if err != nil {
				return nil, err
			}
			validValuesMap[value.Path] = typeValue
		}
	}

	var finalValues []*websocket_pb.ExtraValue
	for _, element := range cfg.Elements {
		finalValues = append(finalValues, &websocket_pb.ExtraValue{
			Path:  element.Path,
			Value: fmt.Sprintf("%v", validValuesMap[element.Path]),
		})
	}
	return finalValues, nil
}

func typedValue(element *mars.Element, input string) (any, error) {
	switch element.Type {
	case mars.ElementType_ElementTypeSwitch:
		if input == "" {
			input = "false"
		}
		v, err := strconv.ParseBool(input)
		if err != nil {
			return nil, fmt.Errorf("%s 字段类型不正确，应该为 bool，你传入的是 %s", element.Path, input)
		}
		return v, nil
	case mars.ElementType_ElementTypeInputNumber:
		if input == "" {
			input = "0"
		}
		v, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s 字段类型不正确，应该为整数，你传入的是 %s", element.Path, input)
		}
		return v, nil
	case mars.ElementType_ElementTypeRadio,
		mars.ElementType_ElementTypeSelect,
		mars.ElementType_ElementTypeNumberSelect,
		mars.ElementType_ElementTypeNumberRadio:
		var in bool
		for _, selectValue := range element.SelectValues {
			if input == selectValue {
				in = true
				break
			}
		}
		if !in {
			return nil, fmt.Errorf("%s 必须在 '%v' 里面, 你传的是 %s", element.Path, strings.Join(element.SelectValues, ","), input)
		}
		if element.Type == mars.ElementType_ElementTypeNumberSelect ||
			element.Type == mars.ElementType_ElementTypeNumberRadio {
			if atoi, err := strconv.Atoi(input); err == nil {
				return atoi, nil
			}
			return nil, fmt.Errorf("[ElementsLoader]: '%v' 非 number 类型, 无法转换 %v %v", input, element.Path, element.Type)
		}

		return input, nil
	default:
		return input, nil
	}
}

func SplitManifests(manifest string) []string {
	mapManifests := releaseutil.SplitManifests(manifest)
	var manifests []string
	for _, s := range mapManifests {
		manifests = append(manifests, s)
	}
	return manifests
}

func migrateGitProjectToRepo(gdb *gorm.DB, edb *ent.Client) {
	var gitProjects []gorm2.GitProject
	gdb.Find(&gitProjects)
	log.Println("gitProjects len: ", len(gitProjects))
	if edb.Repo.Query().CountX(context.TODO()) != 0 {
		log.Println("[skip]: migrateGitProjectToRepo")
		return
	}
	err := WithTx(context.TODO(), edb, func(edb *ent.Tx) error {
		for _, gpro := range gitProjects {
			var marsC mars.Config
			if err := json.Unmarshal([]byte(gpro.GlobalConfig), &marsC); err != nil {
				log.Println(gpro.ID)
				if gpro.Enabled {
					return err
				}
			}
			name := gpro.Name
			if marsC.DisplayName != "" {
				name = marsC.DisplayName
			}
			split := strings.Split(marsC.LocalChartPath, "|")
			if len(split) != 3 {
				marsC.LocalChartPath = fmt.Sprintf("%v|%v|%v", gpro.GitProjectId, gpro.DefaultBranch, marsC.LocalChartPath)
				log.Println(marsC.DisplayName, marsC.LocalChartPath)
			}
			//log.Println(marsC.DisplayName, gpro.ID, gpro.Name)
			if _, err := edb.Repo.Create().
				SetName(name).
				SetGitProjectID(int32(gpro.GitProjectId)).
				SetGitProjectName(gpro.Name).
				SetDefaultBranch(gpro.DefaultBranch).
				SetEnabled(gpro.Enabled).
				SetNeedGitRepo(true).
				SetMarsConfig(&marsC).Save(context.TODO()); err != nil {
				return err
			}
		}
		return nil
	})
	logFatal(err)
}

func migrateNamespace(gdb *gorm.DB, edb *ent.Client) {
	var gormNamespaces []gorm2.Namespace
	gdb.Find(&gormNamespaces)
	log.Println("gormNamespaces len: ", len(gormNamespaces))
	if edb.Namespace.Query().CountX(context.TODO()) != 0 {
		log.Println("[skip]: migrateNamespace")
		return
	}
	err := WithTx(context.TODO(), edb, func(edb *ent.Tx) error {
		for _, gns := range gormNamespaces {
			split := strings.Split(gns.ImagePullSecrets, ",")
			_, err := edb.Namespace.Create().SetName(gns.Name).SetImagePullSecrets(split).Save(context.TODO())
			if err != nil {
				return err
			}
		}
		return nil
	})
	logFatal(err)
}

func migrateAccessToken(gdb *gorm.DB, edb *ent.Client) {
	var gormAccessTokens []gorm2.AccessToken
	gdb.Model(&gorm2.AccessToken{}).Find(&gormAccessTokens)

	log.Println("gormAccessTokens len: ", len(gormAccessTokens))

	if edb.AccessToken.Query().CountX(context.TODO()) != 0 {
		log.Println("[skip]: migrateAccessToken")
		return
	}
	err := WithTx(context.TODO(), edb, func(tx *ent.Tx) error {
		for _, token := range gormAccessTokens {
			var user schematype.UserInfo
			if err := json.Unmarshal([]byte(token.UserInfo), &user); err != nil {
				return err
			}
			if _, err := tx.AccessToken.Create().
				SetToken(token.Token).
				SetUsage(token.Usage).
				SetExpiredAt(token.ExpiredAt).
				SetEmail(token.Email).
				SetUserInfo(user).
				SetLastUsedAt(token.LastUsedAt.Time).Save(context.TODO()); err != nil {
				return err
			}
		}
		return nil
	})
	logFatal(err)
}

func WithTx(ctx context.Context, db *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func connGorm(cfg DB) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	log.Println(dsn)
	switch cfg.Driver {
	case "mysql":
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "sqlite":
		return gorm.Open(sqlite.Open(cfg.Database), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	}
	return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
}

func (c *DB) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.Username, c.Password, c.Host, c.Port, c.Database)
}

func InitEnt(cfg DB) (*ent.Client, error) {
	var (
		drv dialect.Driver
		err error
	)
	drv, err = sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}
	// Get the underlying sql.DB object of the driver.
	db := drv.(*sql.Driver).DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
	dbCli := ent.NewClient(
		ent.Driver(drv),
	)

	return dbCli, nil
}
