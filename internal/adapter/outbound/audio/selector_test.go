package audio

import (
	"testing"
	"gosper/internal/domain"
)

func TestResolveDeviceID(t *testing.T) {
	devs := []domain.Device{{ID:"index:0", Name:"MacBook Pro Microphone"}, {ID:"index:1", Name:"External USB Mic"}}
	cases := []struct{ sel string; want string }{
		{"index:1", "index:1"},
		{"external usb mic", "index:1"},
		{"macbook", "index:0"},
		{"usb", "index:1"},
	}
	for _, c := range cases {
		got := ResolveDeviceID(devs, c.sel)
		if got != c.want { t.Fatalf("sel %q: got %q want %q", c.sel, got, c.want) }
	}
}
