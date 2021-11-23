import ajax from "./ajax";
import pb from "./compiled"

export function version() {
  return ajax.get<pb.VersionResponse>("/api/version");
}