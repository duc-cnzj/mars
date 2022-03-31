import { message } from "antd";

export function copy(text: string, msg?: string) {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(text);
  } else {
    // 创建text area
    let textArea = document.createElement("textarea");
    textArea.value = text;
    // 使text area不在viewport，同时设置不可见
    textArea.style.position = "absolute";
    textArea.style.opacity = "0";
    textArea.style.left = "-999999px";
    textArea.style.top = "-999999px";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    document.execCommand("copy");
    textArea.remove();
  }
  message.success(msg ? msg : "已复制");
}
