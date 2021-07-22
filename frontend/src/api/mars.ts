import ajax from "./ajax";

export async function marsConfig(
  projectID: number,
  { branch }: { branch?: string }
) {
  return ajax
    .get<{ data: { branch: string; config: string } }>(
      `/api/gitlab/projects/${projectID}/mars_config?branch=${branch || ""}`
    )
    .then(
      (
        res
      ): {
        branch: string;
        config: string;
      } => {
        return {
          branch: res.data.data.branch,
          config: res.data.data.config,
        };
      }
    );
}

export async function toggleGlobalEnabled(projectID: number, enabled: boolean) {
  return ajax.post<{ data: { enabled: boolean; config: string } }>(
    `/api/gitlab/projects/${projectID}/toggle_enabled`,
    { enabled }
  );
}

export async function globalConfig(projectID: number) {
  return ajax
    .get<{ data: { enabled: boolean; config: string } }>(
      `/api/gitlab/projects/${projectID}/global_config`
    )
    .then(
      (
        res
      ): {
        enabled: boolean;
        config: string;
      } => {
        return {
          enabled: res.data.data.enabled,
          config: res.data.data.config,
        };
      }
    );
}

interface Config {
  global_config: string;
}

export async function updateGlobalConfig(projectID: number, config: string) { 
  return ajax
    .put<{ data: Config }>(
      `/api/gitlab/projects/${projectID}/mars_config`,
      {config: config}
    )
}