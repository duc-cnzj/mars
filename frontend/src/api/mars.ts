import ajax from "./ajax";
import yaml from "js-yaml";

export async function marsConfig(
  projectID: number,
  { branch }: { branch?: string }
) {
  return ajax
    .get<{ data: {branch: string, config: string} }>(
      `/api/gitlab/projects/${projectID}/mars_config?branch=${branch || ""}`
    )
    .then(
      (
        res
      ): {
        branch: string;
        config: API.Mars;
      } => {
        const doc = yaml.load(res.data.data.config);
        return {
          branch: res.data.data.branch,
          config: doc as API.Mars,
        };
      }
    );
}
