import { message } from "antd";
import ajax from "./ajax";
import pb from "./compiled";

export function deleteFile({ id }: pb.file.DeleteRequest) {
  return ajax.delete<pb.file.DeleteResponse>(`/api/files/${id}`);
}

export function showRecords(id: number) {
  return ajax.get<pb.file.ShowRecordsResponse>(`/api/record_files/${id}`);
}

export function diskInfo() {
  return ajax.get<pb.file.DiskInfoResponse>(`/api/files/disk_info`);
}

export function downloadFile(id: number) {
  return download(`/api/download_file/${id}`);
}

export function downloadConfig(pid?: number) {
  let url = `/api/config/export`;
  if (pid && pid > 0) {
    url += `/${pid}`;
  }
  return download(url);
}

export function maxUploadSize() {
  return ajax.get<pb.file.MaxUploadSizeResponse>(`/api/files/max_upload_size`);
}

const download = (url: string) => {
  return ajax
    .get(url, { responseType: "blob" })
    .then((res) => {
      const url = window.URL.createObjectURL(res.data);
      const contentDisposition = res.headers["content-disposition"];
      console.log(contentDisposition);
      let fileName = "unknown";
      if (contentDisposition) {
        const fileNameMatch = contentDisposition.match(/filename="(.+)"/);
        if (fileNameMatch?.length === 2) fileName = fileNameMatch[1];
      }
      console.log(fileName);
      let a = document.createElement("a");
      a.style.display = "none";
      a.href = url;
      a.setAttribute("download", fileName);
      document.body.appendChild(a);
      a.click(); //执行下载
      window.URL.revokeObjectURL(a.href); //释放url
      document.body.removeChild(a); //释放标签
    })
    .catch((e) => {
      if (e.response.status === 404) {
        message.error("文件不存在");
      }
    });
};
