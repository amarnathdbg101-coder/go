package useragent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParse_DesktopBrowsers(t *testing.T) {
	// Chrome on Windows
	chromeUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.60 Safari/537.36"
	res := Parse(chromeUA)

	if res.Browser != "Chrome" || res.OS != "Windows" || res.Device != DeviceDesktop || !res.IsDesktop {
		t.Errorf("Chrome/Win parse mismatch: %+v", res)
	}
	if res.BrowserVersion != "124.0.6367.60" {
		t.Errorf("expected Chrome version 124.0.6367.60, got %s", res.BrowserVersion)
	}

	// Firefox on Linux
	ffUA := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0"
	resFF := Parse(ffUA)
	if resFF.Browser != "Firefox" || resFF.OS != "Linux" || !resFF.IsDesktop {
		t.Errorf("Firefox/Linux parse mismatch: %+v", resFF)
	}

	// Edge on Windows
	edgeUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.2478.51"
	resEdge := Parse(edgeUA)
	if resEdge.Browser != "Edge" || resEdge.OS != "Windows" || !resEdge.IsDesktop {
		t.Errorf("Edge/Win parse mismatch: %+v", resEdge)
	}
}

func TestParse_MobileAndTablet(t *testing.T) {
	// iPhone Mobile
	iphoneUA := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Mobile/15E148 Safari/604.1"
	resIPhone := Parse(iphoneUA)
	if resIPhone.OS != "iOS" || resIPhone.Device != DeviceMobile || !resIPhone.IsMobile || resIPhone.Browser != "Safari" {
		t.Errorf("iPhone parse mismatch: %+v", resIPhone)
	}

	// iPad Tablet
	ipadUA := "Mozilla/5.0 (iPad; CPU OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	resIPad := Parse(ipadUA)
	if resIPad.OS != "iPadOS" || resIPad.Device != DeviceTablet || !resIPad.IsTablet {
		t.Errorf("iPad parse mismatch: %+v", resIPad)
	}

	// Android Mobile
	androidUA := "Mozilla/5.0 (Linux; Android 14; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6367.54 Mobile Safari/537.36"
	resAndroid := Parse(androidUA)
	if resAndroid.OS != "Android" || resAndroid.Device != DeviceMobile || !resAndroid.IsMobile {
		t.Errorf("Android mobile parse mismatch: %+v", resAndroid)
	}
}

func TestParse_Bots(t *testing.T) {
	bots := []string{
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"curl/7.88.1",
		"PostmanRuntime/7.36.0",
		"facebookexternalhit/1.1",
	}

	for _, botUA := range bots {
		res := Parse(botUA)
		if !res.IsBot || res.Device != DeviceBot {
			t.Errorf("expected bot detection for %s, got %+v", botUA, res)
		}
	}
}

func TestFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15")

	res := FromRequest(req)
	if res.Browser != "Safari" || res.OS != "macOS" || !res.IsDesktop {
		t.Errorf("FromRequest mismatch: %+v", res)
	}

	// Empty request
	resEmpty := FromRequest(nil)
	if resEmpty.Device != DeviceUnknown {
		t.Errorf("expected DeviceUnknown for nil request")
	}
}
