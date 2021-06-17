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
    console.log("在发送请求之前做些什么", config.url);
    return config;
  },
  (error) => {
    console.log("对请求错误做些什么");
    return Promise.reject(error);
  }
);

// 添加响应拦截器
ajax.interceptors.response.use(
  (response) => {
    // 对响应数据做点什么
    console.log("对响应数据做点什么");
    console.log(response);
    return response;
  },
  (error) => {
    console.log(error);
    // 对响应错误做点什么
    console.log("对响应错误做点什么");
    return Promise.reject(error);
  }
);

export default ajax;
