import ajax from "./ajax";
import pb from "./compiled"

export async function bg({random}: pb.BackgroundRequest) {
  return ajax.get<pb.BackgroundResponse>(`/api/picture/background?random=${random}`);
}
