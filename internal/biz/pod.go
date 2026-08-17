package biz

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/spf13/cast"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// allStatus 定义 Pod 阶段的排序权重，数值小的排前面。
var allStatus = map[corev1.PodPhase]int{
	corev1.PodRunning:   1,
	corev1.PodSucceeded: 2,
	corev1.PodFailed:    3,
	corev1.PodPending:   4,
	corev1.PodUnknown:   5,
}

// SortStatePod 按展示顺序对 StatePod 排序：阶段权重优先，其次新/旧、OrderIndex、
// Terminating、创建时间。
type SortStatePod []StatePod

// Len 返回 pod 个数。
func (s SortStatePod) Len() int {
	return len(s)
}

// Less 判定第 i 个 pod 应排在第 j 个之前：按阶段权重升序，同阶段时旧版本 pod
// 排后、OrderIndex 大的排后、Terminating 的排后，最后按创建时间升序兜底。
func (s SortStatePod) Less(i, j int) bool {
	if allStatus[s[i].Pod.Status.Phase] < allStatus[s[j].Pod.Status.Phase] {
		return true
	}

	if s[i].Pod.Status.Phase == s[j].Pod.Status.Phase {
		if !s[i].IsOld && s[j].IsOld {
			return true
		}

		if s[i].OrderIndex > s[j].OrderIndex && s[i].IsOld == s[j].IsOld {
			return true
		}

		if !s[i].Terminating && s[j].Terminating && s[i].IsOld == s[j].IsOld {
			return true
		}

		if s[i].OrderIndex == s[j].OrderIndex && s[i].IsOld == s[j].IsOld && s[i].Terminating == s[j].Terminating {
			return s[i].Pod.CreationTimestamp.Time.Before(s[j].Pod.CreationTimestamp.Time)
		}
	}

	return false
}

// Swap 交换两个 pod 的位置。
func (s SortStatePod) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// isContainerReady 判断容器是否 Ready（取第一个匹配的 ContainerStatus）。
func isContainerReady(pod *corev1.Pod, containerName string) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			return containerStatus.Ready
		}
	}
	return false
}

// buildStateContainers 从项目的 PodSelectors 出发推导出活跃容器的状态列表。
// 领域逻辑：Deployment 通过 ReplicaSet 的 deployment revision 注解识别滚动发布中的
// 旧版本副本；StatefulSet/DaemonSet 按 controller-revision-hash 识别旧版本副本；
// 并过滤掉标注了 IgnoreContainerNames 的 sidecar 容器。
func buildStateContainers(ctx context.Context, k8sRepo K8sRepo, proj *Project) ([]*types.StateContainer, error) {
	if len(proj.PodSelectors) == 0 {
		return nil, nil
	}
	pods, err := k8sRepo.ListPodsBySelectors(proj.Namespace.Name, proj.PodSelectors)
	if err != nil {
		return nil, err
	}
	var list = make(map[string]*corev1.Pod)
	for _, pod := range pods {
		if pod.Status.Phase != corev1.PodFailed {
			list[pod.Name] = pod
		}
	}

	var m = make(map[string]*appsv1.ReplicaSet)
	var objectMap = make(map[string]runtime.Object)
	var oldReplicaMap = make(map[string]struct{})

	for _, pod := range list {
		for _, reference := range pod.OwnerReferences {
			if reference.Kind == "ReplicaSet" {
				var (
					rs  *appsv1.ReplicaSet
					err error
					ok  bool
				)
				if _, ok = m[string(reference.UID)]; !ok {
					rs, err = k8sRepo.GetReplicaSet(pod.Namespace, reference.Name)
					if err != nil {
						continue
					}
					m[string(reference.UID)] = rs
					for _, re := range rs.OwnerReferences {
						if re.Kind == "Deployment" {
							uniqueKey := string(re.UID)
							if old, found := objectMap[uniqueKey]; found {
								accessor1, _ := meta.Accessor(old)
								accessor2, _ := meta.Accessor(rs)
								accessor1Revision := accessor1.GetAnnotations()[RevisionAnnotation]
								accessor2Revision := accessor2.GetAnnotations()[RevisionAnnotation]
								if accessor1Revision != "" && accessor2Revision != "" && accessor1Revision != accessor2Revision {
									// Deployment revision 注解恒为十进制数字；字符串比较会把 "10" < "2" 判为真，
									// 导致 revision ≥10 时新旧判定错乱，这里显式转数字比较，解析失败退化为字符串比较。
									older := accessor1Revision < accessor2Revision
									if rev1, err1 := strconv.Atoi(accessor1Revision); err1 == nil {
										if rev2, err2 := strconv.Atoi(accessor2Revision); err2 == nil {
											older = rev1 < rev2
										}
									}
									if older {
										oldReplicaMap[string(accessor1.GetUID())] = struct{}{}
										objectMap[uniqueKey] = rs
									} else {
										oldReplicaMap[string(accessor2.GetUID())] = struct{}{}
									}
								}
							} else {
								objectMap[uniqueKey] = rs
							}
							break
						}
					}
				}
			}
		}
	}

	// 收集 StatefulSet/DaemonSet 的旧版本 pod：其 pod 无 ReplicaSet 属主，无法走
	// 上面的 revision 判定，需按 controller-revision-hash 与状态值单独识别。
	var podList []*corev1.Pod
	for _, pod := range list {
		podList = append(podList, pod)
	}
	oldPodNames := collectWorkloadOldPods(k8sRepo, podList)

	var newList SortStatePod
	for _, pod := range list {
		var isOld bool
		if _, ok := oldPodNames[pod.Name]; ok {
			isOld = true
		}
		for _, reference := range pod.OwnerReferences {
			if _, ok := oldReplicaMap[string(reference.UID)]; ok {
				isOld = true
				break
			}
		}

		idx := pod.Annotations[PodOrderIndex]

		newList = append(newList, StatePod{
			IsOld:       isOld,
			Terminating: pod.DeletionTimestamp != nil,
			Pending:     pod.Status.Phase == corev1.PodPending,
			OrderIndex:  cast.ToInt(idx),
			Pod:         pod.DeepCopy(),
		})
	}

	sort.Sort(newList)

	var containerList []*types.StateContainer
	for _, item := range newList {
		var ignores = make(map[string]struct{})
		if s, ok := item.Pod.Annotations[IgnoreContainerNames]; ok {
			split := strings.Split(s, ",")
			for _, sp := range split {
				ignores[strings.TrimSpace(sp)] = struct{}{}
			}
		}
		for _, c := range item.Pod.Spec.Containers {
			if _, found := ignores[c.Name]; found {
				continue
			}
			containerList = append(containerList,
				&types.StateContainer{
					Namespace:   proj.Namespace.Name,
					Pod:         item.Pod.Name,
					Container:   c.Name,
					IsOld:       item.IsOld,
					Terminating: item.Terminating,
					Pending:     item.Pending,
					Ready:       isContainerReady(item.Pod, c.Name),
				},
			)
		}
	}

	return containerList, nil
}

// collectWorkloadOldPods 识别 StatefulSet/DaemonSet 的旧版本 pod 名集合。
// StatefulSet：controller-revision-hash != status.updateRevision 视为旧副本；
// DaemonSet：状态无 revision 字段，按 controller-revision-hash 分组，取创建时间
// 最新的一组为新版本，其余视为旧副本。读取失败或未滚动的负载不标记旧（保守不误标）。
func collectWorkloadOldPods(k8sRepo K8sRepo, pods []*corev1.Pod) map[string]struct{} {
	oldPods := make(map[string]struct{})
	stsGroups := make(map[string][]*corev1.Pod)
	dsGroups := make(map[string][]*corev1.Pod)
	for _, pod := range pods {
		for _, ref := range pod.OwnerReferences {
			switch ref.Kind {
			case "StatefulSet":
				stsGroups[pod.Namespace+"/"+ref.Name] = append(stsGroups[pod.Namespace+"/"+ref.Name], pod)
			case "DaemonSet":
				dsGroups[pod.Namespace+"/"+ref.Name] = append(dsGroups[pod.Namespace+"/"+ref.Name], pod)
			}
		}
	}
	for key, group := range stsGroups {
		parts := strings.SplitN(key, "/", 2)
		sts, err := k8sRepo.GetStatefulSet(parts[0], parts[1])
		if err != nil || sts.Status.UpdateRevision == "" {
			continue
		}
		for _, pod := range group {
			if pod.Labels[appsv1.ControllerRevisionHashLabelKey] != sts.Status.UpdateRevision {
				oldPods[pod.Name] = struct{}{}
			}
		}
	}
	for _, group := range dsGroups {
		// daemonSetNewPods 对非空 group 恒返回非空（无 hash 时兜底返回全部），无空集分支。
		newPods := daemonSetNewPods(group)
		newNames := make(map[string]struct{}, len(newPods))
		for _, pod := range newPods {
			newNames[pod.Name] = struct{}{}
		}
		for _, pod := range group {
			if _, ok := newNames[pod.Name]; !ok {
				oldPods[pod.Name] = struct{}{}
			}
		}
	}
	return oldPods
}
