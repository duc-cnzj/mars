import ajax from "./ajax";
import pb from "./compiled"

export function allProjects() {
  return ajax.get<pb.GitAllProjectsResponse>("/api/gitproject/projects");
}

export function projectOptions() {
  return ajax.get<pb.GitProjectOptionsResponse>("/api/gitproject/project_options");
}

export function branchOptions({git_project_id, all}: pb.GitBranchOptionsRequest) {
  return ajax.get<pb.GitBranchOptionsResponse>(`/api/gitproject/projects/${git_project_id}/branch_options?all=${all}`);
}

export function commitOptions({git_project_id, branch}: pb.GitCommitOptionsRequest) {
  return ajax.get<pb.GitCommitOptionsResponse>(
    `/api/gitproject/projects/${git_project_id}/branches/${branch}/commit_options`
  );
}

export function commit({git_project_id, branch, commit}: pb.GitCommitRequest) {
  return ajax.get<pb.GitCommitResponse>(
    `/api/gitproject/projects/${git_project_id}/branches/${branch}/commits/${commit}`
  );
}

export function pipelineInfo({git_project_id, branch, commit}: pb.GitPipelineInfoRequest) {
  return ajax.get<pb.GitPipelineInfoResponse>(
    `/api/gitproject/projects/${git_project_id}/branches/${branch}/commits/${commit}/pipeline_info`
  );
}

export function configFile({git_project_id, branch}: pb.GitConfigFileRequest) {
  return ajax.get<pb.GitConfigFileResponse>(
    `/api/gitproject/projects/${git_project_id}/branches/${branch}/config_file`
  );
}

export function enabledProject({git_project_id}: pb.GitEnableProjectRequest) {
  return ajax.post(`/api/gitproject/projects/enable`, { git_project_id });
}

export function disabledProject({git_project_id}: pb.GitDisableProjectRequest) {
  return ajax.post(`/api/gitproject/projects/disable`, { git_project_id });
}
