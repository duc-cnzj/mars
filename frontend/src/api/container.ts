import ajax from "./ajax";
import pb from "./compiled";

export function containerLog({
  pod,
  namespace,
  container,
}: pb.container.LogRequest) {
  return ajax.get<pb.container.LogResponse>(
    `/api/containers/namespaces/${namespace}/pods/${pod}/containers/${container}/logs`
  );
}
