package schematype

// MarsAdmin 是内置管理员角色名。
const MarsAdmin = "mars_admin"

// SuperAdminEmail 是内置超级管理员身份的固定邮箱，单一事实来源：
// admin 登录返回的 adminUserInfo 与 data 层 namespace 创建者"超级管理员"
// 展示映射、以及各接口 is_super_admin 字段填充共用此值。
const SuperAdminEmail = "1025434218@qq.com"

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

// IsSuperAdmin 判断用户是否为内置超级管理员（固定邮箱身份，登录绕过 OIDC）。
func (ui *UserInfo) IsSuperAdmin() bool {
	return ui.Email == SuperAdminEmail
}

// UploadType 是上传类型分类：标识文件存储在本地磁盘还是 S3 对象存储。
type UploadType string

// Local 标识文件存储在本地磁盘。
const Local UploadType = "local"

// S3 标识文件存储在 S3 对象存储。
const S3 UploadType = "s3"
