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

export function getNamespaceCpuMemory({namespace_id}: pb.MetricsCpuMemoryInNamespaceRequest) {
  return ajax.get<pb.MetricsCpuMemoryInNamespaceResponse>(`/api/metrics/namespace/${namespace_id}/cpu_memory`);
}

export function getProjectServiceEndpoints({project_id}: pb.EndpointInProjectRequest) {
  return ajax.get<pb.EndpointInProjectResponse>(`/api/endpoints/projects/${project_id}`);
}

export function getNamespaceServiceEndpoints({namespace_id}: pb.EndpointInNamespaceRequest) {
  return ajax.get<pb.EndpointInNamespaceResponse>(`/api/endpoints/namespaces/${namespace_id}`);
}
