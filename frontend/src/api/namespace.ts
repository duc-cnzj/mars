import ajax from "./ajax";
import pb from "./compiled"

export function createNamespace(namespace: string) {
  return ajax.post<pb.NsStoreResponse>("/api/namespaces", { namespace });
}

export function listNamespaces() {
  return ajax.get<pb.NamespaceList>("/api/namespaces");
}

export function deleteNamespace({namespace_id}: pb.NamespaceID) {
  return ajax.delete(`/api/namespaces/${namespace_id}`);
}

export function getNamespaceCpuAndMemory({namespace_id}: pb.NamespaceID) {
  return ajax.get<pb.CpuAndMemoryResponse>(`/api/namespaces/${namespace_id}/cpu_and_memory`);
}

export function getServiceEndpoints({project_name, namespace_id}: pb.ServiceEndpointsRequest) {
  let url = `/api/namespaces/${namespace_id}/service_endpoints`
  if (project_name) {
    url += "?project_name="+project_name
  }
  return ajax.get<pb.ServiceEndpointsResponse>(url);
}
