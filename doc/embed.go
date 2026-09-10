package doc

import _ "embed"

// SwaggerJson 是编译期嵌入的 OpenAPI 规范原文（openapi.yaml），
// 由 swaggerHandler 在 /doc/swagger.json 原样输出给 swagger UI。
//
//go:embed openapi.yaml
var SwaggerJson []byte
