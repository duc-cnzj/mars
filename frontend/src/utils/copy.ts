import { message } from "antd";

export function copy(text: string, msg?: string) {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(text);
    message.success(msg ? msg : "已复制");
  } else {
    message.error("你的浏览器不支持该操作");
  }
}
