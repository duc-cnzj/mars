import ajax from "./ajax";
import pb from "./compiled";

export async function login({ username, password }: pb.LoginRequest) {
  return ajax.post<pb.LoginResponse>(`/api/auth/login`, {
    username,
    password,
  });
}
export async function info() {
  return ajax.get<pb.InfoResponse>(`/api/auth/info`);
}
export async function settings() {
  return ajax.get<pb.SettingsResponse>(`/api/auth/settings`);
}

export async function exchange({ code }: pb.ExchangeRequest) {
  return ajax.post<pb.LoginResponse>(`/api/auth/exchange`, { code });
}
