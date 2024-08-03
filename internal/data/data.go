package data

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/closeable"
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

	_ "github.com/duc-cnzj/mars/v4/internal/ent/runtime"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type Data struct {
	Cfg       *config.Config
	Oidc      OidcConfig
	DB        *ent.Client
	MinioCli  *minio.Client
	K8sClient *K8sClient
}

func NewData(cfg *config.Config, logger mlog.Logger) (*Data, func(), error) {
	client, err := OpenDB(cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	var miniocli *minio.Client
	if miniocli, err = NewS3(cfg); err != nil {
		return nil, nil, err
	}
	newOidc := NewOidc(cfg)
	var sClient *K8sClient
	if cfg.KubeConfig != "" {
		if sClient, err = NewK8sClient(cfg, logger); err != nil {
			return nil, nil, err
		}
	}

	cleanup := func() {
		logger.Flush()
		client.Close()
	}

	return &Data{
		Cfg:       cfg,
		Oidc:      newOidc,
		DB:        client,
		MinioCli:  miniocli,
		K8sClient: sClient,
	}, cleanup, nil
}

func (data *Data) WithTx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	tx, err := data.DB.Tx(ctx)
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

func NewOidc(cfg *config.Config) OidcConfig {
	var oidcConfig OidcConfig = make(OidcConfig)
	for _, setting := range cfg.Oidc {
		if !setting.Enabled {
			continue
		}
		var (
			err      error
			provider *oidc.Provider
		)
		if provider, err = oidc.NewProvider(context.TODO(), setting.ProviderUrl); err != nil {
			return nil
		}

		var ev extraValues
		if err = provider.Claims(&ev); err != nil {
			return nil
		}
		addOidcCfg(provider, ev, setting, oidcConfig)
	}

	return oidcConfig
}

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

func OpenDB(cfg *config.Config, logger mlog.Logger) (*ent.Client, error) {
	logger.Debug("connecting to mysql...")
	defer logger.Debug("mysql connected!")

	drv, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}
	// Get the underlying sql.DB object of the driver.
	db := drv.DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
	cli := ent.NewClient(
		ent.Driver(drv),
		ent.Log(func(a ...any) {
			logger.Debug(fmt.Sprint(a...))
		}),
	)
	if cfg.Debug {
		cli = cli.Debug()
	}
	return cli, nil
}

func NewS3(cfg *config.Config) (*minio.Client, error) {
	var (
		endpoint        = cfg.S3Endpoint
		accessKeyID     = cfg.S3AccessKeyID
		secretAccessKey = cfg.S3SecretAccessKey
		useSSL          = cfg.S3UseSSL
	)
	if !cfg.S3Enabled {
		return nil, nil
	}
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		return nil, errors.New("s3 config error")
	}

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return minioClient, nil
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

func NewK8sClient(cfg *config.Config, logger mlog.Logger) (*K8sClient, error) {
	var (
		config   *restclient.Config
		err      error
		nsPrefix = cfg.NsPrefix

		eventFanOutObj = &fanOut[*eventsv1.Event]{
			name:      "event",
			ch:        make(chan Obj[*eventsv1.Event], 100),
			listeners: make(map[string]chan<- Obj[*eventsv1.Event]),
		}

		podFanOutObj = &fanOut[*corev1.Pod]{
			name:      "pod",
			ch:        make(chan Obj[*corev1.Pod], 100),
			listeners: make(map[string]chan<- Obj[*corev1.Pod]),
		}
	)

	runtime.ErrorHandlers = []func(err error){
		func(err error) {
			logger.Warning(err)
		},
	}

	if cfg.KubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = restclient.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	// 客户端不限速，有可能会把集群打死。
	config.QPS = -1

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metrics, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, err
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
				podFanOutObj.ch <- NewObj[*corev1.Pod](nil, obj.(*corev1.Pod), Add)
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
				podFanOutObj.ch <- NewObj[*corev1.Pod](nil, obj.(*corev1.Pod), Delete)
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

	return &K8sClient{
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
	}, nil
}

func (k *K8sClient) Start(done <-chan struct{}) {
	go func() {
		defer k.logger.HandlePanic("[FANOUT]: event Distribute")

		k.EventFanOut.Distribute(done)
	}()
	go func() {
		defer k.logger.HandlePanic("[FANOUT]: pod Distribute")

		k.PodFanOut.Distribute(done)
	}()
	cache.WaitForCacheSync(nil, k.EventInformer.HasSynced, k.PodInformer.HasSynced, k.SecretInformer.HasSynced)
	k.factory.Start(done)
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
	AddListener(key string, ch chan<- Obj[T])
	Distribute(done <-chan struct{})
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

func (f *fanOut[T]) AddListener(key string, ch chan<- Obj[T]) {
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
	_, ok := f.listeners[key]
	if ok {
		delete(f.listeners, key)
		metrics.K8sInformerFanOutListenerCount.With(prometheus.Labels{"type": f.name}).Dec()
	}
}

func (f *fanOut[T]) Distribute(done <-chan struct{}) {
	defer f.logger.Debug(fmt.Sprintf("[FANOUT]: '%s' Exit", f.name))
	if !f.started.start() {
		return
	}
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
