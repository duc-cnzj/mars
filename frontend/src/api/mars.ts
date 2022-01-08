import ajax from "./ajax";
import pb from "./compiled";

export async function marsConfig({ project_id, branch }: pb.MarsShowRequest) {
  return ajax
    .get<pb.MarsShowResponse>(
      `/api/gitlab/projects/${project_id}/mars_config?branch=${branch || ""}`
    )
}

export async function toggleGlobalEnabled({
  project_id,
  enabled,
}: pb.MarsToggleEnabledRequest) {
  return ajax.post<pb.MarsToggleEnabledResponse>(`/api/gitlab/projects/${project_id}/toggle_enabled`, {
    enabled,
  });
}

export async function globalConfig({ project_id }: pb.MarsGlobalConfigRequest) {
  return ajax.get<pb.MarsGlobalConfigResponse>(
    `/api/gitlab/projects/${project_id}/global_config`
  );
}

export async function updateGlobalConfig({
  project_id,
  config,
}: pb.MarsUpdateRequest) {
  return ajax.put<pb.MarsUpdateResponse>(
    `/api/gitlab/projects/${project_id}/mars_config`,
    { config: config }
  );
}

export async function getDefaultValues({
  project_id,
  branch,
}: pb.MarsDefaultChartValuesRequest) {
  return ajax.get<pb.MarsDefaultChartValuesResponse>(
    `/api/gitlab/projects/${project_id}/default_values?branch=${branch}`
  );
}
