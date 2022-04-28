import ajax from "./ajax";
import pb from "./compiled";

export async function login({ username, password }: pb.auth.LoginRequest) {
  return ajax.post<pb.auth.LoginResponse>(`/api/auth/login`, {
    username,
    password,
  });
}
export async function info() {
  return ajax.get<pb.auth.InfoResponse>(`/api/auth/info`);
}
export async function settings() {
  return ajax.get<pb.auth.SettingsResponse>(`/api/auth/settings`);
}

export async function exchange({ code }: pb.auth.ExchangeRequest) {
  return ajax.post<pb.auth.LoginResponse>(`/api/auth/exchange`, { code });
}
