import ajax from "./ajax";
import pb from "./compiled"

export async function changelogs({project_id, only_changed}: pb.ChangelogShowRequest) {
  return ajax.get<pb.ChangelogShowResponse>(`/api/projects/${project_id}/changelogs?only_changed=${only_changed}`);
}
