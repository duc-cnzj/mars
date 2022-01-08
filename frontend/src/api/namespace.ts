import ajax from "./ajax";
import pb from "./compiled"

export function createNamespace(namespace: string) {
  return ajax.post<pb.NamespaceCreateResponse>("/api/namespaces", { namespace });
}

export function allNamespaces() {
  return ajax.get<pb.NamespaceAllResponse>("/api/namespaces");
}

export function deleteNamespace({namespace_id}: pb.NamespaceDeleteRequest) {
  return ajax.delete<pb.NamespaceDeleteResponse>(`/api/namespaces/${namespace_id}`);
}

export function getNamespaceCpuMemory({namespace_id}: pb.NamespaceCpuMemoryRequest) {
  return ajax.get<pb.NamespaceCpuMemoryResponse>(`/api/namespaces/${namespace_id}/cpu_memory`);
}

export function getServiceEndpoints({project_name, namespace_id}: pb.NamespaceServiceEndpointsRequest) {
  let url = `/api/namespaces/${namespace_id}/service_endpoints`
  if (project_name) {
    url += "?project_name="+project_name
  }
  return ajax.get<pb.NamespaceServiceEndpointsResponse>(url);
}
