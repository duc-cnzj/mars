import ajax from "./ajax";
import pb from "./compiled"

export function detailProject(namespaceId:number, projectId:number) {
  return ajax.get<pb.ProjectShowResponse>(`/api/namespaces/${namespaceId}/projects/${projectId}`);
}

export function deleteProject(namespaceId:number, projectId:number) {
  return ajax.delete(`/api/namespaces/${namespaceId}/projects/${projectId}`);
}

export function containerList({namespace_id, project_id}: pb.AllPodContainersRequest) {
  return ajax.get<pb.PodContainerLogResponse>(`/api/namespaces/${namespace_id}/projects/${project_id}/containers`);
}

export function containerLog({namespace_id, pod, project_id, container}: pb.PodContainerLogRequest) {
  return ajax.get<pb.PodContainerLogResponse>(`/api/namespaces/${namespace_id}/projects/${project_id}/pods/${pod}/containers/${container}/logs`);
}


