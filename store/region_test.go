package store

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDeploymentRegionStampedOnLoad QODER_REGION=intl 时，既有空 region 账号与裸
// PAT 迁移账号都要补成 global，否则 intl 的 PAT 会被发往国内网关。
func TestDeploymentRegionStampedOnLoad(t *testing.T) {
	body := `{"host":"127.0.0.1","port":10081,"pat":"intl","password":"pw","api_keys":[],` +
		`"accounts":[{"id":"a1","pat":"pt-old","note":"旧账号","created_at":1}]}`
	write := func(t *testing.T) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "data.json")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Setenv("QODER_REGION", "intl")
	s, err := New(write(t))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	accs := s.ListAccounts()
	if len(accs) != 2 {
		t.Fatalf("expected 2 accounts (existing + migrated PAT), got %d", len(accs))
	}
	for _, acc := range accs {
		if acc.Region != "global" {
			t.Errorf("account %s region = %q, want global", acc.ID, acc.Region)
		}
	}

	// 国内部署保持空 region（CN 默认），不写入 global。
	t.Setenv("QODER_REGION", "")
	s2, err := New(write(t))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	for _, acc := range s2.ListAccounts() {
		if acc.Region != "" {
			t.Errorf("account %s region = %q, want empty (CN default)", acc.ID, acc.Region)
		}
	}
}

// TestGetIdeIdentity 只有 securityOauthToken 存在时才算有身份，字段按 jobToken 命名。
func TestGetIdeIdentity(t *testing.T) {
	s := tempStore(t)
	if _, ok := s.GetIdeIdentity(); ok {
		t.Fatal("empty config must not report an IDE identity")
	}
	s.mu.Lock()
	s.config.SecurityOauthToken = "sot"
	s.config.RefreshToken = "rt"
	s.config.UID = "uid-1"
	s.config.Nickname = "tester"
	s.config.ExpireTime = 1700000000000
	s.mu.Unlock()

	id, ok := s.GetIdeIdentity()
	if !ok {
		t.Fatal("identity not reported")
	}
	if id["securityOauthToken"] != "sot" || id["refreshToken"] != "rt" || id["id"] != "uid-1" || id["name"] != "tester" {
		t.Errorf("unexpected identity: %+v", id)
	}
	if id["expireTime"] != float64(1700000000000) {
		t.Errorf("expireTime = %v", id["expireTime"])
	}
}
