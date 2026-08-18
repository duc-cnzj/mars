package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/duc-cnzj/mars/api/v6/http/rest"
)

// 本文件只放不在任何 proto 里的自定义路由（生成器覆盖不了，永远手写）：
// multipart 上传、二进制下载、从 pod 拷文件。标准 CRUD 方法见 rest/file.gen.http.go
// （由 gen 生成，自动与 proto 对齐）。
//
// 与自动生成的分开放：手写代码留在本包（http），生成代码在 rest 子包。
// 通过嵌入 *rest.FileSvc，cli.File() 这一个访问器既能调生成的 CRUD，也能调这里
// 的手写路由，调用方式完全一致。

// FileAPI 是 File service 的完整客户端：嵌入生成的 *rest.FileSvc（CRUD 方法），
// 再补上手写路由（生成器覆盖不到的 multipart/二进制/copy_from_pod）。
type FileAPI struct {
	*rest.FileSvc
	c *Client
}

// UploadResponse 上传接口返回体 {"id": N}。
type UploadResponse struct {
	ID int `json:"id"`
}

// UploadFile 以 multipart 方式上传文件，字段名为 "file"，返回文件 ID。
func (f *FileAPI) UploadFile(ctx context.Context, filename string, r io.Reader) (*UploadResponse, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	// CreateFormFile 的字段名硬编码为 "file"，恒合法；mw 写 bytes.Buffer，Close 恒成功——
	// 两处错误分支 provably unreachable，直接丢弃（S 级零死代码）。
	fw, _ := mw.CreateFormFile("file", filepath.Base(filename))
	if _, err := io.Copy(fw, r); err != nil {
		return nil, err
	}
	_ = mw.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.c.baseURL+"/api/files", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if tok := f.c.authToken(); tok != "" {
		req.Header.Set("Authorization", tok)
	}
	f.c.applyHeaders(req)

	resp, err := f.c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	// 接受任意 2xx（网关可能配成 200 而非 201），与 DownloadFile/CopyFromPod/doReq 的 2xx 判定语义对齐。
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, f.c.errFromStatus(resp.StatusCode, data)
	}
	var out UploadResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DownloadInfo 下载响应元信息。
type DownloadInfo struct {
	Filename string
	Size     int64
}

// DownloadFile 下载已上传的文件，返回文件流与元信息。
// 返回的 io.ReadCloser 由调用方负责关闭。
func (f *FileAPI) DownloadFile(ctx context.Context, id int) (io.ReadCloser, *DownloadInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/download_file/%d", f.c.baseURL, id), nil)
	if err != nil {
		return nil, nil, err
	}
	if tok := f.c.authToken(); tok != "" {
		req.Header.Set("Authorization", tok)
	}
	f.c.applyHeaders(req)

	resp, err := f.c.hc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, nil, f.c.errFromStatus(resp.StatusCode, data)
	}

	info := &DownloadInfo{Size: resp.ContentLength}
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			info.Filename = params["filename"]
		}
	}
	return resp.Body, info, nil
}

// CopyFromPodRequest 对应 POST /api/copy_from_pod 的 JSON 体。
type CopyFromPodRequest struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
	FilePath  string `json:"filepath"`
}

// CopyFromPod 从 pod 拷贝文件到本地，返回文件流。
// 返回的 io.ReadCloser 由调用方负责关闭。
func (f *FileAPI) CopyFromPod(ctx context.Context, req *CopyFromPodRequest) (io.ReadCloser, *DownloadInfo, error) {
	// CopyFromPodRequest 是纯 string 平铺结构体，json.Marshal 恒不报错，错误分支丢弃。
	data, _ := json.Marshal(req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, f.c.baseURL+"/api/copy_from_pod", bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if tok := f.c.authToken(); tok != "" {
		httpReq.Header.Set("Authorization", tok)
	}
	f.c.applyHeaders(httpReq)

	resp, err := f.c.hc.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, nil, f.c.errFromStatus(resp.StatusCode, body)
	}

	info := &DownloadInfo{Size: resp.ContentLength}
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			info.Filename = params["filename"]
		}
	}
	return resp.Body, info, nil
}
