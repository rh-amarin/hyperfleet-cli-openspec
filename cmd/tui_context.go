package cmd

import (
	"github.com/rh-amarin/hyperfleet-cli/internal/tui"
)

func tuiContextSnapshot() tui.ContextInfo {
	s, err := loadConfig()
	if err != nil {
		return tui.ContextInfo{}
	}

	env, _ := s.RequireActiveEnvironment()
	info := tui.ContextInfo{
		ActiveEnv:   env,
		APIURL:      s.Get("hyperfleet", "api-url"),
		KubeContext: s.Get("kubernetes", "context"),
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
