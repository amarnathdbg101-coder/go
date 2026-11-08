package featureflag

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestEngine_BooleanFlags(t *testing.T) {
	e := New()
	e.SetBoolean("new-checkout", true)
	e.SetBoolean("legacy-reports", false)

	if !e.IsEnabled("new-checkout", Context{}) {
		t.Errorf("expected 'new-checkout' to be enabled")
	}
	if e.IsEnabled("legacy-reports", Context{}) {
		t.Errorf("expected 'legacy-reports' to be disabled")
	}
	if e.IsEnabled("nonexistent-flag", Context{}) {
		t.Errorf("expected nonexistent flag to be disabled")
	}
}

func TestEngine_PercentageRollout(t *testing.T) {
	e := New()
	e.Register(Flag{
		Key:        "beta-ui",
		Enabled:    true,
		Percentage: 30, // 30% rollout
	})

	// Same user should deterministically get the same result
	ctx1 := Context{UserID: "user-42"}
	firstRes := e.IsEnabled("beta-ui", ctx1)
	secondRes := e.IsEnabled("beta-ui", ctx1)
	if firstRes != secondRes {
		t.Errorf("expected deterministic result for user-42")
	}

	// Over a large population, approximately 30% should be enabled
	total := 1000
	enabledCount := 0
	for i := 0; i < total; i++ {
		if e.IsEnabled("beta-ui", Context{UserID: fmt.Sprintf("u-%d", i)}) {
			enabledCount++
		}
	}

	// Between 25% and 35%
	pct := float64(enabledCount) / float64(total) * 100.0
	if pct < 24.0 || pct > 36.0 {
		t.Errorf("expected rollout ~30%%, got %.2f%% (%d/%d)", pct, enabledCount, total)
	}
}

func TestEngine_AllowedUsersWhitelist(t *testing.T) {
	e := New()
	e.Register(Flag{
		Key:          "admin-panel",
		Enabled:      true,
		Percentage:   0, // 0% general rollout
		AllowedUsers: []string{"admin-1", "tester-2"},
	})

	if !e.IsEnabled("admin-panel", Context{UserID: "admin-1"}) {
		t.Errorf("whitelisted user should have flag enabled")
	}
	if e.IsEnabled("admin-panel", Context{UserID: "regular-user"}) {
		t.Errorf("regular user should not have flag enabled")
	}
}

func TestEngine_AttributeTargeting(t *testing.T) {
	e := New()
	e.Register(Flag{
		Key:     "india-festive-discount",
		Enabled: true,
		RequiredAttrs: map[string]string{
			"country": "IN",
			"tier":    "gold",
		},
	})

	validCtx := Context{
		UserID: "user-1",
		Attributes: map[string]string{
			"country": "IN",
			"tier":    "gold",
		},
	}
	if !e.IsEnabled("india-festive-discount", validCtx) {
		t.Errorf("expected valid context to pass attribute targeting")
	}

	invalidCtx := Context{
		UserID: "user-2",
		Attributes: map[string]string{
			"country": "US",
			"tier":    "gold",
		},
	}
	if e.IsEnabled("india-festive-discount", invalidCtx) {
		t.Errorf("expected invalid country to fail attribute targeting")
	}
}

func TestEngine_TimeWindows(t *testing.T) {
	e := New()

	now := time.Now()
	past := now.Add(-10 * time.Minute)
	future := now.Add(10 * time.Minute)

	// Expired flag
	e.Register(Flag{
		Key:       "expired-promo",
		Enabled:   true,
		StartTime: &past,
		EndTime:   &past,
	})

	if e.IsEnabled("expired-promo", Context{}) {
		t.Errorf("expired flag should be disabled")
	}

	// Future flag
	e.Register(Flag{
		Key:       "future-promo",
		Enabled:   true,
		StartTime: &future,
	})

	if e.IsEnabled("future-promo", Context{}) {
		t.Errorf("future flag should not be enabled yet")
	}
}

func TestEngine_ConcurrentAccess(t *testing.T) {
	e := New()
	e.SetBoolean("live-feature", true)

	var wg sync.WaitGroup
	const routines = 30

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				e.SetPercentage("live-feature", (id*10)%100)
			} else {
				e.IsEnabled("live-feature", Context{UserID: fmt.Sprintf("u-%d", id)})
			}
		}(i)
	}

	wg.Wait()
}
