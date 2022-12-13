import ajax from "./ajax";
import pb from "./compiled";

export async function changelogs({
  project_id,
  only_changed,
}: pb.changelog.ShowRequest) {
  return ajax.get<pb.changelog.ShowResponse>(
    `/api/projects/${project_id}/changelogs?only_changed=${only_changed}`
  );
}
