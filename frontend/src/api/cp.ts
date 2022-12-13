import ajax from "./ajax";
import pb from "./compiled";

export async function copyToPod({
  pod,
  namespace,
  container,
  file_id,
}: pb.container.CopyToPodRequest) {
  return ajax.post<pb.container.CopyToPodResponse>(
    `/api/containers/copy_to_pod`,
    { pod, namespace, container, file_id }
  );
}
