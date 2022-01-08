import ajax from "./ajax";
import pb from "./compiled"

export function detailProject(projectId:number) {
  return ajax.get<pb.ProjectShowResponse>(`/api/projects/${projectId}`);
}

export function deleteProject(projectId:number) {
  return ajax.delete<pb.ProjectDeleteResponse>(`/api/projects/${projectId}`);
}

export function allPodContainers({project_id}: pb.ProjectAllPodContainersRequest) {
  return ajax.get<pb.ProjectAllPodContainersResponse>(`/api/projects/${project_id}/containers`);
}

export function containerLog({pod, project_id, container}: pb.ProjectPodContainerLogRequest) {
  return ajax.get<pb.ProjectPodContainerLogResponse>(`/api/projects/${project_id}/pods/${pod}/containers/${container}/logs`);
}

export function isPodRunning({namespace, pod}: pb.ProjectIsPodRunningRequest) {
  return ajax.get<pb.ProjectIsPodRunningResponse>(`/api/namespaces/${namespace}/pod/${pod}/status`);
}


