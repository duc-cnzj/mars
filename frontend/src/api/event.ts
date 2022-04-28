import ajax from "./ajax";
import pb from "./compiled";

export async function events({ page, page_size }: pb.event.ListRequest) {
  return ajax.get<pb.event.ListResponse>(`/api/events`, {
    params: { page, page_size },
  });
}
