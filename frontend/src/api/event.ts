import ajax from "./ajax";
import pb from "./compiled";

export async function events({
  page,
  page_size,
  action_type,
  search,
}: pb.event.ListRequest) {
  return ajax.get<pb.event.ListResponse>(`/api/events`, {
    params: { page, page_size, action_type, search },
  });
}

export async function showEvent(id: number) {
  return ajax.get<pb.event.ShowResponse>(`/api/events/${id}`);
}
