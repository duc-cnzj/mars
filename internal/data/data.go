package data

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/metrics"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/closeable"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	runtime2 "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	eventsv1lister "k8s.io/client-go/listers/events/v1"
	networkingv1 "k8s.io/client-go/listers/networking/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	_ "github.com/duc-cnzj/mars/v5/internal/ent/runtime"
	_ "github.com/go-sql-driver/mysql"
)

//go:generate mockgen -destination ./mock_data.go -package data github.com/duc-cnzj/mars/v5/internal/data Data

type Data interface {
	Config() *config.Config
	DB() *ent.Client
	MinioCli() *minio.Client
	K8sClient() *K8sClient
	OidcConfig() OidcConfig

	InitDB() (func() error, error)
	InitS3() error
	InitK8s(ch <-chan struct{}) (err error)
	InitOidcProvider()

	WithTx(ctx context.Context, fn func(tx *ent.Tx) error) error
}

var _ Data = (*dataImpl)(nil)

type dataImpl struct {
	cfg       *config.Config
	oidc      OidcConfig
	db        *ent.Client
	minioCli  *minio.Client
	k8sClient *K8sClient

	logger mlog.Logger

	initDBOnce  sync.Once
	initK8sOnce sync.Once
	initS3Once  sync.Once
	oidcOnce    sync.Once
}

type NewDataParams struct {
	Cfg       *config.Config
	Oidc      OidcConfig
	DB        *ent.Client
	MinioCli  *minio.Client
	K8sClient *K8sClient
	Logger    mlog.Logger
}

func NewDataImpl(input *NewDataParams) Data {
	return &dataImpl{
		cfg:       input.Cfg,
		oidc:      input.Oidc,
		db:        input.DB,
		minioCli:  input.MinioCli,
		k8sClient: input.K8sClient,
		logger:    input.Logger,
	}
}

func NewData(cfg *config.Config, logger mlog.Logger) Data {
	return NewDataImpl(&NewDataParams{
		Cfg:    cfg,
		Logger: logger.WithModule("data/data"),
	})
}

func (data *dataImpl) Config() *config.Config {
	return data.cfg
}

func (data *dataImpl) DB() *ent.Client {
	return data.db
}

func (data *dataImpl) MinioCli() *minio.Client {
	return data.minioCli
}

func (data *dataImpl) K8sClient() *K8sClient {
	return data.k8sClient
}

func (data *dataImpl) OidcConfig() OidcConfig {
	return data.oidc
}

func (data *dataImpl) InitDB() (func() error, error) {
	var closeFunc func() error

	data.initDBOnce.Do(func() {
		var logger = data.logger
		logger.Debug("connecting to mysql...")
		defer logger.Debug("mysql connected!")

		cfg := data.Config()

		drv, err := OpenDB(cfg)
		if err != nil {
			return
		}
		data.db, err = InitDB(
			drv,
			logger,
			cfg.DBSlowLogEnabled,
			cfg.DBSlowLogThreshold,
			timer.NewRealTimer(),
		)
		if err != nil {
			return
		}

		if cfg.DBDebug {
			data.db = data.DB().Debug()
		}
		closeFunc = func() error {
			return data.DB().Close()
		}
	})
	return closeFunc, nil
}

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
			err = errors.New("s3 config error")
			return
		}

		// Initialize minio client object.
		data.minioCli, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL,
		})
	})
	return err
}

func (data *dataImpl) InitK8s(ch <-chan struct{}) (err error) {
	data.initK8sOnce.Do(func() {
		var (
			cfg      = data.Config()
			logger   = data.logger
			config   *restclient.Config
			nsPrefix = cfg.NsPrefix

			eventFanOutObj = &fanOut[*eventsv1.Event]{
				logger:    logger,
				name:      "event",
				ch:        make(chan Obj[*eventsv1.Event], 10000),
				listeners: make(map[string]chan<- Obj[*eventsv1.Event]),
			}

			podFanOutObj = &fanOut[*corev1.Pod]{
				name:      "pod",
				ch:        make(chan Obj[*corev1.Pod], 10000),
				logger:    logger,
				listeners: make(map[string]chan<- Obj[*corev1.Pod]),
			}
		)
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
				return
			}
		} else {
			config, err = restclient.InClusterConfig()
			if err != nil {
				return
			}
		}

		// 客户端不限速，有可能会把集群打死。
		config.QPS = -1

		var clientset kubernetes.Interface
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return
		}

		var metrics versioned.Interface
		metrics, err = metricsv.NewForConfig(config)
		if err != nil {
			return
		}
		inf := informers.NewSharedInformerFactory(clientset, 0)
		svcLister := inf.Core().V1().Services().Lister()
		ingLister := inf.Networking().V1().Ingresses().Lister()
		rsLister := inf.Apps().V1().ReplicaSets().Lister()
		podInf := inf.Core().V1().Pods().Informer()
		podLister := inf.Core().V1().Pods().Lister()
		secretInf := inf.Core().V1().Secrets().Informer()
		secretLister := inf.Core().V1().Secrets().Lister()
		podInf.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: filterPod(nsPrefix),
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj any) {
					select {
					case podFanOutObj.ch <- NewObj[*corev1.Pod](nil, obj.(*corev1.Pod), Add):
					default:
						logger.Warningf("[INFORMER]: podFanOutObj full")
					}
				},
				UpdateFunc: func(oldObj, newObj any) {
					old := oldObj.(*corev1.Pod)
					curr := newObj.(*corev1.Pod)
					if old.ResourceVersion != curr.ResourceVersion {
						select {
						case podFanOutObj.ch <- NewObj[*corev1.Pod](old, curr, Update):
						default:
							logger.Warningf("[INFORMER]: podFanOutObj full")
						}
					}
				},
				DeleteFunc: func(obj any) {
					select {
					case podFanOutObj.ch <- NewObj[*corev1.Pod](nil, obj.(*corev1.Pod), Delete):
					default:
						logger.Warningf("[INFORMER]: podFanOutObj full")
					}
				},
			},
		})
		eventInf := inf.Events().V1().Events().Informer()
		eventInf.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: filterEvent(nsPrefix),
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(current any) {
					event := current.(*eventsv1.Event)
					select {
					case eventFanOutObj.ch <- NewObj[*eventsv1.Event](nil, event, Add):
					default:
						logger.Warningf("[INFORMER]: eventFanOutObj full")
					}
				},
			},
		})
		eventLister := inf.Events().V1().Events().Lister()
		data.k8sClient = &K8sClient{
			logger:           logger,
			factory:          inf,
			Client:           clientset,
			MetricsClient:    metrics,
			RestConfig:       config,
			PodInformer:      podInf,
			PodLister:        podLister,
			SecretInformer:   secretInf,
			SecretLister:     secretLister,
			ReplicaSetLister: rsLister,
			ServiceLister:    svcLister,
			IngressLister:    ingLister,
			EventInformer:    eventInf,
			EventFanOut:      eventFanOutObj,
			PodFanOut:        podFanOutObj,
			EventLister:      eventLister,
		}
		data.k8sClient.start(ch)
	})
	return
}

func (data *dataImpl) InitOidcProvider() {
	data.oidcOnce.Do(func() {
		var (
			oidcConfig OidcConfig = make(OidcConfig)
			cfg                   = data.Config()
			logger                = data.logger
		)
		logger.Info("init oidc provider...")
		for _, setting := range cfg.Oidc {
			if !setting.Enabled {
				continue
			}
			var (
				err      error
				provider *oidc.Provider
			)
			if provider, err = oidc.NewProvider(context.TODO(), setting.ProviderUrl); err != nil {
				return
			}

			var ev extraValues
			if err = provider.Claims(&ev); err != nil {
				return
			}
			addOidcCfg(provider, ev, setting, oidcConfig)
		}
	})
}

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

type extraValues struct {
	CheckSessionIFrame string   `json:"check_session_iframe"`
	ScopesSupported    []string `json:"scopes_supported"`
	EndSessionEndpoint string   `json:"end_session_endpoint"`
}

type OidcConfigItem struct {
	Provider           *oidc.Provider
	Config             oauth2.Config
	EndSessionEndpoint string
}

type OidcConfig map[string]OidcConfigItem

func addOidcCfg(provider *oidc.Provider, extraValues extraValues, setting config.OidcSetting, oidcConfig OidcConfig) {
	scopes := extraValues.ScopesSupported
	if len(scopes) < 1 {
		scopes = []string{oidc.ScopeOpenID}
	}

	oauth2Config := oauth2.Config{
		ClientID:     setting.ClientID,
		ClientSecret: setting.ClientSecret,
		RedirectURL:  setting.RedirectUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}
	oidcConfig[setting.Name] = OidcConfigItem{
		Provider:           provider,
		Config:             oauth2Config,
		EndSessionEndpoint: extraValues.EndSessionEndpoint,
	}
}

type K8sClient struct {
	logger        mlog.Logger
	factory       informers.SharedInformerFactory
	Client        kubernetes.Interface
	MetricsClient versioned.Interface
	RestConfig    *restclient.Config

	PodInformer cache.SharedIndexInformer
	PodLister   v1.PodLister

	SecretInformer cache.SharedIndexInformer
	SecretLister   v1.SecretLister

	ReplicaSetLister appsv1.ReplicaSetLister
	ServiceLister    v1.ServiceLister
	IngressLister    networkingv1.IngressLister

	EventInformer cache.SharedIndexInformer

	EventFanOut FanOutInterface[*eventsv1.Event]
	PodFanOut   FanOutInterface[*corev1.Pod]
	EventLister eventsv1lister.EventLister
}

func (k *K8sClient) start(done <-chan struct{}) {
	go func() {
		defer k.logger.HandlePanic("[FANOUT]: event Distribute")

		k.EventFanOut.Distribute(done)
	}()
	go func() {
		defer k.logger.HandlePanic("[FANOUT]: pod Distribute")

		k.PodFanOut.Distribute(done)
	}()
	k.factory.Start(done)
	cache.WaitForCacheSync(done, k.EventInformer.HasSynced, k.PodInformer.HasSynced, k.SecretInformer.HasSynced)
}

type FanOutType int

const (
	Add FanOutType = iota
	Update
	Delete
)

func filterEvent(nsPrefix string) func(obj any) bool {
	return func(obj any) bool {
		e, ok := obj.(*eventsv1.Event)
		if !ok {
			return false
		}
		return strings.HasPrefix(e.Namespace, nsPrefix) && e.Regarding.Kind == "Pod" && e.Reason != "Unhealthy"
	}
}
func filterPod(nsPrefix string) func(obj any) bool {
	return func(obj any) bool {
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			return false
		}
		return strings.HasPrefix(pod.Namespace, nsPrefix)
	}
}

type Startable struct {
	c closeable.Closeable
}

func (s *Startable) start() bool {
	return s.c.Close()
}

type FanOutInterface[T runtime2.Object] interface {
	RemoveListener(key string)
	AddListener(key string, ch chan Obj[T])
	Distribute(done <-chan struct{})
	RemoveAll()
}

type Obj[T runtime2.Object] interface {
	Type() FanOutType
	Old() T
	Current() T
}

type fanOut[T runtime2.Object] struct {
	name string
	ch   chan Obj[T]

	started Startable
	logger  mlog.Logger

	listenerMu sync.Mutex
	listeners  map[string]chan<- Obj[T]
}

func NewFanOut[T runtime2.Object](
	logger mlog.Logger,
	name string,
	ch chan Obj[T],
	listeners map[string]chan<- Obj[T],
) FanOutInterface[T] {
	return &fanOut[T]{name: name, ch: ch, logger: logger, listeners: listeners}
}

func (f *fanOut[T]) AddListener(key string, ch chan Obj[T]) {
	f.listenerMu.Lock()
	defer f.listenerMu.Unlock()
	_, ok := f.listeners[key]
	if ok {
		f.logger.Warningf("[FANOUT]: FanOut already exists %s", key)
		return
	}
	f.logger.Infof("%s add fanOut listener: %v", f.name, key)
	metrics.K8sInformerFanOutListenerCount.With(prometheus.Labels{"type": f.name}).Inc()
	f.listeners[key] = ch
}

func (f *fanOut[T]) RemoveListener(key string) {
	f.listenerMu.Lock()
	defer f.listenerMu.Unlock()
	f.logger.Infof("[FANOUT]: remove listener %s", key)
	ch, ok := f.listeners[key]
	if ok {
		delete(f.listeners, key)
		close(ch)
		metrics.K8sInformerFanOutListenerCount.With(prometheus.Labels{"type": f.name}).Dec()
	}
}

func (f *fanOut[T]) Distribute(done <-chan struct{}) {
	defer f.logger.Debug(fmt.Sprintf("[FANOUT]: '%s' Exit", f.name))
	if !f.started.start() {
		return
	}
	defer f.RemoveAll()
	f.logger.Infof("[FANOUT]: '%s' start", f.name)
	for {
		select {
		case <-done:
			f.logger.Infof("[FANOUT]: '%s' exited!", f.name)
			return
		case obj, ok := <-f.ch:
			if !ok {
				f.logger.Warningf("[FANOUT]: '%s' Exit!", f.name)
				return
			}
			func() {
				f.listenerMu.Lock()
				defer f.listenerMu.Unlock()
				for k, s := range f.listeners {
					select {
					case s <- obj:
					default:
						f.logger.Warningf("[FANOUT]: '%s' drop %s %v", f.name, k, obj)
					}
				}
			}()
		}
	}
}

func (f *fanOut[T]) RemoveAll() {
	f.listenerMu.Lock()
	defer f.listenerMu.Unlock()
	for k, s := range f.listeners {
		close(s)
		delete(f.listeners, k)
	}
}

type obj[T runtime2.Object] struct {
	old, current T
	t            FanOutType
}

func NewObj[T runtime2.Object](old T, current T, t FanOutType) Obj[T] {
	return &obj[T]{old: old, current: current, t: t}
}

func (o *obj[T]) Type() FanOutType {
	return o.t
}

func (o *obj[T]) Old() T {
	return o.old
}

func (o *obj[T]) Current() T {
	return o.current
}
