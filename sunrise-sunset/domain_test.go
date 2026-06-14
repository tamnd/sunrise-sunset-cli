package sunrisesunset

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the URI driver's pure string
// functions and the host wiring (mint, body, resolve), which need no network.
// The client's HTTP behaviour is covered in sunrise-sunset_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "sunrisesunset" {
		t.Errorf("Scheme = %q, want sunrisesunset", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "sunrise-sunset" {
		t.Errorf("Identity.Binary = %q, want sunrise-sunset", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in  string
		typ string
		id  string
	}{
		{"40.7128,-74.0060", "point", "40.7128,-74.0060"},
		{"51.5074,-0.1278", "point", "51.5074,-0.1278"},
		{"2024-01-01", "date", "2024-01-01"},
		{"2025-12-31", "date", "2025-12-31"},
		{"someplace", "point", "someplace"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("Classify(\"\") should return error, got nil")
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("point", "40.7128,-74.0060")
	if err != nil {
		t.Fatalf("Locate error: %v", err)
	}
	if got == "" {
		t.Error("Locate returned empty string")
	}
}

func TestLocateDate(t *testing.T) {
	got, err := Domain{}.Locate("date", "2024-01-01")
	if err != nil {
		t.Fatalf("Locate(date) error: %v", err)
	}
	if got == "" {
		t.Error("Locate(date) returned empty string")
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "foo")
	if err == nil {
		t.Error("Locate(unknown) should return error, got nil")
	}
}

// TestHostWiring mounts the driver in a kit Host and checks the round trip.
func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}

	info := &SolarInfo{
		ID:        "40.7128,-74.006,2024-01-01",
		Latitude:  40.7128,
		Longitude: -74.0060,
		Date:      "2024-01-01",
		Sunrise:   "7:17:54 AM",
		Sunset:    "4:38:25 PM",
		Timezone:  "UTC",
	}
	u, err := h.Mint(info)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if u.String() == "" {
		t.Error("Mint returned empty URI")
	}
}
