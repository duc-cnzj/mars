import ajax from "./ajax";
import pb from "./compiled";

export async function clusterInfo() {
  return ajax.get<pb.cluster.InfoResponse>(`/api/cluster_info`);
}
