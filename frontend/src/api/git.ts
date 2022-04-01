import ajax from "./ajax";
import pb from "./compiled"

export function allProjects() {
  return ajax.get<pb.GitAllProjectsResponse>("/api/git/projects");
}

export function projectOptions() {
  return ajax.get<pb.GitProjectOptionsResponse>("/api/git/project_options");
}

export function branchOptions({git_project_id, all}: pb.GitBranchOptionsRequest) {
  return ajax.get<pb.GitBranchOptionsResponse>(`/api/git/projects/${git_project_id}/branch_options?all=${all}`);
}

export function commitOptions({git_project_id, branch}: pb.GitCommitOptionsRequest) {
  return ajax.get<pb.GitCommitOptionsResponse>(
    `/api/git/projects/${git_project_id}/branches/${branch}/commit_options`
  );
}

export function commit({git_project_id, branch, commit}: pb.GitCommitRequest) {
  return ajax.get<pb.GitCommitResponse>(
    `/api/git/projects/${git_project_id}/branches/${branch}/commits/${commit}`
  );
}

export function pipelineInfo({git_project_id, branch, commit}: pb.GitPipelineInfoRequest) {
  return ajax.get<pb.GitPipelineInfoResponse>(
    `/api/git/projects/${git_project_id}/branches/${branch}/commits/${commit}/pipeline_info`
  );
}

export function configFile({git_project_id, branch}: pb.GitConfigFileRequest) {
  return ajax.get<pb.GitConfigFileResponse>(
    `/api/git/projects/${git_project_id}/branches/${branch}/config_file`
  );
}

export function enabledProject({git_project_id}: pb.GitEnableProjectRequest) {
  return ajax.post(`/api/git/projects/enable`, { git_project_id });
}

export function disabledProject({git_project_id}: pb.GitDisableProjectRequest) {
  return ajax.post(`/api/git/projects/disable`, { git_project_id });
}
