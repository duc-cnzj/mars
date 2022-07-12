import ajax from "./ajax";
import pb from "./compiled";

export async function events({ page, page_size, action_type, message }: pb.event.ListRequest) {
  return ajax.get<pb.event.ListResponse>(`/api/events`, {
    params: { page, page_size, action_type, message },
  });
}
