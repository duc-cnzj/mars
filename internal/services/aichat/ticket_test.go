package aichat

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCall 构造一次用于签发票据的工具调用。
func testCall(id, name, args string) toolCall {
	return toolCall{ID: id, Type: "function", Function: toolFunction{Name: name, Arguments: args}}
}

// TestTicketRoundTrip 验证签发票据后能取回全部绑定字段。
func TestTicketRoundTrip(t *testing.T) {
	signer := newTicketSigner("key")
	call := testCall("call_1", "restart_pod", `{"namespace":"ns","pod":"p"}`)

	token := signer.issue("u@example.com", call)

	payload, err := signer.verify(token)
	require.NoError(t, err)
	assert.Equal(t, "u@example.com", payload.Email)
	assert.Equal(t, "call_1", payload.ToolCallID)
	assert.Equal(t, "restart_pod", payload.ToolName)
	assert.Equal(t, argsHash(call.Function.Arguments), payload.ArgsSHA256)
}

// TestTicketExpired 验证超过 TTL 的票据被拒绝。
func TestTicketExpired(t *testing.T) {
	signer := newTicketSigner("key")
	now := time.Unix(1_700_000_000, 0)
	signer.now = func() time.Time { return now }

	token := signer.issue("u@example.com", testCall("call_1", "restart_pod", "{}"))

	signer.now = func() time.Time { return now.Add(ticketTTL + time.Second) }
	_, err := signer.verify(token)
	assert.ErrorIs(t, err, errTicketExpired)
}

// TestTicketMalformed 验证格式非法的票据被拒绝。
func TestTicketMalformed(t *testing.T) {
	signer := newTicketSigner("key")
	cases := []string{
		"",
		"only-one-part",
		"a.b.c",
		".signature",
		"payload.",
		"!!!.???",
	}
	for _, token := range cases {
		_, err := signer.verify(token)
		assert.ErrorIs(t, err, errTicketMalformed, token)
	}
}

// TestTicketSignatureTampered 验证篡改 payload 后签名校验失败。
func TestTicketSignatureTampered(t *testing.T) {
	signer := newTicketSigner("key")
	token := signer.issue("u@example.com", testCall("call_1", "restart_pod", "{}"))

	parts := strings.Split(token, ".")
	require.Len(t, parts, 2)
	fakePayload := base64.RawURLEncoding.EncodeToString([]byte(`{"email":"attacker@example.com"}`))
	_, err := signer.verify(fakePayload + "." + parts[1])
	assert.ErrorIs(t, err, errTicketSignature)
}

// TestTicketDifferentKey 验证由其他密钥签发的票据被拒绝。
func TestTicketDifferentKey(t *testing.T) {
	token := newTicketSigner("key-a").issue("u@example.com", testCall("call_1", "restart_pod", "{}"))

	_, err := newTicketSigner("key-b").verify(token)
	assert.ErrorIs(t, err, errTicketSignature)
}

// TestTicketSignatureNotBase64 验证签名字段不是合法 base64url 时按格式错误拒绝。
func TestTicketSignatureNotBase64(t *testing.T) {
	signer := newTicketSigner("key")
	token := signer.issue("u@example.com", testCall("call_1", "restart_pod", "{}"))

	payloadPart := strings.Split(token, ".")[0]
	_, err := signer.verify(payloadPart + ".!!!")
	assert.ErrorIs(t, err, errTicketMalformed)
}

// TestTicketPayloadNotJSON 验证签名有效但 payload 不是合法 JSON 时按格式错误拒绝。
func TestTicketPayloadNotJSON(t *testing.T) {
	signer := newTicketSigner("key")
	raw := []byte("not-json")
	token := base64.RawURLEncoding.EncodeToString(raw) + "." +
		base64.RawURLEncoding.EncodeToString(signer.sign(raw))

	_, err := signer.verify(token)
	assert.ErrorIs(t, err, errTicketMalformed)
}

// TestArgsHash 验证参数哈希对不同参数产生不同结果。
func TestArgsHash(t *testing.T) {
	assert.Equal(t, argsHash("{}"), argsHash("{}"))
	assert.NotEqual(t, argsHash("{}"), argsHash(`{"a":1}`))
}
