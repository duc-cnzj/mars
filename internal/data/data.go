package data

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/migrate"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	gwclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
	"sigs.k8s.io/gateway-api/pkg/client/informers/externalversions"
	gwinformers "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/apis/v1"
	gatewaylisterv1 "sigs.k8s.io/gateway-api/pkg/client/listers/apis/v1"
)

//go:generate go tool mockgen -destination ./mock_data.go -package data github.com/duc-cnzj/mars/v6/internal/data Data

// dataCommon 是 data 域对内通用的访问基座：配置读取（Config）+ 事务（WithTx）。
// Data 与 dataStore 都内嵌它，避免两接口重复声明同一组方法；未导出，仅 data
// 包内可见。OidcConfig 不进基座——repo 不读它，只由启动门面 Data 对外暴露。
type dataCommon interface {
	// Config 返回全局配置。
	Config() *config.Config
	// WithTx 在事务中执行闭包。
	WithTx(ctx context.Context, fn func(tx *ent.Tx) error) error
}

// Data 是数据访问的启动期门面（对外）：只暴露配置读取（Config/OidcConfig）、
// 事务（WithTx）与初始化能力（建连、schema 迁移、provider 装配），不含运行时
// DB/k8s/minio 访问。需要读写的组件应依赖各自 Repository 端口，而非持有 Data
// 门面——基础设施细节被收在 data 包内，禁止跨层摸 ent/k8s 客户端。
// 消费方：app/plugins 的 bootstrapper。
type Data interface {
	dataCommon
	// OidcConfig 返回 OIDC 配置。
	OidcConfig() biz.OidcConfig

	// InitDB 建连并迁移 schema，返回关闭函数与首个错误。
	InitDB() (func() error, error)
	// InitS3 初始化 minio 客户端。
	InitS3() error
	// InitK8s 初始化 k8s 客户端与 informer。
	InitK8s(ch <-chan struct{}) (err error)
	// InitOidcProvider 装配 OIDC 登录 provider。
	InitOidcProvider()
	// Migrate 执行数据库 schema 自动迁移（DBBootstrapper 在 DBAutoMigrate 时调用）。
	Migrate() error
}

// dataStore 是 data 包内 Repository 的存储访问端口（对内）：在 dataCommon 基座
// 之上追加 DB()/K8s() 两个 repo 实际使用的基础设施访问器。未导出，仅 data
// 包内 repo 构造器注入使用。MinioCli 不经 dataStore——repo 不摸 minio 客户端，
// cmd 装配期只经 MinioGetter 窄端口取 MinioCli 惰性函数，避免 ent/k8s 客户端
// 整体跨层泄漏。
type dataStore interface {
	dataCommon
	// DB 返回 ent 客户端。
	DB() *ent.Client
	// K8s 返回封装的 K8s 客户端。
	K8s() *K8sClient
}

// MinioGetter 是对外窄端口（cmd 装配期）：只暴露 MinioCli 惰性取数，供
// provideMinioClient 转成 uploader 需要的闭包；cmd 不持有 dataStore 全貌。
type MinioGetter interface {
	// MinioCli 返回 minio 客户端。
	MinioCli() *minio.Client
}

// DBGetter 是对外窄端口（cmd 装配期）：只暴露 DB 惰性取数，供 provideDBGetter
// 转成 locker 需要的闭包；cmd 不持有 dataStore 全貌。
type DBGetter interface {
	// DB 返回 ent 客户端。
	DB() *ent.Client
}

var _ Data = (*dataImpl)(nil)
var _ dataStore = (*dataImpl)(nil)

// dataImpl 是 Data/dataStore 接口的默认实现：持有配置、OIDC、DB、Minio、K8s 客户端
// 与日志，用 sync.Once 保证各 Init* 初始化只执行一次。
type dataImpl struct {
	cfg       *config.Config
	oidc      biz.OidcConfig
	db        *ent.Client
	minioCli  *minio.Client
	k8sClient *K8sClient

	logger mlog.Logger

	initDBOnce  sync.Once
	initK8sOnce sync.Once
	initS3Once  sync.Once
	oidcOnce    sync.Once
}

// NewDataParams 是 NewDataImpl 的依赖注入参数：显式传入配置与各客户端，避免构造器从全局拉取。
type NewDataParams struct {
	Cfg       *config.Config
	Oidc      biz.OidcConfig
	DB        *ent.Client
	MinioCli  *minio.Client
	K8sClient *K8sClient
	Logger    mlog.Logger
}

// NewDataImpl 用显式参数构造 dataImpl 实现。
func NewDataImpl(input *NewDataParams) *dataImpl {
	return &dataImpl{
		cfg:       input.Cfg,
		oidc:      input.Oidc,
		db:        input.DB,
		minioCli:  input.MinioCli,
		k8sClient: input.K8sClient,
		logger:    input.Logger,
	}
}

// NewData 以配置与日志构造 dataImpl（日志带 data/data 模块名）；DB/k8s/minio 客户端由后续 Init* 初始化注入。
func NewData(cfg *config.Config, logger mlog.Logger) *dataImpl {
	return NewDataImpl(&NewDataParams{
		Cfg:    cfg,
		Logger: logger.WithModule("data/data"),
	})
}

// Config 返回全局配置（dataCommon 基座方法）。
func (data *dataImpl) Config() *config.Config {
	return data.cfg
}

// DB 返回 ent 客户端（dataStore 基座方法）。
func (data *dataImpl) DB() *ent.Client {
	return data.db
}

// MinioCli 返回 minio 客户端（MinioGetter 窄端口实现）。
func (data *dataImpl) MinioCli() *minio.Client {
	return data.minioCli
}

// K8s 返回封装的 K8s 客户端（dataStore 基座方法）。
func (data *dataImpl) K8s() *K8sClient {
	return data.k8sClient
}

// OidcConfig 返回 OIDC 配置（Data 门面对外暴露）。
func (data *dataImpl) OidcConfig() biz.OidcConfig {
	return data.oidc
}

// AdminPassword 返回内置 admin 账号的登录密码，供 biz.AuthConfigProvider 取数。
func (data *dataImpl) AdminPassword() string {
	return data.cfg.AdminPassword
}

// InitDB 建立数据库连接并执行 schema 迁移（once 幂等）；返回关闭函数与首个错误。
func (data *dataImpl) InitDB() (func() error, error) {
	var (
		closeFunc func() error
		err       error
	)

	data.initDBOnce.Do(func() {
		var logger = data.logger
		logger.Debug("connecting to mysql...")
		defer logger.Debug("mysql connected!")

		cfg := data.Config()

		// 错误经外层 err 透出，否则 once.Do 吞错 → 调用方（DBBootstrapper）
		// 在 DB 初始化失败时无感知继续启动。与 InitS3 的捕获模式一致。
		drv, openErr := OpenDB(cfg)
		if openErr != nil {
			err = errs.Wrap(openErr, "open db")
			return
		}
		// InitDB 已无错误返回（签名不带 error），OpenDB 的失败已在上方早退。
		data.db = InitDB(
			drv,
			logger,
			cfg.DBSlowLogEnabled,
			cfg.DBSlowLogThreshold,
			timer.NewReal(),
		)

		if cfg.DBDebug {
			data.db = data.DB().Debug()
		}
		closeFunc = func() error {
			return errs.Wrap(data.DB().Close(), "close db")
		}
	})
	return closeFunc, err
}

// Migrate 执行数据库 schema 自动迁移（drop index/column 策略与 DBAutoMigrate 联动）。
// 迁移细节收敛在 data 包内，bootstrapper 不再触碰 ent 客户端。
func (data *dataImpl) Migrate() error {
	return errs.Wrap(data.DB().Schema.Create(context.TODO(), migrate.WithDropIndex(true), migrate.WithDropColumn(true)), "migrate database")
}

// InitS3 按配置初始化 minio 客户端（once 幂等）；S3 未启用或缺省凭据时跳过。
func (data *dataImpl) InitS3() error {
	var err error

	data.initS3Once.Do(func() {
		var (
			cfg             = data.Config()
			endpoint        = cfg.S3Endpoint
			accessKeyID     = cfg.S3AccessKeyID
			secretAccessKey = cfg.S3SecretAccessKey
			useSSL          = cfg.S3UseSSL
		)
		data.logger.Info("init s3 client...")
		if !cfg.S3Enabled {
			return
		}
		if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
			err = errs.Wrap(errors.New("s3 config error"), "init s3")
			return
		}

		// Initialize minio client object.
		data.minioCli, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			err = errs.Wrap(err, "init s3")
		}
	})
	return err
}

// newK8sClientset 是 kubernetes 客户端构造缝：测试可替换为 fake clientset
// （内存 list/watch），以覆盖 InitK8s 成功路径的 informer 装配。
var newK8sClientset = func(config *restclient.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(config)
}

// InitK8s 建立 k8s 客户端与各类 informer/lister（once 幂等），装配事件/容器事件
// 扇出通道并启动监听；依赖真实集群，属集成边界。
func (data *dataImpl) InitK8s(ch <-chan struct{}) (err error) {
	data.initK8sOnce.Do(func() {
		var (
			cfg      = data.Config()
			logger   = data.logger
			config   *restclient.Config
			nsPrefix = cfg.NsPrefix
		)
		// 事件/容器事件扇出的输入通道：informer 回调写入、Distribute 消费广播。
		// 通道独立持有、经 newFanOut 注入 fanOut，回调闭包只写通道、不摸私有字段。
		eventCh := make(chan Obj[*eventsv1.Event], 1000)
		eventFanOutObj := newFanOut(logger, "event", eventCh, make(map[string]chan<- Obj[*eventsv1.Event]))
		podCh := make(chan Obj[*corev1.Pod], 1000)
		podFanOutObj := newFanOut(logger, "pod", podCh, make(map[string]chan<- Obj[*corev1.Pod]))
		logger.Info("init k8s client...")

		runtime.ErrorHandlers = []func(err error){
			func(err error) {
				logger.Warning(err)
			},
		}

		logger.Warning(cfg.KubeConfig)
		if cfg.KubeConfig != "" {
			config, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
			if err != nil {
				err = errs.Wrap(err, "build kubeconfig")
				return
			}
		} else {
			config, err = restclient.InClusterConfig()
			if err != nil {
				err = errs.Wrap(err, "load in-cluster config")
				return
			}
		}

		// 客户端不限速，有可能会把集群打死。
		config.QPS = -1

		var clientset kubernetes.Interface
		clientset, err = newK8sClientset(config)
		if err != nil {
			err = errs.Wrap(err, "new k8s clientset")
			return
		}

		var gwinstalled bool
		// 后续 apiextensions/metrics/gateway 客户端与 clientset 共享同一份 rest 配置、
		// 同一套 HTTPClientFor 校验，前文 newK8sClientset 已用同一 config 成功——
		// 此处构造不可能失败，用 OrDie 收敛（杜绝不可达错误分支）。
		crdList, crdErr := apiextensionsv1.NewForConfigOrDie(config).
			ApiextensionsV1().
			CustomResourceDefinitions().
			List(context.TODO(), metav1.ListOptions{})
		if crdErr != nil {
			err = errs.Wrap(crdErr, "list gateway crds")
			return
		}
		for _, crd := range crdList.Items {
			if crd.Name == "httproutes.gateway.networking.k8s.io" {
				gwinstalled = true
				break
			}
		}

		metrics := versioned.NewForConfigOrDie(config)
		inf := informers.NewSharedInformerFactory(clientset, 0)

		var httpRouteLister gatewaylisterv1.HTTPRouteLister
		var gwhttprouteinformer externalversions.SharedInformerFactory
		if gwinstalled {
			logger.Info("gateway api installed")
			gwhttprouteinformer = externalversions.NewSharedInformerFactoryWithOptions(gwclientset.NewForConfigOrDie(config), 0)
			httpRouteLister = gwinformers.New(gwhttprouteinformer, corev1.NamespaceAll, nil).
				HTTPRoutes().
				Lister()
		}

		svcLister := inf.Core().V1().Services().Lister()
		ingLister := inf.Networking().V1().Ingresses().Lister()
		rsLister := inf.Apps().V1().ReplicaSets().Lister()
		deployLister := inf.Apps().V1().Deployments().Lister()
		stsLister := inf.Apps().V1().StatefulSets().Lister()
		dsLister := inf.Apps().V1().DaemonSets().Lister()
		podInf := inf.Core().V1().Pods().Informer()
		podLister := inf.Core().V1().Pods().Lister()
		secretInf := inf.Core().V1().Secrets().Informer()
		secretLister := inf.Core().V1().Secrets().Lister()
		podInf.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: filterPod(nsPrefix),
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj any) {
					sendOrDrop(podCh, newObj[*corev1.Pod](nil, obj.(*corev1.Pod), Add), logger, "podFanOutObj")
				},
				UpdateFunc: func(oldObj, newPod any) {
					old := oldObj.(*corev1.Pod)
					curr := newPod.(*corev1.Pod)
					if old.ResourceVersion != curr.ResourceVersion {
						sendOrDrop(podCh, newObj[*corev1.Pod](old, curr, Update), logger, "podFanOutObj")
					}
				},
				DeleteFunc: func(obj any) {
					sendOrDrop(podCh, newObj[*corev1.Pod](nil, obj.(*corev1.Pod), Delete), logger, "podFanOutObj")
				},
			},
		})
		eventInf := inf.Events().V1().Events().Informer()
		eventInf.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: filterEvent(nsPrefix),
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(current any) {
					event := current.(*eventsv1.Event)
					sendOrDrop(eventCh, newObj[*eventsv1.Event](nil, event, Add), logger, "eventFanOutObj")
				},
			},
		})
		eventLister := inf.Events().V1().Events().Lister()
		data.k8sClient = &K8sClient{
			GatewayApiInstalled: gwinstalled,
			HTTPRouteLister:     httpRouteLister,
			gwFactory:           gwhttprouteinformer,
			logger:              logger,
			factory:             inf,
			Client:              clientset,
			MetricsClient:       metrics,
			RestConfig:          config,
			PodInformer:         podInf,
			PodLister:           podLister,
			SecretInformer:      secretInf,
			SecretLister:        secretLister,
			ReplicaSetLister:    rsLister,
			DeploymentLister:    deployLister,
			StatefulSetLister:   stsLister,
			DaemonSetLister:     dsLister,
			ServiceLister:       svcLister,
			IngressLister:       ingLister,
			eventFanOut:         eventFanOutObj,
			podFanOut:           podFanOutObj,
			EventLister:         eventLister,
		}
		data.k8sClient.start(ch)
	})
	return
}

// WithTx 在事务中执行 fn：出错回滚、panic 回滚后重抛、成功提交。
func (data *dataImpl) WithTx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	tx, err := data.DB().Tx(ctx)
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
