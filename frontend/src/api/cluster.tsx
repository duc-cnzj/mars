import ajax from "./ajax";

export async function clusterInfo() {
  return ajax.get<{data: API.ClusterInfo}>(`/api/cluster_info`);
}
