import ajax from "./ajax";
import pb from "./compiled"

export function detailProject(projectId:number) {
  return ajax.get<pb.ProjectShowResponse>(`/api/projects/${projectId}`);
}

export function deleteProject(projectId:number) {
  return ajax.delete<pb.ProjectDeleteResponse>(`/api/projects/${projectId}`);
}

export function allPodContainers({project_id}: pb.ProjectAllContainersRequest) {
  return ajax.get<pb.ProjectAllContainersResponse>(`/api/projects/${project_id}/containers`);
}

export function containerLog({pod, namespace, container}: pb.ContainerLogRequest) {
  return ajax.get<pb.ContainerLogResponse>(`/api/containers/namespaces/${namespace}/pods/${pod}/containers/${container}/logs`);
}

export function isPodRunning({namespace, pod}: pb.ContainerIsPodRunningRequest) {
  return ajax.post<pb.ContainerIsPodRunningResponse>(`/api/containers/pod_running_status`, {namespace, pod});
}

export function isPodExists({namespace, pod}: pb.ContainerIsPodExistsRequest) {
  return ajax.post<pb.ContainerIsPodExistsResponse>(`/api/containers/pod_exists`, {namespace, pod});
}