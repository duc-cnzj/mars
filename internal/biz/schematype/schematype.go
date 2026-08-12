package schematype

// MarsAdmin 是内置管理员角色名。
const MarsAdmin = "mars_admin"

// UserInfo 是登录用户信息模型：OIDC 声明的领域映射，
// 供认证与会话链路使用，json 字段名与 OIDC userinfo 契约一致。
type UserInfo struct {
	ID      string   `json:"id"`
	Email   string   `json:"email"`
	Name    string   `json:"name"`
	Picture string   `json:"picture"`
	Roles   []string `json:"roles"`

	LogoutUrl string `json:"logout_url"`
}

// IsAdmin 判断用户是否拥有管理员角色。
func (ui *UserInfo) IsAdmin() bool {
	for _, role := range ui.Roles {
		if role == MarsAdmin {
			return true
		}
	}
	return false
}

// UploadType 是上传类型分类：标识文件存储在本地磁盘还是 S3 对象存储。
type UploadType string

// Local 标识文件存储在本地磁盘。
const Local UploadType = "local"

// S3 标识文件存储在 S3 对象存储。
const S3 UploadType = "s3"
