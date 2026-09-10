package swagger_ui

import "embed"

// SwaggerUI 是编译期嵌入的 swagger-ui 静态站点（本目录全部文件），
// 由 swaggerHandler 挂载到 /docs/ 供浏览器直接打开。
//
//go:embed *
var SwaggerUI embed.FS
