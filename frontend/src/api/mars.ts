import ajax from "./ajax";
import pb from "./compiled";

export async function marsConfig({ git_project_id, branch }: pb.GitConfigShowRequest) {
  return ajax
    .get<pb.GitConfigShowResponse>(
      `/api/git/projects/${git_project_id}/mars_config?branch=${branch || ""}`
    )
}

export async function toggleGlobalEnabled({
  git_project_id,
  enabled,
}: pb.GitConfigToggleGlobalStatusRequest) {
  return ajax.post<pb.GitConfigToggleGlobalStatusResponse>(`/api/git/projects/${git_project_id}/toggle_status`, {
    enabled,
  });
}

export async function globalConfig({ git_project_id }: pb.GitConfigGlobalConfigRequest) {
  return ajax.get<pb.GitConfigGlobalConfigResponse>(
    `/api/git/projects/${git_project_id}/global_config`
  );
}

export async function updateGlobalConfig({
  git_project_id,
  config,
}: pb.GitConfigUpdateRequest) {
  return ajax.put<pb.GitConfigUpdateResponse>(
    `/api/git/projects/${git_project_id}/mars_config`,
    { config: config }
  );
}

export async function getDefaultValues({
  git_project_id,
  branch,
}: pb.GitConfigDefaultChartValuesRequest) {
  return ajax.get<pb.GitConfigDefaultChartValuesResponse>(
    `/api/git/projects/${git_project_id}/default_values?branch=${branch}`
  );
}
