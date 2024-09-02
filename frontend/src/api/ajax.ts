import { message } from "antd";
import {
  getToken,
  removeToken,
  getLogoutUrl,
  removeLogoutUrl,
} from "./../utils/token";
import createClient, { Middleware } from "openapi-fetch";
import { paths } from "./schema";
import _ from "lodash";

const login401alert = _.debounce(() => {
  message.error("登录过期，请重新登录");
}, 500);

const ajax = createClient<paths>({
  baseUrl: process.env.REACT_APP_BASE_URL,
  headers: {
    "X-Requested-With": "XMLHttpRequest",
    "Accept-Language": "zh",
  },
});

const myMiddleware: Middleware = {
  async onRequest({ request, options }) {
    request.headers.set("Authorization", getToken());
    return request;
  },
  async onResponse({ request, response, options }) {
    // 对响应错误做点什么
    if (response.status === 401) {
      if (getToken()) {
        removeToken();
        login401alert();
      }
      setTimeout(() => {
        if (window.location.pathname !== "/login") {
          let href = getLogoutUrl() || "/login";
          removeLogoutUrl();
          window.location.href = href;
        }
      }, 1000);
    }
    return response;
  },
};

ajax.use(myMiddleware);

export default ajax;
