import ajax from "./ajax";
export interface ProjectDetail {
  id: number;
  name: string;
  gitlab_project_id: string;
  gitlab_branch: string;
  gitlab_commit: string;
  config: string;
  namespace: {
    id: number;
    name: string;
  }

  cpu:string;
  memory:string;
  created_at:string;

  gitlab_commit_title:string;
  gitlab_commit_web_url:string;
  gitlab_commit_author:string;
}

export function detailProject(namespaceId:number, projectId:number) {
  return ajax.get<{data: ProjectDetail}>(`/api/namespaces/${namespaceId}/projects/${projectId}`);
}

export function deleteProject(namespaceId:number, projectId:number) {
  return ajax.delete(`/api/namespaces/${namespaceId}/projects/${projectId}`);
}

export interface PodContainerItem {
  pod_name: string;
  container_name: string;
  log?: string;
}

export function containerList(namespaceId:number, projectId:number) {
  return ajax.get<{data: PodContainerItem[]}>(`/api/namespaces/${namespaceId}/projects/${projectId}/containers`);
}

export function containerLog(namespaceId:number, projectId:number, podName:string, containerName:string) {
  return ajax.get<{data: PodContainerItem}>(`/api/namespaces/${namespaceId}/projects/${projectId}/pods/${podName}/containers/${containerName}/logs`);
}


