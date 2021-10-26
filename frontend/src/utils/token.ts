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
export const setState = (state: string) => {
  window.localStorage.setItem("state", state);
};
export const removeState = () => {
  window.localStorage.removeItem("state");
};
export const getState = (): string => {
  return window.localStorage.getItem("state") || "";
};
export const isRandomBg = (): boolean => {
  return window.localStorage.getItem("random_bg") === "1";
};
export const toggleRandomBg = (): boolean => {
  let r = window.localStorage.getItem("random_bg");
  if (r === "1") {
    window.localStorage.setItem("random_bg", "0");
    return false;
  } else {
    window.localStorage.setItem("random_bg", "1");
    return true;
  }
};
