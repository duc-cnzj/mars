import ajax from "./ajax";

export interface Options {
  value: string;
  label: string;
  type: string;
  isLeaf: boolean;

  children?: Options[];
  projectId?: number;
  branch?: string;
}
export interface Info {
  id: number;
  name: string;
  path: string;
  web_url: string;
  avatar_url: string;
  description: string;
  enabled: boolean;
  global_enabled: boolean;
}

export function projectList() {
  return ajax.get<{ data: Info[] }>("/api/gitlab/project_list");
}

export function projects() {
  return ajax.get<{ data: Options[] }>("/api/gitlab/projects");
}

export function branches(id: number) {
  return ajax.get<{ data: Options[] }>(`/api/gitlab/projects/${id}/branches`);
}

export function commits(id: number, branch: string) {
  return ajax.get<{ data: Options[] }>(
    `/api/gitlab/projects/${id}/branches/${branch}/commits`
  );
}

export function commit(id: number, branch: string, commit: string) {
  return ajax.get<{ data: Options }>(
    `/api/gitlab/projects/${id}/branches/${branch}/commits/${commit}`
  );
}

export function pipelineInfo(id: number, branch: string, commit: string) {
  return ajax.get<{ data: { status: string; web_url: string } }>(
    `/api/gitlab/projects/${id}/branches/${branch}/commits/${commit}/pipeline_info`
  );
}

export function configFile(id: number, branch: string) {
  return ajax.get<{ data: { data: string; type: string } }>(
    `/api/gitlab/projects/${id}/branches/${branch}/config_file`
  );
}

export function enabledProject(id: number) {
  return ajax.post(`/api/gitlab/projects/enable`, { gitlab_project_id: id });
}
export function disabledProject(id: number) {
  return ajax.post(`/api/gitlab/projects/disable`, { gitlab_project_id: id });
}
