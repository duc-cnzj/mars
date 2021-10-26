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
export const getRandom = (): boolean => {
  return window.localStorage.getItem("random") === "1";
};
export const toggleRandom = (): boolean => {
  let r = window.localStorage.getItem("random");
  if (r === "1") {
    window.localStorage.setItem("random", "0");
    return false;
  } else {
    window.localStorage.setItem("random", "1");
    return true;
  }
};
