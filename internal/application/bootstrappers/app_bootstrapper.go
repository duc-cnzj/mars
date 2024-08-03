package bootstrappers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/util"
	"github.com/duc-cnzj/mars/v4/internal/util/mars"
	"github.com/duc-cnzj/mars/v4/plugins/domainmanager"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/namespace"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type AppBootstrapper struct{}

func (a *AppBootstrapper) Tags() []string {
	return []string{}
}

func (a *AppBootstrapper) Bootstrap(app application.App) error {
	app.BeforeServerRunHooks(blockForFuncForever("projectPodEventListener", projectPodEventListener))
	app.BeforeServerRunHooks(lockFunc("updateCerts", updateCerts))
	app.BeforeServerRunHooks(lockFunc("syncImagePullSecrets", syncImagePullSecrets))

	return nil
}

func blockForFuncForever(title string, cb application.Callback) func(app application.App) {
	return func(appli application.App) {
		go func() {
			blockLockFunc(appli, title, func(releaseFn func()) {
				appli.RegisterAfterShutdownFunc(func(app application.App) {
					releaseFn()
				})
				cb(appli)
			}, 5*time.Second, 180, 150, appli.Done())
		}()
	}
}

func lockFunc(key string, callback application.Callback) application.Callback {
	return func(app application.App) {
		releaseFn, acquired := app.Locker().RenewalAcquire(key, 180, 150)
		if !acquired {
			return
		}
		defer releaseFn()
		callback(app)
	}
}

func blockLockFunc(app application.App, key string, fn func(releaseFn func()), tickerDuration time.Duration, seconds, renewalSeconds int64, done <-chan struct{}) {
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()
	var (
		acquired  bool
		releaseFn func()
	)
Loop:
	for {
		select {
		case <-ticker.C:
			releaseFn, acquired = app.Locker().RenewalAcquire(key, seconds, renewalSeconds)
			if acquired {
				break Loop
			}
		case <-done:
			return
		}
	}
	fn(releaseFn)
}

func updateCerts(app application.App) {
	name, key, crt := app.PluginMgr().Domain().GetCerts()
	domainmanager.UpdateCertTls(app.DB(), app.Data().K8sClient, app.Logger(), name, key, crt)
}

// syncImagePullSecrets
// 少了自动加上，更新了自动修改
// 自动同步包括
// 1. db image_pull_secrets 丢失(不会自动删除之前的 secret)
// 2. k8s secrets 丢失
// 3. 密码更新
// 4. 删除 config
// 4. 新增 config
func syncImagePullSecrets(app application.App) {
	var (
		namespaceList       []*ent.Namespace
		cfgImagePullSecrets = app.Config().ImagePullSecrets
		k8sClient           = app.Data().K8sClient
	)
	var serverMap = make(map[string]util.DockerConfigEntry)
	for _, s := range cfgImagePullSecrets {
		serverMap[s.Server] = util.DockerConfigEntry{
			Username: s.Username,
			Password: s.Password,
			Email:    s.Email,
			Auth:     base64.StdEncoding.EncodeToString([]byte(s.Username + ":" + s.Password)),
		}
	}
	namespaceList, _ = app.DB().Namespace.Query().Select(
		namespace.FieldID,
		namespace.FieldName,
		namespace.FieldImagePullSecrets,
	).All(context.TODO())
	for _, namespace := range namespaceList {
		var (
			checked = make(map[string]struct{})
			missing config.DockerAuths
			ns      = namespace
		)
		for _, secretName := range ns.ImagePullSecrets {
			secret, err := k8sClient.SecretLister.Secrets(ns.Name).Get(secretName)
			if err != nil {
				app.Logger().Warningf("[syncImagePullSecrets]: error get secret '%s', err %v", secretName, err)
				if apierrors.IsNotFound(err) {
					deleteSecret(app.Logger(), k8sClient.Client, ns, secretName)
				}
				continue
			}
			if secret.Type == v1.SecretTypeDockerConfigJson {
				var dockerJsonKeyData []byte = secret.Data[v1.DockerConfigJsonKey]
				res, err := util.DecodeDockerConfigJSON(dockerJsonKeyData)
				if err != nil {
					app.Logger().Warningf("[syncImagePullSecrets]: decode secret '%s', err %v", secretName, err)
					continue
				}
				var newConfigJson = util.DockerConfigJSON{
					Auths:       map[string]util.DockerConfigEntry{},
					HttpHeaders: map[string]string{},
				}
				for server, cfg := range serverMap {
					for s := range res.Auths {
						if server == s {
							newConfigJson.Auths[server] = cfg
							checked[server] = struct{}{}
							break
						}
					}
				}
				if len(newConfigJson.Auths) == 0 {
					deleteSecret(app.Logger(), k8sClient.Client, ns, secretName)
					continue
				}

				if !reflect.DeepEqual(newConfigJson.Auths, res.Auths) {
					app.Logger().Warningf("[syncImagePullSecrets]: Find Diff, Auto Sync: '%s'", secretName)
					marshal, _ := json.Marshal(&newConfigJson)
					secret.Data[v1.DockerConfigJsonKey] = marshal
					k8sClient.Client.CoreV1().Secrets(ns.Name).Update(context.TODO(), secret, metav1.UpdateOptions{})
				}
			}
		}

		for s, cfg := range serverMap {
			if _, ok := checked[s]; !ok {
				missing = append(missing, &config.DockerAuth{
					Username: cfg.Username,
					Password: cfg.Password,
					Email:    cfg.Email,
					Server:   s,
				})
			}
		}

		if len(missing) > 0 {
			secret, err := util.CreateDockerSecrets(k8sClient.Client, ns.Name, missing)
			if err == nil {
				app.Logger().Warningf("[syncImagePullSecrets]: Missing %v", missing)

				ns.Update().SetImagePullSecrets(append(ns.ImagePullSecrets, secret.Name)).SaveX(context.TODO())
			}
		}
	}
}

func deleteSecret(logger mlog.Logger, client kubernetes.Interface, ns *ent.Namespace, secretName string) {
	logger.Warningf("[syncImagePullSecrets]: DELETE: %s", secretName)

	client.CoreV1().Secrets(ns.Name).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	var newNsArray []string
	for _, name := range ns.ImagePullSecrets {
		if name != secretName {
			newNsArray = append(newNsArray, name)
		}
	}
	ns.Update().SetImagePullSecrets(newNsArray).SaveX(context.TODO())
}

func projectPodEventListener(app application.App) {
	var cfg = app.Config()
	ch := make(chan data.Obj[*v1.Pod], 100)
	listener := "pod-watcher"
	namespacePublisher := app.PluginMgr().Ws().New("", "").(application.ProjectPodEventPublisher)
	podFanOut := app.Data().K8sClient.PodFanOut
	podFanOut.AddListener(listener, ch)

	go func() {
		defer app.Logger().HandlePanic(listener)
		defer func() {
			app.Logger().Debug("[PodEventListener]: pod-watcher exit")
			podFanOut.RemoveListener(listener)
			close(ch)
		}()

		for {
			select {
			case <-app.Done():
				return
			case obj, ok := <-ch:
				if !ok {
					return
				}
				switch obj.Type() {
				case data.Update:
					if obj.Old().Status.Phase != obj.Current().Status.Phase || containerStatusChanged(app.Logger(), obj.Old(), obj.Current()) {
						app.Logger().Debugf("old: '%s' new '%s'", obj.Old().Status.Phase, obj.Current().Status.Phase)
						if ns, err := app.DB().Namespace.Query().Where(namespace.NameEQ(mars.GetMarsNamespace(obj.Current().Namespace, cfg.NsPrefix))).Only(context.TODO()); err == nil {
							if err := namespacePublisher.Publish(int64(ns.ID), obj.Current()); err != nil {
								app.Logger().Errorf("[PodEventListener]: %v", err)
							}
						}
					}
				case data.Add:
					fallthrough
				case data.Delete:
					if ns, err := app.DB().Namespace.Query().Where(namespace.NameEQ(mars.GetMarsNamespace(obj.Current().Namespace, cfg.NsPrefix))).Only(context.TODO()); err == nil {
						app.Logger().Debugf("[PodEventListener]: pod '%v': '%s' '%s' '%d' '%s'", obj.Type(), obj.Current().Name, obj.Current().Namespace, ns.ID, obj.Current().Status.Phase)
						if err := namespacePublisher.Publish(int64(ns.ID), obj.Current()); err != nil {
							app.Logger().Errorf("[PodEventListener]: %v", err)
						}
					}
				default:
				}
			}
		}
	}()
}

type watchContainerStatus struct {
	Ready bool
}

func containerStatusChanged(logger mlog.Logger, old *v1.Pod, current *v1.Pod) bool {
	if len(old.Status.ContainerStatuses) != len(current.Status.ContainerStatuses) {
		return true
	}
	var oldMap = map[string]watchContainerStatus{}
	for _, status := range old.Status.ContainerStatuses {
		oldMap[status.Name] = watchContainerStatus{
			Ready: status.Ready,
		}
	}
	var currentMap = map[string]watchContainerStatus{}
	for _, status := range current.Status.ContainerStatuses {
		currentMap[status.Name] = watchContainerStatus{
			Ready: status.Ready,
		}
	}

	for k, v := range currentMap {
		if b, ok := oldMap[k]; !(ok && b == v) {
			logger.Debugf("ContainerStatus old: %v current: %v", b, v)
			return true
		}
	}

	return false
}
