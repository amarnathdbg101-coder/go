# go-featureflag 🚩

`go-featureflag` ek lightweight, thread-safe dynamic feature flag aur canary rollout engine hai Go ke liye. Ye production me features ko bina restart ya redeployment ke enable/disable karne, gradual percentage rollouts (A/B testing), user attribute targeting, aur time-window promotions ke liye banaya gaya hai.

---

## ✨ Features

- **Binary Boolean Flags**: Fast global On/Off switch (`SetBoolean("new-checkout", true)`).
- **Deterministic Canary Rollouts**: FNV-1a hashing algorithm use karke consistent percentage buckets (e.g. 20% users) assign karta hai bina database state store kiye.
- **Whitelist Overrides**: Specific VIP users, QA testers, ya internal admins ke liye direct bypass (`AllowedUsers`).
- **User Attribute Segmentation**: Region, subscription tier, ya role ke basis par flags toggle karein (`RequiredAttrs`).
- **Time-Windowed Activations**: Automated start time aur expiry end time for flash sales and festival offers.
- **Thread-Safe Runtime Updates**: Production traffic chalte hue dynamically flags update karein (`sync.RWMutex`).
- **Zero Third-Party Dependencies**: Pure Go Standard Library.

---

## 🚀 Quick Start

### Installation
```bash
go get github.com/shopsilo/go-featureflag
```

### Complete Example
```go
package main

import (
	"fmt"

	"github.com/shopsilo/go-featureflag"
)

func main() {
	engine := featureflag.New()

	// 1. Simple Boolean Flag
	engine.SetBoolean("maintenance-mode", false)

	// 2. 25% Canary Rollout with Admin Whitelist
	engine.Register(featureflag.Flag{
		Key:          "new-recommendation-engine",
		Enabled:      true,
		Percentage:   25, // 25% of users
		AllowedUsers: []string{"dev-lead", "qa-tester"},
	})

	// 3. User Attribute Targeting
	engine.Register(featureflag.Flag{
		Key:     "diwali-special-offer",
		Enabled: true,
		RequiredAttrs: map[string]string{
			"country": "IN",
			"tier":    "gold",
		},
	})

	// Evaluation
	userCtx := featureflag.Context{
		UserID: "user-1002",
		Attributes: map[string]string{
			"country": "IN",
			"tier":    "gold",
		},
	}

	if engine.IsEnabled("diwali-special-offer", userCtx) {
		fmt.Println("Special offer applied!")
	}

	if engine.IsEnabled("new-recommendation-engine", userCtx) {
		fmt.Println("User is part of 25% canary test group!")
	}
}
```

---

## 🧪 Testing

```bash
go test -v -count=1 ./...
```
Output:
```
=== RUN   TestEngine_BooleanFlags
--- PASS: TestEngine_BooleanFlags (0.00s)
=== RUN   TestEngine_PercentageRollout
--- PASS: TestEngine_PercentageRollout (0.00s)
=== RUN   TestEngine_AllowedUsersWhitelist
--- PASS: TestEngine_AllowedUsersWhitelist (0.00s)
=== RUN   TestEngine_AttributeTargeting
--- PASS: TestEngine_AttributeTargeting (0.00s)
=== RUN   TestEngine_TimeWindows
--- PASS: TestEngine_TimeWindows (0.00s)
=== RUN   TestEngine_ConcurrentAccess
--- PASS: TestEngine_ConcurrentAccess (0.00s)
PASS
ok  	go-featureflag	0.487s
```

---

## 📄 License
MIT License.
