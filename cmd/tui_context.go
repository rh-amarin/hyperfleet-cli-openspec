package cmd

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/kube"
	"github.com/rh-amarin/hyperfleet-cli/internal/tui"
)

func tuiContextSnapshot() tui.ContextInfo {
	s, err := loadConfig()
	if err != nil {
		return tui.ContextInfo{}
	}

	env, _ := s.RequireActiveEnvironment()
	info := tui.ContextInfo{
		ActiveEnv:       env,
		APIURL:          s.Get("hyperfleet", "api-url"),
		KubeContext:     s.Get("kubernetes", "context"),
		AutoPortForward: s.Get("hyperfleet", "auto-port-forward") == "true",
	}

	if info.AutoPortForward {
		info.PortForwards = autoPortForwardLines(s)
		return info
	}

	for _, svc := range servicesForArgs(s, nil) {
		err := checkPortForwardConnectivity(svc.name, svc.localPort, s)
		info.PortForwards = append(info.PortForwards, tui.PortForwardLine{
			Name:      svc.name,
			LocalPort: svc.localPort,
			Connected: err == nil,
		})
	}
	return info
}

func autoPortForwardLines(s interface{ Get(string, string) string }) []tui.PortForwardLine {
	specs := []struct {
		name     string
		endpoint func() string
		check    func(int) error
	}{
		{
			name:     "hyperfleet-api",
			endpoint: func() string { return s.Get("hyperfleet", "api-url") },
			check:    kube.CheckAPIConnectivity,
		},
		{
			name:     "maestro-http",
			endpoint: func() string { return s.Get("maestro", "http-endpoint") },
			check:    kube.CheckMaestroHTTPConnectivity,
		},
		{
			name:     "maestro-grpc",
			endpoint: func() string { return s.Get("maestro", "grpc-endpoint") },
			check:    kube.CheckMaestroGRPCConnectivity,
		},
	}

	lines := make([]tui.PortForwardLine, 0, len(specs))
	for _, spec := range specs {
		port, ok := localPortFromEndpoint(spec.endpoint())
		connected := false
		if ok {
			connected = spec.check(port) == nil
		}
		lines = append(lines, tui.PortForwardLine{
			Name:      spec.name,
			LocalPort: port,
			Connected: connected,
		})
	}
	return lines
}

func localPortFromEndpoint(endpoint string) (int, bool) {
	if endpoint == "" {
		return 0, false
	}
	raw := endpoint
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return 0, false
	}
	_, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return 0, false
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		return 0, false
	}
	return port, true
}
