import ajax from "./ajax";

export function createNamespace(namespace: string) {
  return ajax.post<{ data: {id: number; name: string} }>("/api/namespaces", { namespace });
}

export interface NamespaceItem {
    id: number;
    name: string;
    created_at: string;
    updated_at: string;
    projects?: {
      id: number;
      name: string;
      status: string;
    }[];
}

export function listNamespaces() {
  return ajax.get<{data: NamespaceItem[]}>("/api/namespaces");
}

export function deleteNamespace(id: number) {
  return ajax.delete(`/api/namespaces/${id}`);
}

export function getNamespaceCpuAndMemory(id: number) {
  return ajax.get<{data: {cpu: string, memory: string}}>(`/api/namespaces/${id}/cpu_and_memory`);
}

export function getServiceEndpoints(id: number, projectName?: string) {
  let url = `/api/namespaces/${id}/service_endpoints`
  if (projectName) {
    url += "?project_name="+projectName
  }
  return ajax.get<{data: {[name:string]:string[];}}>(url);
}
