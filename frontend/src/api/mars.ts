import ajax from "./ajax";
import pb from "./compiled";

export async function marsConfig({ git_project_id, branch }: pb.gitconfig.ShowRequest) {
  return ajax
    .get<pb.gitconfig.ShowResponse>(
      `/api/git/projects/${git_project_id}/mars_config?branch=${branch || ""}`
    )
}

export async function toggleGlobalEnabled({
  git_project_id,
  enabled,
}: pb.gitconfig.ToggleGlobalStatusRequest) {
  return ajax.post<pb.gitconfig.ToggleGlobalStatusResponse>(`/api/git/projects/${git_project_id}/toggle_status`, {
    enabled,
  });
}

export async function globalConfig({ git_project_id }: pb.gitconfig.GlobalConfigRequest) {
  return ajax.get<pb.gitconfig.GlobalConfigResponse>(
    `/api/git/projects/${git_project_id}/global_config`
  );
}

export async function updateGlobalConfig({
  git_project_id,
  config,
}: pb.gitconfig.UpdateRequest) {
  return ajax.put<pb.gitconfig.UpdateResponse>(
    `/api/git/projects/${git_project_id}/mars_config`,
    { config: config }
  );
}

export async function getDefaultValues({
  git_project_id,
  branch,
}: pb.gitconfig.DefaultChartValuesRequest) {
  return ajax.get<pb.gitconfig.DefaultChartValuesResponse>(
    `/api/git/projects/${git_project_id}/default_values?branch=${branch}`
  );
}
