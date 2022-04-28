import ajax from "./ajax";
import pb from "./compiled"

export function detailProject(projectId:number) {
  return ajax.get<pb.project.ShowResponse>(`/api/projects/${projectId}`);
}

export function deleteProject(projectId:number) {
  return ajax.delete<pb.project.DeleteResponse>(`/api/projects/${projectId}`);
}

export function allPodContainers({project_id}: pb.project.AllContainersRequest) {
  return ajax.get<pb.project.AllContainersResponse>(`/api/projects/${project_id}/containers`);
}

export function containerLog({pod, namespace, container}: pb.container.LogRequest) {
  return ajax.get<pb.container.LogResponse>(`/api/containers/namespaces/${namespace}/pods/${pod}/containers/${container}/logs`);
}

export function isPodRunning({namespace, pod}: pb.container.IsPodRunningRequest) {
  return ajax.post<pb.container.IsPodRunningResponse>(`/api/containers/pod_running_status`, {namespace, pod});
}

export function isPodExists({namespace, pod}: pb.container.IsPodExistsRequest) {
  return ajax.post<pb.container.IsPodExistsResponse>(`/api/containers/pod_exists`, {namespace, pod});
}