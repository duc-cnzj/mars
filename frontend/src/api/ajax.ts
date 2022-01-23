import { message } from "antd";
import {
  getToken,
  removeToken,
  getLogoutUrl,
  removeLogoutUrl,
} from "./../utils/token";
import axios from "axios";

const ajax = axios.create({
  baseURL: process.env.REACT_APP_BASE_URL,
  //   timeout: 1000,
  headers: {
    "X-Requested-With": "XMLHttpRequest",
    "Accept-Language": "zh",
  },
});

// 添加请求拦截器
ajax.interceptors.request.use(
  (config) => {
    if (config.headers) {
      config.headers["Authorization"] = getToken();
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 添加响应拦截器
ajax.interceptors.response.use(
  (response) => {
    // 对响应数据做点什么
    return response;
  },
  (error) => {
    // 对响应错误做点什么
    if (error.response.status === 401) {
      if (getToken()) {
        removeToken();
        message.error("登录过期，请重新登录");
      }
      setTimeout(() => {
        if (window.location.pathname !== "/login") {
          let href = getLogoutUrl() || "/login";
          removeLogoutUrl();
          window.location.href = href;
        }
      }, 1000);
    }
    return Promise.reject(error);
  }
);

export default ajax;
