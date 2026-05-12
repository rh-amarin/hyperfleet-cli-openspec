// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/kube"
	"github.com/spf13/cobra"
)

var kubeConfigFlag string

// kubeCmd is the top-level group for Kubernetes operations.
var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "Perform Kubernetes operations without requiring kubectl",
	Long: `Perform Kubernetes operations without requiring kubectl.

Subcommands: port-forward, curl, debug.`,
}

// portForwardCmd groups port-forward subcommands.
var portForwardCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Manage port-forwards to HyperFleet in-cluster services",
}

// pfStartCmd starts one or all predefined port-forwards.
var pfStartCmd = &cobra.Command{
	Use:   "start [name] [localPort:remotePort]",
	Short: "Start port-forward(s) to predefined services",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		kubeconfig := resolvedKubeconfig(s)
		kubeCtx := s.Get("kubernetes", "context")
		if ctxName, err := kube.ResolvedContext(kubeconfig, kubeCtx); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "[WARN] Could not resolve kubernetes context: %v\n", err)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Kubernetes context: %s\n", ctxName)
		}
		services := servicesForArgs(s, args)
		for _, svc := range services {
			sr, err := kube.StartPortForward(kubeconfig, svc.namespace, svc.name, svc.podPattern, svc.localPort, svc.remotePort, kubeCtx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] %s: %v\n", svc.name, err)
				continue
			}
			if sr.PodName != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Started %s (%s/%s): localhost:%d → %d (pid %d)\n",
					sr.Name, sr.Namespace, sr.PodName, sr.LocalPort, sr.RemotePort, sr.PID)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Started %s (%s): localhost:%d → %d (pid %d)\n",
					sr.Name, sr.Namespace, sr.LocalPort, sr.RemotePort, sr.PID)
			}
		}
		time.Sleep(time.Second)
		printPortForwardStatus(cmd.OutOrStdout(), s)
		return nil
	},
}

// pfStopCmd stops one or all running port-forwards.
var pfStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop port-forward(s)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if err := kube.StopPortForward(args[0]); err != nil {
				return fmt.Errorf("[ERROR] %v", err)
			}
			fmt.Printf("[INFO] Stopped %s\n", args[0])
			return nil
		}
		pfs, _ := kube.ListPortForwards()
		if len(pfs) == 0 {
			fmt.Println("[INFO] No port-forwards running.")
			return nil
		}
		for _, pf := range pfs {
			if err := kube.StopPortForward(pf.Name); err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] %s: %v\n", pf.Name, err)
			} else {
				fmt.Printf("[INFO] Stopped %s\n", pf.Name)
			}
		}
		return nil
	},
}

// pfStatusCmd shows the status of all port-forwards.
var pfStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show port-forward status",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		kubeconfig := resolvedKubeconfig(s)
		kubeCtx := s.Get("kubernetes", "context")
		if ctxName, err := kube.ResolvedContext(kubeconfig, kubeCtx); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "[WARN] Could not resolve kubernetes context: %v\n", err)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Kubernetes context: %s\n", ctxName)
		}
		printPortForwardStatus(cmd.OutOrStdout(), s)
		return nil
	},
}

// pfDaemonCmd is the hidden daemon invoked by StartPortForward as a detached subprocess.
var pfDaemonCmd = &cobra.Command{
	Use:    "_daemon <kubeconfig> <namespace> <podName> <localPort> <remotePort> <context>",
	Hidden: true,
	Args:   cobra.ExactArgs(6),
	RunE: func(cmd *cobra.Command, args []string) error {
		localPort, err := strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("invalid localPort: %v", err)
		}
		remotePort, err := strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid remotePort: %v", err)
		}
		return kube.RunPortForwardDaemon(args[0], args[1], args[2], localPort, remotePort, args[5])
	},
}

// kubeCurlCmd runs curl from inside an in-cluster pod.
var kubeCurlCmd = &cobra.Command{
	Use:   "curl [--] [curl-flags...] <url>",
	Short: "Run curl from inside the Kubernetes cluster",
	Long: `Execute a curl command from a persistent in-cluster pod (hf-curl).

Curl flags starting with '-' must be preceded by '--' to avoid Cobra flag parsing.

Example:
  hf kube curl -- -H 'Content-Type: application/json' https://my-service/`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("[ERROR] usage: hf kube curl [--] <curl-args>")
		}
		s, err := loadConfig()
		if err != nil {
			return err
		}
		namespace := s.Get("hyperfleet", "namespace")
		kubeCtx := s.Get("kubernetes", "context")
		return kube.RunCurlPod(context.Background(), resolvedKubeconfig(s), namespace, kubeCtx, args, os.Stdout)
	},
}

// kubeDebugCmd creates a debug pod from a deployment template and execs into it.
var kubeDebugCmd = &cobra.Command{
	Use:   "debug <partial-deployment-name>",
	Short: "Create and exec into a debug pod from a deployment template",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		namespace := s.Get("hyperfleet", "namespace")
		kubeconfig := resolvedKubeconfig(s)
		kubeCtx := s.Get("kubernetes", "context")
		cs, err := kube.NewClientset(kubeconfig, kubeCtx)
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}
		config, err := kube.BuildConfig(kubeconfig, kubeCtx)
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}
		ctx := context.Background()
		fmt.Fprintf(os.Stderr, "[INFO] Creating debug pod from deployment matching %q...\n", args[0])
		podName, err := kube.CreateDebugPod(ctx, cs, namespace, args[0])
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}
		fmt.Printf("[INFO] Debug pod ready: %s\n", podName)
		fmt.Printf("[INFO] To re-attach: kubectl exec -it -n %s %s -- /bin/sh\n", namespace, podName)
		return kube.ExecShell(ctx, cs, config, namespace, podName)
	},
}

func init() {
	kubeCmd.PersistentFlags().StringVar(&kubeConfigFlag, "kubeconfig", "",
		"path to kubeconfig (default: KUBECONFIG env or ~/.kube/config)")

	portForwardCmd.AddCommand(pfStartCmd, pfStopCmd, pfStatusCmd, pfDaemonCmd)
	kubeCmd.AddCommand(portForwardCmd, kubeCurlCmd, kubeDebugCmd)
	rootCmd.AddCommand(kubeCmd)
}

// ---- helpers ----

type configGetter interface {
	Get(section, key string) string
}

// printPortForwardStatus renders the live port-forward status table to w using
// protocol-aware connectivity checks per predefined service.
func printPortForwardStatus(w io.Writer, s configGetter) {
	pfs, _ := kube.ListPortForwards()
	if len(pfs) == 0 {
		fmt.Fprintln(w, "No port-forwards tracked.")
		return
	}
	for _, pf := range pfs {
		err := checkPortForwardConnectivity(pf.Name, pf.LocalPort, s)
		connected := err == nil

		tick := "\033[32m✓\033[0m"
		suffix := ""
		pid := pf.PID
		if connected {
			if activePID, pidErr := kube.PIDForPort(pf.LocalPort); pidErr == nil {
				pid = activePID
			}
		} else {
			tick = "\033[31m✗\033[0m"
			suffix = " [NOT CONNECTED]"
		}
		fmt.Fprintf(w, "  %s %s - localhost:%d (PID: %d)%s\n",
			tick, pf.Name, pf.LocalPort, pid, suffix)
	}
}

// checkPortForwardConnectivity dispatches to the appropriate protocol checker by service name.
func checkPortForwardConnectivity(name string, localPort int, s configGetter) error {
	switch name {
	case "hyperfleet-api":
		return kube.CheckAPIConnectivity(localPort)
	case "postgresql":
		host := s.Get("database", "host")
		dbname := s.Get("database", "name")
		user := s.Get("database", "user")
		password := s.Get("database", "password")
		return kube.CheckPostgresConnectivity(localPort, host, dbname, user, password)
	case "maestro-http":
		return kube.CheckMaestroHTTPConnectivity(localPort)
	case "maestro-grpc":
		return kube.CheckMaestroGRPCConnectivity(localPort)
	default:
		if kube.IsPortListening(localPort) {
			return nil
		}
		return fmt.Errorf("port %d not listening", localPort)
	}
}

// resolvedKubeconfig returns the kubeconfig path from the persistent flag or empty string
// (kube.BuildConfig resolves KUBECONFIG env / ~/.kube/config when empty).
func resolvedKubeconfig(_ interface{ Get(string, string) string }) string {
	return kubeConfigFlag
}

type serviceSpec struct {
	name       string
	podPattern string
	namespace  string
	localPort  int
	remotePort int
}

// servicesForArgs resolves which services to port-forward based on CLI args.
func servicesForArgs(s interface{ Get(string, string) string }, args []string) []serviceSpec {
	maestroNS := s.Get("maestro", "namespace")
	hfNS := s.Get("hyperfleet", "namespace")
	all := []serviceSpec{
		{"hyperfleet-api", "hyperfleet-api", hfNS,
			portVal(s, "port-forward", "api-port", 8000), 8000},
		{"postgresql", "postgresql", hfNS,
			portVal(s, "port-forward", "pg-port", 5432), 5432},
		{"maestro-http", "maestro", maestroNS,
			portVal(s, "port-forward", "maestro-http-port", 8100),
			portVal(s, "port-forward", "maestro-http-remote-port", 8000)},
		{"maestro-grpc", "maestro", maestroNS,
			portVal(s, "port-forward", "maestro-grpc-port", 8090),
			portVal(s, "port-forward", "maestro-grpc-remote-port", 8090)},
	}
	if len(args) == 0 {
		return all
	}
	name := args[0]
	for _, svc := range all {
		if svc.name == name {
			if len(args) >= 2 {
				lp, rp, err := parsePortSpec(args[1])
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					return nil
				}
				svc.localPort = lp
				svc.remotePort = rp
			}
			return []serviceSpec{svc}
		}
	}
	// Generic: treat first arg as pod pattern + second as port spec.
	if len(args) >= 2 {
		lp, rp, err := parsePortSpec(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return nil
		}
		ns := s.Get("hyperfleet", "namespace")
		return []serviceSpec{{name: name, podPattern: name, namespace: ns, localPort: lp, remotePort: rp}}
	}
	fmt.Fprintf(os.Stderr, "[ERROR] Unknown service %q. Known: hyperfleet-api, postgresql, maestro-http, maestro-grpc\n", name)
	return nil
}

func parsePortSpec(spec string) (int, int, error) {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("[ERROR] Invalid port spec %q. Expected <localPort>:<remotePort>", spec)
	}
	lp, err1 := strconv.Atoi(parts[0])
	rp, err2 := strconv.Atoi(parts[1])
	if err1 != nil || lp < 1 || lp > 65535 {
		return 0, 0, fmt.Errorf("[ERROR] Invalid port '%s'. Must be an integer between 1 and 65535.", parts[0])
	}
	if err2 != nil || rp < 1 || rp > 65535 {
		return 0, 0, fmt.Errorf("[ERROR] Invalid port '%s'. Must be an integer between 1 and 65535.", parts[1])
	}
	return lp, rp, nil
}

func portVal(s interface{ Get(string, string) string }, section, key string, def int) int {
	v := s.Get(section, key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 || n > 65535 {
		return def
	}
	return n
}
