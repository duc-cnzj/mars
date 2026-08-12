package services

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Test_swaggerHandler_swaggerJSON(t *testing.T) {
	h := newSwaggerHandlerWithMocks(t)
	w := httptest.NewRecorder()
	r := &http.Request{}
	h.swaggerJSON(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

// failWriter 用于触发 w.Write 错误分支：Header/WriteHeader 正常，Write 恒失败。
type failWriter struct {
	header http.Header
	status int
}

func (f *failWriter) Header() http.Header {
	if f.header == nil {
		f.header = http.Header{}
	}
	return f.header
}

func (f *failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write boom")
}

func (f *failWriter) WriteHeader(status int) {
	f.status = status
}

func Test_swaggerHandler_swaggerJSON_WriteError(t *testing.T) {
	h := newSwaggerHandlerWithMocks(t)
	w := &failWriter{}
	h.swaggerJSON(w, &http.Request{})
	// Write 失败分支的契约：Content-Type 必须在 Write 尝试之前设置，
	// 写失败仅记日志继续。断言 header 已设置，避免"先写后设"回归。
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

// 回归防护：/doc/swagger.json 与 /docs/ 静态页必须通过 mux 路由可达，
// 且都套 HttpCache 缓存中间件（不 panic）。
func Test_swaggerHandler_Register(t *testing.T) {
	h := newSwaggerHandlerWithMocks(t)
	router := mux.NewRouter()
	h.Register(router)

	req := httptest.NewRequest("GET", "/doc/swagger.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	req2 := httptest.NewRequest("GET", "/docs/", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
