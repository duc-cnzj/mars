package bootstrappers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
	"github.com/duc-cnzj/mars/v4/internal/utils/tls"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type AppBootstrapper struct{}

func (a *AppBootstrapper) Tags() []string {
	return []string{}
}

func (a *AppBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	// 预加载插件
	plugins.GetWsSender()
	plugins.GetPicture()
	plugins.GetGitServer()
	plugins.GetDomainManager()

	app.BeforeServerRunHooks(blockForFuncForever("projectPodEventListener", projectPodEventListener))
	app.BeforeServerRunHooks(lockFunc("updateCerts", updateCerts))
	app.BeforeServerRunHooks(lockFunc("syncImagePullSecrets", syncImagePullSecrets))

	return nil
}

func blockForFuncForever(title string, cb contracts.Callback) func(app contracts.ApplicationInterface) {
	return func(app contracts.ApplicationInterface) {
		go func() {
			blockLockFunc(app, title, func(releaseFn func()) {
				app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
					releaseFn()
				})
				cb(app)
			}, 5*time.Second, 180, 150, app.Done())
		}()
	}
}

func lockFunc(key string, callback contracts.Callback) contracts.Callback {
	return func(app contracts.ApplicationInterface) {
		releaseFn, acquired := app.CacheLock().RenewalAcquire(key, 180, 150)
		if !acquired {
			return
		}
		defer releaseFn()
		callback(app)
	}
}

func blockLockFunc(app contracts.ApplicationInterface, key string, fn func(releaseFn func()), tickerDuration time.Duration, seconds, renewalSeconds int64, done <-chan struct{}) {
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
			releaseFn, acquired = app.CacheLock().RenewalAcquire(key, seconds, renewalSeconds)
			if acquired {
				break Loop
			}
		case <-done:
			return
		}
	}
	fn(releaseFn)
}

func updateCerts(app contracts.ApplicationInterface) {
	tls.UpdateCertTls(plugins.GetDomainManager().GetCerts())
}

// syncImagePullSecrets
// 少了自动加上，更新了自动修改
// 自动同步包括
// 1. db image_pull_secrets 丢失(不会自动删除之前的 secret)
// 2. k8s secrets 丢失
// 3. 密码更新
// 4. 删除 config
// 4. 新增 config
func syncImagePullSecrets(app contracts.ApplicationInterface) {
	var (
		namespaceList       []models.Namespace
		cfgImagePullSecrets = app.Config().ImagePullSecrets
		k8sClient           = app.K8sClient()
	)
	var serverMap = make(map[string]utils.DockerConfigEntry)
	for _, s := range cfgImagePullSecrets {
		serverMap[s.Server] = utils.DockerConfigEntry{
			Username: s.Username,
			Password: s.Password,
			Email:    s.Email,
			Auth:     base64.StdEncoding.EncodeToString([]byte(s.Username + ":" + s.Password)),
		}
	}
	app.DB().Select("ID", "Name", "ImagePullSecrets").Find(&namespaceList)
	for _, namespace := range namespaceList {
		var (
			checked = make(map[string]struct{})
			missing config.DockerAuths
			ns      = namespace
		)
		for _, secretName := range ns.ImagePullSecretsArray() {
			secret, err := k8sClient.SecretLister.Secrets(ns.Name).Get(secretName)
			if err != nil {
				mlog.Warningf("[syncImagePullSecrets]: error get secret '%s', err %v", secretName, err)
				if apierrors.IsNotFound(err) {
					deleteSecret(app, k8sClient.Client, &ns, secretName)
				}
				continue
			}
			if secret.Type == v1.SecretTypeDockerConfigJson {
				var dockerJsonKeyData []byte = secret.Data[v1.DockerConfigJsonKey]
				res, err := utils.DecodeDockerConfigJSON(dockerJsonKeyData)
				if err != nil {
					mlog.Warningf("[syncImagePullSecrets]: decode secret '%s', err %v", secretName, err)
					continue
				}
				var newConfigJson = utils.DockerConfigJSON{
					Auths:       map[string]utils.DockerConfigEntry{},
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
					deleteSecret(app, k8sClient.Client, &ns, secretName)
					continue
				}

				if !reflect.DeepEqual(newConfigJson.Auths, res.Auths) {
					mlog.Warningf("[syncImagePullSecrets]: Find Diff, Auto Sync: '%s'", secretName)
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
			secret, err := utils.CreateDockerSecrets(k8sClient.Client, ns.Name, missing)
			if err == nil {
				mlog.Warningf("[syncImagePullSecrets]: Missing %v", missing)

				app.DB().Model(&ns).Updates(map[string]any{
					"image_pull_secrets": strings.Join(append(ns.ImagePullSecretsArray(), secret.Name), ","),
				})
			}
		}
	}
}

func deleteSecret(app contracts.ApplicationInterface, client kubernetes.Interface, ns *models.Namespace, secretName string) {
	mlog.Warningf("[syncImagePullSecrets]: DELETE: %s", secretName)

	client.CoreV1().Secrets(ns.Name).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	var newNsArray []string
	for _, name := range ns.ImagePullSecretsArray() {
		if name != secretName {
			newNsArray = append(newNsArray, name)
		}
	}
	app.DB().Model(&ns).Updates(map[string]any{
		"image_pull_secrets": strings.Join(newNsArray, ","),
	})
}

func projectPodEventListener(app contracts.ApplicationInterface) {
	ch := make(chan contracts.Obj[*v1.Pod], 100)
	listener := "pod-watcher"
	namespacePublisher := plugins.GetWsSender().New("", "").(contracts.ProjectPodEventPublisher)
	podFanOut := app.K8sClient().PodFanOut
	podFanOut.AddListener(listener, ch)

	go func() {
		defer recovery.HandlePanic(listener)
		defer func() {
			mlog.Debug("[PodEventListener]: pod-watcher exit")
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
				case contracts.Update:
					if obj.Old().Status.Phase != obj.Current().Status.Phase || containerStatusChanged(obj.Old(), obj.Current()) {
						mlog.Debugf("old: '%s' new '%s'", obj.Old().Status.Phase, obj.Current().Status.Phase)
						var ns models.Namespace
						if app.DB().Where("`name` = ?", utils.GetMarsNamespace(obj.Current().Namespace)).First(&ns).Error == nil {
							if err := namespacePublisher.Publish(int64(ns.ID), obj.Current()); err != nil {
								mlog.Errorf("[PodEventListener]: %v", err)
							}
						}
					}
				case contracts.Add:
					fallthrough
				case contracts.Delete:
					var ns models.Namespace
					if app.DB().Where("`name` = ?", utils.GetMarsNamespace(obj.Current().Namespace)).First(&ns).Error == nil {
						mlog.Debugf("[PodEventListener]: pod '%v': '%s' '%s' '%d' '%s'", obj.Type(), obj.Current().Name, obj.Current().Namespace, ns.ID, obj.Current().Status.Phase)
						if err := namespacePublisher.Publish(int64(ns.ID), obj.Current()); err != nil {
							mlog.Errorf("[PodEventListener]: %v", err)
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

func containerStatusChanged(old *v1.Pod, current *v1.Pod) bool {
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
			mlog.Debugf("ContainerStatus old: %v current: %v", b, v)
			return true
		}
	}

	return false
}
