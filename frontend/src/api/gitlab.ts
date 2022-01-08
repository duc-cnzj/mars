import ajax from "./ajax";
import pb from "./compiled"

export function allProjects() {
  return ajax.get<pb.GitAllProjectsResponse>("/api/gitserver/projects");
}

export function projectOptions() {
  return ajax.get<pb.GitProjectOptionsResponse>("/api/gitserver/project_options");
}

export function branchOptions({project_id, all}: pb.GitBranchOptionsRequest) {
  return ajax.get<pb.GitBranchOptionsResponse>(`/api/gitserver/projects/${project_id}/branch_options?all=${all}`);
}

export function commitOptions({project_id, branch}: pb.GitCommitOptionsRequest) {
  return ajax.get<pb.GitCommitOptionsResponse>(
    `/api/gitserver/projects/${project_id}/branches/${branch}/commit_options`
  );
}

export function commit({project_id, branch, commit}: pb.GitCommitRequest) {
  return ajax.get<pb.GitCommitResponse>(
    `/api/gitserver/projects/${project_id}/branches/${branch}/commits/${commit}`
  );
}

export function pipelineInfo({project_id, branch, commit}: pb.GitPipelineInfoRequest) {
  return ajax.get<pb.GitPipelineInfoResponse>(
    `/api/gitserver/projects/${project_id}/branches/${branch}/commits/${commit}/pipeline_info`
  );
}

export function configFile({project_id, branch}: pb.GitConfigFileRequest) {
  return ajax.get<pb.GitConfigFileResponse>(
    `/api/gitserver/projects/${project_id}/branches/${branch}/config_file`
  );
}

export function enabledProject({git_project_id}: pb.GitEnableProjectRequest) {
  return ajax.post(`/api/gitserver/projects/enable`, { git_project_id });
}

export function disabledProject({git_project_id}: pb.GitDisableProjectRequest) {
  return ajax.post(`/api/gitserver/projects/disable`, { git_project_id });
}
