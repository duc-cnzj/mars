package biz

// oidc_test.go 覆盖 OIDC 授权码换发链路：defaultAuthProvider 适配器（换发/验签）
// 与 authBiz.Exchange 编排（遍历 provider/跳过 nil/claims 解码/错误聚合）。

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newOIDCProvider spins up a minimal OIDC discovery server (discovery + JWKS)
// returning the issuer and the private key used to sign ID tokens.
func newOIDCProvider(t *testing.T) (issuer string, key *rsa.PrivateKey) {
	t.Helper()
	var err error
	key, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	jwks := map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"use": "sig",
			"alg": "RS256",
			"kid": "test-key",
			"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}},
	}
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                ts.URL,
			"jwks_uri":                              ts.URL + "/keys",
			"authorization_endpoint":                ts.URL + "/auth",
			"token_endpoint":                        ts.URL + "/token",
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/keys", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	})
	t.Cleanup(ts.Close)
	return ts.URL, key
}

// signIDToken builds a compact JWS (RS256) with the given issuer/audience.
func signIDToken(t *testing.T, key *rsa.PrivateKey, issuer, aud string) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT","kid":"test-key"}`))
	now := time.Now()
	payload, err := json.Marshal(map[string]any{
		"iss": issuer,
		"sub": "subject",
		"aud": aud,
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
	})
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// newTokenServer starts a token endpoint that returns an id_token signed for aud.
func newTokenServer(t *testing.T, key *rsa.PrivateKey, issuer, aud string) (oauth2.Config, func()) {
	t.Helper()
	idt := signIDToken(t, key, issuer, aud)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"access","token_type":"Bearer","id_token":%q}`, idt)
	}))
	cfg := oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint:     oauth2.Endpoint{TokenURL: ts.URL},
		Scopes:       []string{"openid"},
	}
	return cfg, ts.Close
}

func TestDefaultAuthProvider_Exchange_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"access","token_type":"Bearer","id_token":"raw-id-token"}`)
	}))
	defer ts.Close()
	cfg := oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint:     oauth2.Endpoint{TokenURL: ts.URL},
		Scopes:       []string{"openid"},
	}
	p := NewDefaultAuthProvider(cfg, nil)
	token, err := p.Exchange(context.TODO(), "auth-code")
	assert.NoError(t, err)
	assert.Equal(t, "raw-id-token", token)
}

func TestDefaultAuthProvider_Exchange_MissingIDToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"access","token_type":"Bearer"}`)
	}))
	defer ts.Close()
	cfg := oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint:     oauth2.Endpoint{TokenURL: ts.URL},
		Scopes:       []string{"openid"},
	}
	p := NewDefaultAuthProvider(cfg, nil)
	_, err := p.Exchange(context.TODO(), "auth-code")
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "auth-code")
}

func TestDefaultAuthProvider_Exchange_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer ts.Close()
	cfg := oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint:     oauth2.Endpoint{TokenURL: ts.URL},
		Scopes:       []string{"openid"},
	}
	p := NewDefaultAuthProvider(cfg, nil)
	_, err := p.Exchange(context.TODO(), "auth-code")
	assert.Error(t, err)
}

func TestDefaultAuthProvider_Verify_Success(t *testing.T) {
	issuer, key := newOIDCProvider(t)
	provider, err := oidc.NewProvider(context.TODO(), issuer)
	if err != nil {
		t.Fatalf("oidc.NewProvider: %v", err)
	}
	p := NewDefaultAuthProvider(oauth2.Config{ClientID: "client-id"}, provider)
	idt := signIDToken(t, key, issuer, "client-id")
	parsed, err := p.Verify(context.TODO(), idt)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	var claims map[string]any
	assert.NoError(t, parsed.Claims(&claims))
	assert.Equal(t, "subject", claims["sub"])
}

func TestDefaultAuthProvider_Verify_Fail(t *testing.T) {
	issuer, key := newOIDCProvider(t)
	provider, err := oidc.NewProvider(context.TODO(), issuer)
	if err != nil {
		t.Fatalf("oidc.NewProvider: %v", err)
	}
	p := NewDefaultAuthProvider(oauth2.Config{ClientID: "client-id"}, provider)
	// wrong audience: verifier configured for client-id, token issued for other-aud
	idt := signIDToken(t, key, issuer, "other-aud")
	_, err = p.Verify(context.TODO(), idt)
	assert.Error(t, err)
}

// newTestAuthBiz 构造 authBiz 测试实例：Exchange 编排只依赖 oidcConfig 与 logger，
// auth 与生效角色解析器在该路径不被触及，可安全传 nil。
func newTestAuthBiz(oidcConfig func() OidcConfig) *authBiz {
	return NewAuthBiz(nil, fakeAuthConfigProvider{oidcConfig: oidcConfig}, nil, mlog.NewForConfig(nil)).(*authBiz)
}

func TestAuthBiz_Exchange_Success(t *testing.T) {
	issuer, key := newOIDCProvider(t)
	provider, err := oidc.NewProvider(context.TODO(), issuer)
	if err != nil {
		t.Fatalf("oidc.NewProvider: %v", err)
	}
	cfg, closeTS := newTokenServer(t, key, issuer, "client-id")
	defer closeTS()

	b := newTestAuthBiz(func() OidcConfig {
		return OidcConfig{
			"test": OidcConfigItem{
				Provider:           provider,
				Config:             cfg,
				EndSessionEndpoint: "https://logout.example",
			},
		}
	})
	u, err := b.Exchange(context.TODO(), "auth-code")
	assert.NoError(t, err)
	if assert.NotNil(t, u) {
		assert.Equal(t, "subject", u.ID)
		assert.Equal(t, "https://logout.example", u.LogoutUrl)
	}
}

func TestAuthBiz_Exchange_AllProvidersNil(t *testing.T) {
	b := newTestAuthBiz(func() OidcConfig {
		return OidcConfig{
			"a": OidcConfigItem{Provider: nil, Config: oauth2.Config{}},
			"b": OidcConfigItem{Provider: nil, Config: oauth2.Config{}},
		}
	})
	_, err := b.Exchange(context.TODO(), "auth-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid code")
}

func TestAuthBiz_Exchange_VerifyFail(t *testing.T) {
	issuer, key := newOIDCProvider(t)
	provider, err := oidc.NewProvider(context.TODO(), issuer)
	if err != nil {
		t.Fatalf("oidc.NewProvider: %v", err)
	}
	// id_token signed for a different audience -> Verify rejects it.
	cfg, closeTS := newTokenServer(t, key, issuer, "other-aud")
	defer closeTS()

	b := newTestAuthBiz(func() OidcConfig {
		return OidcConfig{
			"test": OidcConfigItem{Provider: provider, Config: cfg},
		}
	})
	_, err = b.Exchange(context.TODO(), "auth-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid code")
}

func TestAuthBiz_Exchange_ClaimsDecodeError(t *testing.T) {
	issuer, key := newOIDCProvider(t)
	provider, err := oidc.NewProvider(context.TODO(), issuer)
	if err != nil {
		t.Fatalf("oidc.NewProvider: %v", err)
	}
	// Verify 通过（go-oidc 不校验 email_verified 类型），但 OidcClaims.EmailVerified
	// 是普通 bool，解码字符串失败 → 跳过该 provider，最终 "invalid code"。
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT","kid":"test-key"}`))
	now := time.Now()
	payload, err := json.Marshal(map[string]any{
		"iss":            issuer,
		"sub":            "subject",
		"aud":            "client-id",
		"exp":            now.Add(time.Hour).Unix(),
		"iat":            now.Unix(),
		"email_verified": "not-a-bool",
	})
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	idt := signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"access","token_type":"Bearer","id_token":%q}`, idt)
	}))
	defer ts.Close()
	cfg := oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint:     oauth2.Endpoint{TokenURL: ts.URL},
		Scopes:       []string{"openid"},
	}

	b := newTestAuthBiz(func() OidcConfig {
		return OidcConfig{
			"test": OidcConfigItem{Provider: provider, Config: cfg},
		}
	})
	_, err = b.Exchange(context.TODO(), "auth-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid code")
}

func TestAuthBiz_Exchange_CodeNotEchoed(t *testing.T) {
	// 全部 provider 为 nil → 聚合失败；断言错误信息不回显一次性授权码。
	b := newTestAuthBiz(func() OidcConfig {
		return OidcConfig{"a": OidcConfigItem{Provider: nil}}
	})
	code := "auth-code-SECRET-abc123"
	_, err := b.Exchange(context.TODO(), code)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.NotContains(t, err.Error(), code)
}

func TestAuthBiz_Exchange_SecondProviderSucceeds(t *testing.T) {
	// 第一个 provider 换发失败，第二个成功 → 编排继续并返回成功（验证 continue 语义）。
	issuer, key := newOIDCProvider(t)
	provider, err := oidc.NewProvider(context.TODO(), issuer)
	if err != nil {
		t.Fatalf("oidc.NewProvider: %v", err)
	}
	cfg, closeTS := newTokenServer(t, key, issuer, "client-id")
	defer closeTS()

	b := newTestAuthBiz(func() OidcConfig {
		return OidcConfig{
			"first":  OidcConfigItem{Provider: nil}, // skip
			"second": OidcConfigItem{Provider: provider, Config: cfg},
		}
	})
	u, err := b.Exchange(context.TODO(), "auth-code")
	assert.NoError(t, err)
	if assert.NotNil(t, u) {
		assert.Equal(t, "subject", u.ID)
	}
}

// TestAuthBiz_Exchange_EmptyConfig 覆盖 OidcConfig 为空 map 的聚合失败路径。
func TestAuthBiz_Exchange_EmptyConfig(t *testing.T) {
	b := newTestAuthBiz(func() OidcConfig { return OidcConfig{} })
	_, err := b.Exchange(context.TODO(), "auth-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid code")
}

// TestAuthBiz_Exchange_ExchangeError 覆盖单个 provider 的 cfg.Exchange 失败后
// 跳过继续（token 端点返回 400），最终聚合失败。
func TestAuthBiz_Exchange_ExchangeError(t *testing.T) {
	issuer, _ := newOIDCProvider(t)
	provider, err := oidc.NewProvider(context.TODO(), issuer)
	if err != nil {
		t.Fatalf("oidc.NewProvider: %v", err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer ts.Close()
	cfg := oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint:     oauth2.Endpoint{TokenURL: ts.URL},
		Scopes:       []string{"openid"},
	}

	b := newTestAuthBiz(func() OidcConfig {
		return OidcConfig{
			"test": OidcConfigItem{Provider: provider, Config: cfg},
		}
	})
	_, err = b.Exchange(context.TODO(), "auth-code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid code")
}
