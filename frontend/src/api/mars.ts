import ajax from "./ajax";
import pb from "./compiled";

export async function marsConfig({ git_project_id, branch }: pb.GitProjectConfigShowRequest) {
  return ajax
    .get<pb.GitProjectConfigShowResponse>(
      `/api/gitproject/projects/${git_project_id}/mars_config?branch=${branch || ""}`
    )
}

export async function toggleGlobalEnabled({
  git_project_id,
  enabled,
}: pb.GitProjectConfigToggleGlobalStatusRequest) {
  return ajax.post<pb.GitProjectConfigToggleGlobalStatusResponse>(`/api/gitproject/projects/${git_project_id}/toggle_status`, {
    enabled,
  });
}

export async function globalConfig({ git_project_id }: pb.GitProjectConfigGlobalConfigRequest) {
  return ajax.get<pb.GitProjectConfigGlobalConfigResponse>(
    `/api/gitproject/projects/${git_project_id}/global_config`
  );
}

export async function updateGlobalConfig({
  git_project_id,
  config,
}: pb.GitProjectConfigUpdateRequest) {
  return ajax.put<pb.GitProjectConfigUpdateResponse>(
    `/api/gitproject/projects/${git_project_id}/mars_config`,
    { config: config }
  );
}

export async function getDefaultValues({
  git_project_id,
  branch,
}: pb.GitProjectConfigDefaultChartValuesRequest) {
  return ajax.get<pb.GitProjectConfigDefaultChartValuesResponse>(
    `/api/gitproject/projects/${git_project_id}/default_values?branch=${branch}`
  );
}
