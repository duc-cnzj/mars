import ajax from "./ajax";
import pb from "./compiled";

export async function login({ username, password }: pb.AuthLoginRequest) {
  return ajax.post<pb.AuthLoginResponse>(`/api/auth/login`, {
    username,
    password,
  });
}
export async function info() {
  return ajax.get<pb.AuthInfoResponse>(`/api/auth/info`);
}
export async function settings() {
  return ajax.get<pb.AuthSettingsResponse>(`/api/auth/settings`);
}

export async function exchange({ code }: pb.AuthExchangeRequest) {
  return ajax.post<pb.AuthLoginResponse>(`/api/auth/exchange`, { code });
}
