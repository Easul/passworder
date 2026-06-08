package embedded

import "testing"

func TestListenerNetworkUsesIPv4ForIPv4Hosts(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{name: "loopback ipv4", host: "127.0.0.1", want: "tcp4"},
		{name: "wildcard ipv4", host: "0.0.0.0", want: "tcp4"},
		{name: "empty host", host: "", want: "tcp"},
		{name: "hostname", host: "localhost", want: "tcp"},
		{name: "loopback ipv6", host: "::1", want: "tcp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listenerNetwork(tt.host); got != tt.want {
				t.Fatalf("listenerNetwork(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}
