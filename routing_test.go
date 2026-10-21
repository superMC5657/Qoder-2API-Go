package main

import (
	"testing"
	"time"

	"qoder2api/store"
)

// TestUsableExemptsIntlFromQuotaGate 国际版 Free 账号被上游长期标记超额（reset 时间
// 也早已过期），但网关照常服务：超额闸门不能把它踢出候选，否则可用账号会变成误导性
// 的 401。冷却（失败退避/权威扣减）对国际版仍然生效。
func TestUsableExemptsIntlFromQuotaGate(t *testing.T) {
	cfg := store.AutoSwitch{}
	now := time.Now()
	newR := func(snap *store.AccountSnapshot) *router {
		return newRouter(routerDeps{
			snapshot: func(string) *store.AccountSnapshot { return snap },
			config:   func() store.AutoSwitch { return cfg },
		})
	}

	exceeded := &store.AccountSnapshot{IsExceeded: true}
	if !newR(exceeded).usable(store.Account{ID: "g", Region: "global"}, nil, cfg, now, false) {
		t.Error("global account flagged exceeded must remain usable")
	}
	if newR(exceeded).usable(store.Account{ID: "c"}, nil, cfg, now, false) {
		t.Error("cn account flagged exceeded must be excluded")
	}

	cooling := &store.AccountSnapshot{CooldownUntil: now.Add(time.Minute).Unix()}
	if newR(cooling).usable(store.Account{ID: "g", Region: "global"}, nil, cfg, now, false) {
		t.Error("cooldown must still exclude a global account")
	}
}
