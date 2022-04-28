import ajax from "./ajax";
import pb from "./compiled"

export function allProjects() {
  return ajax.get<pb.git.AllProjectsResponse>("/api/git/projects");
}

export function projectOptions() {
  return ajax.get<pb.git.ProjectOptionsResponse>("/api/git/project_options");
}

export function branchOptions({git_project_id, all}: pb.git.BranchOptionsRequest) {
  return ajax.get<pb.git.BranchOptionsResponse>(`/api/git/projects/${git_project_id}/branch_options?all=${all}`);
}

export function commitOptions({git_project_id, branch}: pb.git.CommitOptionsRequest) {
  return ajax.get<pb.git.CommitOptionsResponse>(
    `/api/git/projects/${git_project_id}/branches/${branch}/commit_options`
  );
}

export function commit({git_project_id, branch, commit}: pb.git.CommitRequest) {
  return ajax.get<pb.git.CommitResponse>(
    `/api/git/projects/${git_project_id}/branches/${branch}/commits/${commit}`
  );
}

export function pipelineInfo({git_project_id, branch, commit}: pb.git.PipelineInfoRequest) {
  return ajax.get<pb.git.PipelineInfoResponse>(
    `/api/git/projects/${git_project_id}/branches/${branch}/commits/${commit}/pipeline_info`
  );
}

export function configFile({git_project_id, branch}: pb.git.ConfigFileRequest) {
  return ajax.get<pb.git.ConfigFileResponse>(
    `/api/git/projects/${git_project_id}/branches/${branch}/config_file`
  );
}

export function enabledProject({git_project_id}: pb.git.EnableProjectRequest) {
  return ajax.post(`/api/git/projects/enable`, { git_project_id });
}

export function disabledProject({git_project_id}: pb.git.DisableProjectRequest) {
  return ajax.post(`/api/git/projects/disable`, { git_project_id });
}
