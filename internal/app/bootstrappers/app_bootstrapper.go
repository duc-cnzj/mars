package bootstrappers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/internal/utils/recovery"
	"github.com/duc-cnzj/mars/internal/utils/tls"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	app.BeforeServerRunHooks(ProjectPodEventListener)
	app.BeforeServerRunHooks(func(app contracts.ApplicationInterface) { tls.UpdateCertTls(plugins.GetDomainManager().GetCerts()) })
	app.BeforeServerRunHooks(SyncImagePullSecrets)

	return nil
}

// SyncImagePullSecrets
// 少了自动加上，更新了自动修改
// 自动同步包括
// 1. db image_pull_secrets 丢失(不会自动删除之前的 secret)
// 2. k8s secrets 丢失
// 3. 密码更新
// 4. 删除 config
// 4. 新增 config
func SyncImagePullSecrets(app contracts.ApplicationInterface) {
	var (
		namespaceList       []models.Namespace
		cfgImagePullSecrets = app.Config().ImagePullSecrets
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
	for _, ns := range namespaceList {
		var checked = make(map[string]struct{})
		var missing config.DockerAuths

		for _, secretName := range ns.ImagePullSecretsArray() {
			secret, err := app.K8sClient().SecretLister.Secrets(ns.Name).Get(secretName)
			if err != nil {
				mlog.Warningf("[SyncImagePullSecrets]: error get secret '%s', err %v", secretName, err)
				if apierrors.IsNotFound(err) {
					deleteSecret(app, &ns, secretName)
				}
				continue
			}
			if secret.Type == v1.SecretTypeDockerConfigJson {
				var dockerJsonKeyData []byte = secret.Data[v1.DockerConfigJsonKey]
				res, err := utils.DecodeDockerConfigJSON(dockerJsonKeyData)
				if err != nil {
					mlog.Warningf("[SyncImagePullSecrets]: decode secret '%s', err %v", secretName, err)
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
					deleteSecret(app, &ns, secretName)
					continue
				}

				if !reflect.DeepEqual(newConfigJson.Auths, res.Auths) {
					mlog.Warningf("[SyncImagePullSecrets]: Find Diff, Auto Sync: '%s'", secretName)
					marshal, _ := json.Marshal(&newConfigJson)
					secret.Data[v1.DockerConfigJsonKey] = marshal
					app.K8sClient().Client.CoreV1().Secrets(ns.Name).Update(context.TODO(), secret, metav1.UpdateOptions{})
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
			secret, err := utils.CreateDockerSecrets(app.K8sClient().Client, ns.Name, missing)
			if err == nil {
				mlog.Warningf("[SyncImagePullSecrets]: Missing %v", missing)

				app.DB().Model(&ns).Updates(map[string]any{
					"image_pull_secrets": strings.Join(append(ns.ImagePullSecretsArray(), secret.Name), ","),
				})
			}
		}
	}
}

func deleteSecret(app contracts.ApplicationInterface, ns *models.Namespace, secretName string) {
	mlog.Warningf("[SyncImagePullSecrets]: DELETE: %s", secretName)

	app.K8sClient().Client.CoreV1().Secrets(ns.Name).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
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

func ProjectPodEventListener(app contracts.ApplicationInterface) {
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
					if obj.Old().Status.Phase != obj.Current().Status.Phase {
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
