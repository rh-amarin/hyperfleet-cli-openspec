package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/kube"
)

func tuiTogglePortForwards() (string, error) {
	s, err := loadConfig()
	if err != nil {
		return "", err
	}

	services := servicesForArgs(s, nil)
	anyDown := false
	for _, svc := range services {
		if err := checkPortForwardConnectivity(svc.name, svc.localPort, s); err != nil {
			anyDown = true
			break
		}
	}
	if anyDown {
		return tuiStartPortForwards(s)
	}
	return tuiStopPortForwards()
}

func tuiStartPortForwards(s *config.Store) (string, error) {
	kubeconfig := resolvedKubeconfig(s)
	kubeCtx := s.Get("kubernetes", "context")
	services := servicesForArgs(s, nil)

	var lines []string
	for _, svc := range services {
		sr, err := kube.StartPortForward(kubeconfig, svc.namespace, svc.name, svc.serviceName, svc.podPattern, svc.localPort, svc.remotePort, kubeCtx)
		if err != nil {
			lines = append(lines, fmt.Sprintf("[ERROR] %s: %v", svc.name, err))
			continue
		}
		lines = append(lines, fmt.Sprintf("Started %s (%s): localhost:%d (pid %d)",
			sr.Name, formatPortForwardTarget(sr), sr.LocalPort, sr.PID))
	}
	time.Sleep(time.Second)
	if len(lines) == 0 {
		return "[INFO] Port-forwards started", nil
	}
	return "[INFO] " + strings.Join(lines, " · "), nil
}

func tuiStopPortForwards() (string, error) {
	pfs, _ := kube.ListPortForwards()
	if len(pfs) == 0 {
		return "[INFO] No port-forwards running", nil
	}

	var stopped int
	var errs []string
	for _, pf := range pfs {
		if err := kube.StopPortForward(pf.Name); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", pf.Name, err))
		} else {
			stopped++
		}
	}
	if len(errs) > 0 {
		return fmt.Sprintf("[ERROR] stopped %d: %s", stopped, strings.Join(errs, "; ")),
			fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return fmt.Sprintf("[INFO] Stopped %d port-forward(s)", stopped), nil
}
