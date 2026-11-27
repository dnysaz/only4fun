package proxytrust

import "testing"

func TestTrustDefaultsToLoopback(t *testing.T) {
	trust, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, addr := range []string{"127.0.0.1:8080", "[::1]:8080"} {
		if !trust.ContainsRemote(addr) {
			t.Errorf("loopback %s should be trusted", addr)
		}
	}
	for _, addr := range []string{"203.0.113.9:1234", "10.0.0.1:1234"} {
		if trust.ContainsRemote(addr) {
			t.Errorf("remote %s should not be trusted by default", addr)
		}
	}
}

func TestTrustParsesBareIPsAndCIDRs(t *testing.T) {
	trust, err := New([]string{"10.0.0.8", "173.245.48.0/20"})
	if err != nil {
		t.Fatal(err)
	}

	if !trust.ContainsRemote("10.0.0.8:5555") {
		t.Error("bare IP should be trusted")
	}
	if trust.ContainsRemote("10.0.0.9:5555") {
		t.Error("different IP should not be trusted")
	}
	if !trust.ContainsRemote("173.245.48.42:1234") {
		t.Error("CIDR member should be trusted")
	}
	if trust.ContainsRemote("192.0.2.1:1234") {
		t.Error("outside CIDR should not be trusted")
	}
}
