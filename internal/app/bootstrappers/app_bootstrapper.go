package bootstrappers

import (
	"context"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/internal/utils/recovery"
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
	app.BeforeServerRunHooks(UpdateTls)

	return nil
}

func UpdateTls(app contracts.ApplicationInterface) {
	// 需要更新 tls 证书
	name, key, crt := plugins.GetDomainManager().GetCerts()
	if name != "" && key != "" && crt != "" {
		var namespaceList []models.Namespace
		app.DB().Select("ID", "Name").Find(&namespaceList)
		var changed bool
		var changedSecrets []*v1.Secret
		for _, n := range namespaceList {
			secret, err := app.K8sClient().Client.CoreV1().Secrets(n.Name).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					utils.AddTlsSecret(n.Name, name, key, crt)
					continue
				}
			}
			if string(secret.Data["tls.crt"]) != crt || string(secret.Data["tls.key"]) != key {
				changed = true
				changedSecrets = append(changedSecrets, secret.DeepCopy())
				break
			}
		}
		if changed {
			sdata := map[string]string{
				"tls.key": key,
				"tls.crt": crt,
			}
			mlog.Warning("[TLS]: certs changed, updating...")
			for _, secret := range changedSecrets {
				secret.StringData = sdata
				_, err := app.K8sClient().Client.CoreV1().Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
				if err == nil {
					mlog.Infof("[TLS]: namespace: %s, name %s updated", secret.Namespace, secret.Name)
				}
			}
		}
	}
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
