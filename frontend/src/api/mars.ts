import ajax from "./ajax";
import pb from "./compiled";

export async function marsConfig({ project_id, branch }: pb.MarsShowRequest) {
  return ajax
    .get<pb.MarsShowResponse>(
      `/api/gitlab/projects/${project_id}/mars_config?branch=${branch || ""}`
    )
    .then(
      (
        res
      ): {
        branch: string;
        config: string;
      } => {
        return {
          branch: res.data.branch,
          config: res.data.config,
        };
      }
    );
}

export async function toggleGlobalEnabled({project_id, enabled}: pb.ToggleEnabledRequest) {
  return ajax.post(`/api/gitlab/projects/${project_id}/toggle_enabled`, {
    enabled,
  });
}

export async function globalConfig({project_id}: pb.GlobalConfigRequest) {
  return ajax
    .get<pb.GlobalConfigResponse>(
      `/api/gitlab/projects/${project_id}/global_config`
    )
    .then(
      (
        res
      ): {
        enabled: boolean;
        config: string;
      } => {
        return {
          enabled: res.data.enabled,
          config: res.data.config,
        };
      }
    );
}

export async function updateGlobalConfig({project_id, config}: pb.MarsUpdateRequest) {
  return ajax.put<pb.MarsUpdateResponse>(
    `/api/gitlab/projects/${project_id}/mars_config`,
    { config: config }
  );
}

export async function getDefaultValues({project_id, branch}: pb.DefaultChartValuesRequest) {
  return ajax.get<pb.DefaultChartValues>(
    `/api/gitlab/projects/${project_id}/default_values?branch=${branch}`,
  );
}
