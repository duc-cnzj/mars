import ajax from "./ajax";
import pb from "./compiled"

export async function List(obj: pb.token.ListRequest) {
  return ajax.get<pb.token.ListResponse>(`/api/access_tokens`, {params: obj});
}
export async function Grant(req: pb.token.GrantRequest) {
  return ajax.post<pb.token.GrantRequest>(`/api/access_tokens`, req);
}

export async function Lease(req: pb.token.LeaseRequest) {
  return ajax.put<pb.token.LeaseResponse>(`/api/access_tokens/${req.token}`, req);
}

export async function Revoke(req: pb.token.RevokeRequest) {
  return ajax.delete<pb.token.RevokeResponse>(`/api/access_tokens/${req.token}`);
}
