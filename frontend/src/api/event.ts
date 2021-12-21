import ajax from "./ajax";
import pb from "./compiled";

export async function events({ page, page_size }: pb.EventRequest) {
  return ajax.get<pb.EventList>(`/api/events`, {
    params: { page, page_size },
  });
}
