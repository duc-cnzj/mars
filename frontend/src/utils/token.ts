export const setToken = (token: string) => {
  window.localStorage.setItem("token", "Bearer " + token);
};
export const removeToken = () => {
  window.localStorage.removeItem("token");
};
export const getToken = (): string => {
  return window.localStorage.getItem("token") || "";
};
export const setLogoutUrl = (url: string) => {
  window.localStorage.setItem("logout_url", url);
};
export const removeLogoutUrl = () => {
  window.localStorage.removeItem("logout_url");
};
export const getLogoutUrl = (): string => {
  return window.localStorage.getItem("logout_url") || "";
};
