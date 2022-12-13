import ajax from "./ajax";
import pb from "./compiled";

export function createNamespace(namespace: string) {
  return ajax.post<pb.namespace.CreateResponse>("/api/namespaces", {
    namespace,
  });
}

export function allNamespaces() {
  return ajax.get<pb.namespace.AllResponse>("/api/namespaces");
}

export function deleteNamespace({ namespace_id }: pb.namespace.DeleteRequest) {
  return ajax.delete<pb.namespace.DeleteResponse>(
    `/api/namespaces/${namespace_id}`
  );
}

export function getNamespaceCpuMemory({
  namespace_id,
}: pb.metrics.CpuMemoryInNamespaceRequest) {
  return ajax.get<pb.metrics.CpuMemoryInNamespaceResponse>(
    `/api/metrics/namespace/${namespace_id}/cpu_memory`
  );
}

export function getProjectServiceEndpoints({
  project_id,
}: pb.endpoint.InProjectRequest) {
  return ajax.get<pb.endpoint.InProjectResponse>(
    `/api/endpoints/projects/${project_id}`
  );
}

export function getNamespaceServiceEndpoints({
  namespace_id,
}: pb.endpoint.InNamespaceRequest) {
  return ajax.get<pb.endpoint.InNamespaceResponse>(
    `/api/endpoints/namespaces/${namespace_id}`
  );
}
