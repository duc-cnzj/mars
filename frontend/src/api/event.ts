import ajax from "./ajax";
import pb from "./compiled";

export async function events({ page, page_size }: pb.EventListRequest) {
  return ajax.get<pb.EventListResponse>(`/api/events`, {
    params: { page, page_size },
  });
}
