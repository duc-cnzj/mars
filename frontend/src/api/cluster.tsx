import ajax from "./ajax";
import pb from "./compiled"

export async function clusterInfo() {
  return ajax.get<pb.ClusterInfoResponse>(`/api/cluster_info`);
}
