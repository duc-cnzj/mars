import ajax from "./ajax";
import pb from "./compiled";

export async function copyToPod({ pod, namespace, container, file_id }: pb.CopyToPodRequest) {
  return ajax.post<pb.CopyToPodResponse>(`/api/copy_to_pod`, { pod, namespace, container, file_id });
}