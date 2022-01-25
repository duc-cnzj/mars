import ajax from "./ajax";
import pb from "./compiled";

export function deleteFile({ id }: pb.FileDeleteRequest) {
  return ajax.delete<pb.FileDeleteResponse>(`/api/files/${id}`);
}

export function diskInfo() {
  return ajax.get<pb.DiskInfoResponse>(`/api/files/disk_info`);
}

export function deleteUndocumentedFiles() {
  return ajax.delete<pb.DeleteUndocumentedFilesResponse>(`/api/files/delete_undocumented_files`);
}

export function downloadFile(id: number) {
  return ajax
    .get(`/api/download_file/${id}`, { responseType: "blob" })
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
    });
}
