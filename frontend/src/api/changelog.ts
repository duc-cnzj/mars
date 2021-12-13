import ajax from "./ajax";
import pb from "./compiled"

export async function changelogs({project_id, only_changed}: pb.ChangelogGetRequest) {
  return ajax.get<pb.ChangelogGetResponse>(`/api/projects/${project_id}/changelogs?only_changed=${only_changed}`);
}
