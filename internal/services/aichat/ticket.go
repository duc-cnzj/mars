package aichat

// ticket.go 承载写工具人工确认的无状态票据：HMAC 签名把「用户 + 工具调用 + 参数」
// 绑定在一起，服务端不落任何会话状态，多副本部署下任意副本都能校验同一张票据。
//
// 校验顺序固定为：格式 → 签名 → 有效期 → （调用方再核对的）用户身份与调用匹配。
// 签名密钥由 ai.api_key 派生：每个副本配置相同，因此无需共享存储；ai.api_key
// 为空时对话入口本就返回 503，不会走到签发路径。

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	// ticketTTL 是确认票据的有效期：从模型发起调用到用户点「批准」通常只有几十秒，
	// 5 分钟足以覆盖慢操作，又能限制票据被长期重放。
	ticketTTL = 5 * time.Minute
	// ticketKeyPrefix 给 HMAC 密钥加域前缀，避免与 api_key 在其他用途的派生撞车。
	ticketKeyPrefix = "mars-ai-chat-confirm:"
)

var (
	// errTicketMalformed 表示票据不是「payload.signature」两段 base64url 格式。
	errTicketMalformed = errors.New("确认票据格式不正确")
	// errTicketSignature 表示票据被篡改或由其他密钥签发。
	errTicketSignature = errors.New("确认票据签名无效")
	// errTicketExpired 表示票据已超过有效期。
	errTicketExpired = errors.New("确认票据已过期，请重新发起对话")
)

// ticketPayload 是票据签名的内容：只放校验所需的字段，不放对话内容与工具结果。
type ticketPayload struct {
	// Email 是签发时的登录用户，用于拒绝跨用户复用票据。
	Email string `json:"email"`
	// ToolCallID 是模型生成的那一次工具调用的 id。
	ToolCallID string `json:"tool_call_id"`
	// ToolName 是被确认的工具名。
	ToolName string `json:"tool_name"`
	// ArgsSHA256 是工具参数原文的 SHA-256，防止确认后替换参数。
	ArgsSHA256 string `json:"args_sha256"`
	// Exp 是过期时间（Unix 秒）。
	Exp int64 `json:"exp"`
}

// ticketSigner 负责签发与校验确认票据。now 可替换以便测试过期分支。
type ticketSigner struct {
	key []byte
	now func() time.Time
}

// newTicketSigner 由上游 API Key 派生 HMAC 密钥。派生而非直接使用，避免把上游
// 凭据本身当作签名密钥使用。
func newTicketSigner(apiKey string) *ticketSigner {
	sum := sha256.Sum256([]byte(ticketKeyPrefix + apiKey))
	return &ticketSigner{key: sum[:], now: time.Now}
}

// issue 为一次写工具调用签发票据。payload 是固定结构体（字符串与整数），
// 序列化不会失败，因此不返回错误。
func (s *ticketSigner) issue(email string, call toolCall) string {
	payload := ticketPayload{
		Email:      email,
		ToolCallID: call.ID,
		ToolName:   call.Function.Name,
		ArgsSHA256: argsHash(call.Function.Arguments),
		Exp:        s.now().Add(ticketTTL).Unix(),
	}
	raw, _ := json.Marshal(payload)
	signature := s.sign(raw)
	return base64.RawURLEncoding.EncodeToString(raw) + "." +
		base64.RawURLEncoding.EncodeToString(signature)
}

// verify 校验票据格式、签名与有效期，返回票据内容。
func (s *ticketSigner) verify(token string) (ticketPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ticketPayload{}, errTicketMalformed
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ticketPayload{}, errTicketMalformed
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ticketPayload{}, errTicketMalformed
	}
	if !hmac.Equal(signature, s.sign(raw)) {
		return ticketPayload{}, errTicketSignature
	}

	var payload ticketPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ticketPayload{}, errTicketMalformed
	}
	if payload.Exp <= s.now().Unix() {
		return ticketPayload{}, errTicketExpired
	}
	return payload, nil
}

// sign 计算 payload 的 HMAC-SHA256。
func (s *ticketSigner) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, s.key)
	mac.Write(payload)
	return mac.Sum(nil)
}

// argsHash 返回工具参数原文的十六进制 SHA-256。
func argsHash(args string) string {
	sum := sha256.Sum256([]byte(args))
	return hex.EncodeToString(sum[:])
}
