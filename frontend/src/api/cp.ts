import ajax from "./ajax";
import pb from "./compiled";

export async function copyToPod({ pod, namespace, container, file_id }: pb.ContainerCopyToPodRequest) {
  return ajax.post<pb.ContainerCopyToPodResponse>(`/api/containers/copy_to_pod`, { pod, namespace, container, file_id });
}