package bridge

import (
	"context"
	"testing"
	"time"

	"qoder2api/auth"
)

// TestIdeIdentityBootstrapsWithoutExchange 桌面 IDE 身份直接装成会话：冷启动不发起
// jobToken 交换，也拒绝 refresh——凭据由 all2api 管理，过期只能重登 IDE。
func TestIdeIdentityBootstrapsWithoutExchange(t *testing.T) {
	exp := time.Now().Add(6 * time.Hour).UnixMilli()
	b := NewOpenAiBridge("intl", auth.Global, map[string]interface{}{
		"name":               "tester",
		"id":                 "uid-1",
		"userType":           "personal_standard",
		"securityOauthToken": "sot-1",
		"refreshToken":       "rt-1",
		"expireTime":         float64(exp),
	})
	if err := b.bootstrapSession(context.Background()); err != nil {
		t.Fatalf("bootstrap with IDE identity: %v", err)
	}
	if b.currentSess() == nil {
		t.Fatal("session not built from IDE identity")
	}
	id := b.currentIdentity()
	if id == nil || id.Name != "tester" || id.UID != "uid-1" || id.SecurityOauthToken != "sot-1" {
		t.Fatalf("identity not installed: %+v", id)
	}
	if got := b.currentSecurityOauth(); got != "sot-1" {
		t.Errorf("securityOauth = %q, want sot-1", got)
	}
	if err := b.doRenew(context.Background(), true); err == nil {
		t.Fatal("renewal must be refused for IDE-identity sessions")
	}
}

// TestPatBridgeStillExchanges 没有身份时保持原有 PAT 交换路径（不误判为 IDE 身份）。
func TestPatBridgeStillExchanges(t *testing.T) {
	b := NewOpenAiBridge("pt-real", auth.CN)
	if b.ideIdentity != nil {
		t.Fatal("bridge without identity must not carry one")
	}
}
