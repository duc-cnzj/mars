import ajax from "./ajax";
import pb from "./compiled"

export function projectList() {
  return ajax.get<pb.ProjectListResponse>("/api/gitlab/project_list");
}

export function projects() {
  return ajax.get<pb.ProjectsResponse>("/api/gitlab/projects");
}

export function branches({project_id}: pb.BranchesRequest) {
  return ajax.get<pb.BranchesResponse>(`/api/gitlab/projects/${project_id}/branches`);
}

export function commits({project_id, branch}: pb.CommitsRequest) {
  return ajax.get<pb.CommitsResponse>(
    `/api/gitlab/projects/${project_id}/branches/${branch}/commits`
  );
}

export function commit({project_id, branch, commit}: pb.CommitRequest) {
  return ajax.get<pb.CommitResponse>(
    `/api/gitlab/projects/${project_id}/branches/${branch}/commits/${commit}`
  );
}

export function pipelineInfo({project_id, branch, commit}: pb.PipelineInfoRequest) {
  return ajax.get<pb.PipelineInfoResponse>(
    `/api/gitlab/projects/${project_id}/branches/${branch}/commits/${commit}/pipeline_info`
  );
}

export function configFile({project_id, branch}: pb.ConfigFileRequest) {
  return ajax.get<pb.ConfigFileResponse>(
    `/api/gitlab/projects/${project_id}/branches/${branch}/config_file`
  );
}

export function enabledProject({gitlab_project_id}: pb.EnableProjectRequest) {
  return ajax.post(`/api/gitlab/projects/enable`, { gitlab_project_id });
}

export function disabledProject({gitlab_project_id}: pb.DisableProjectRequest) {
  return ajax.post(`/api/gitlab/projects/disable`, { gitlab_project_id });
}
