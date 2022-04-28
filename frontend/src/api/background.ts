import ajax from "./ajax";
import pb from "./compiled"

export async function bg({random}: pb.picture.BackgroundRequest) {
  return ajax.get<pb.picture.BackgroundResponse>(`/api/picture/background?random=${random}`);
}
