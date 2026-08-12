package eventhandler

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	corev1 "k8s.io/api/core/v1"
)

// PodEventPublisher 发布项目 pod 事件。窄接口定义在消费方（本包），
// 实现由组合根（cmd/serve.go）惰性适配 PluginManager.Ws() 注入。
type PodEventPublisher interface {
	// Publish 发布某个项目的 pod 变更事件。
	Publish(nsID int64, pod *corev1.Pod) error
}

// PodEventListener 是 Pod 生命周期事件监听者（常驻用例，非定时任务）：
// 订阅 informer 的 Pod 事件，把状态变化/新增/删除事件通过 PodEventPublisher
// 发布出去。订阅与转换收敛在 K8sRepo 端口内，此处只做领域层消费。
//
// 原实现伪装成定时任务（cron 每 5 秒借锁触发一次、阻塞消费实现单例常驻），
// 归位为常驻 server 后随 app 生命周期启停：Run 阻塞消费直至 ctx 取消，退出前
// 注销订阅；由组合根包装成 app.Server 注册启动。
type PodEventListener struct {
	logger  mlog.Logger
	k8sRepo biz.K8sRepo
	nsRepo  biz.NamespaceRepo
	pub     PodEventPublisher
}

// NewPodEventListener 构造常驻 Pod 事件监听者。
func NewPodEventListener(logger mlog.Logger, k8sRepo biz.K8sRepo, nsRepo biz.NamespaceRepo, pub PodEventPublisher) *PodEventListener {
	return &PodEventListener{
		logger:  logger.WithModule("event/pod-listener"),
		k8sRepo: k8sRepo,
		nsRepo:  nsRepo,
		pub:     pub,
	}
}

// Run 常驻消费 informer 的 Pod 事件直至 ctx 取消，退出前注销订阅并返回 nil。
// 阻塞语义与 app.Server 的异步契约由组合根适配器（cmd）桥接。
func (p *PodEventListener) Run(ctx context.Context) error {
	ch, unsubscribe := p.k8sRepo.SubscribePodEvents("pod-watcher")
	defer unsubscribe()
	defer p.logger.HandlePanic("pod-watcher")

	p.logger.Info("[PodEventListener]: started")
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("[PodEventListener]: stopped")
			return nil
		case ev, ok := <-ch:
			if !ok {
				p.logger.Warning("[PodEventListener]: channel closed")
				return nil
			}
			p.handle(ev)
		}
	}
}

// handle 消费单条 Pod 事件：更新时对比新旧状态相位/容器就绪变化，
// 有变化才经 nsRepo 解析命名空间后发布；新增/删除直接发布。namespace 解析
// 失败只记 Debug 日志（无权限/已删除的 namespace 不应中断监听）。
func (p *PodEventListener) handle(ev biz.PodEvent) {
	switch ev.Type {
	case biz.PodEventUpdate:
		p.logger.Debug("[#### PodEventListener]: update pod", ev.Current.Name, ev.Current.Namespace)
		if ev.Old.Status.Phase != ev.Current.Status.Phase || containerStatusChanged(p.logger, ev.Old, ev.Current) {
			p.logger.Debugf("old: '%s' new '%s'", ev.Old.Status.Phase, ev.Current.Status.Phase)
			if ns, err := p.nsRepo.FindByName(context.TODO(), ev.Current.Namespace); err == nil {
				if err := p.pub.Publish(int64(ns.ID), ev.Current); err != nil {
					p.logger.Errorf("[PodEventListener]: %v", err)
				}
			}
		}
	case biz.PodEventAdd, biz.PodEventDelete:
		p.logger.Debug("[PodEventListener]: add/del pod", ev.Type, ev.Current.Name, ev.Current.Namespace)
		if ns, err := p.nsRepo.FindByName(context.TODO(), ev.Current.Namespace); err == nil {
			p.logger.Debugf("[PodEventListener]: pod '%v': '%s' '%s' '%d' '%s'", ev.Type, ev.Current.Name, ev.Current.Namespace, ns.ID, ev.Current.Status.Phase)
			if err := p.pub.Publish(int64(ns.ID), ev.Current); err != nil {
				p.logger.Errorf("[PodEventListener]: %v", err)
			}
		}
	default:
	}
}

type watchContainerStatus struct {
	Ready bool
}

// containerStatusChanged 对比新旧 pod 的容器状态：容器数量不一致或任一
// 容器的 Ready 状态翻转即视为变化。供 handle 判定更新事件是否值得发布。
func containerStatusChanged(logger mlog.Logger, old *corev1.Pod, current *corev1.Pod) bool {
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
		if b, ok := oldMap[k]; !ok || b != v {
			logger.Debugf("ContainerStatus old: %v current: %v", b, v)
			return true
		}
	}

	return false
}
