package biz

// 以下常量是 mars 部署在 Kubernetes 上使用的注解（annotation）键，
// 统一收拢于此，避免在 util 下以零依赖常量的形式割裂存放。

// RevisionAnnotation 是 Deployment 滚动发布时标注 ReplicaSet 版本的注解键。
const RevisionAnnotation = "deployment.kubernetes.io/revision"

// IgnoreContainerNames 过滤 sidecar 这样的容器
const IgnoreContainerNames = "mars.duc-cnzj.github.io/ignore-containers"

// PodOrderIndex GetAllActiveContainers 排序时，index 高的在前面
const PodOrderIndex = "mars.duc-cnzj.github.io/order-index"
