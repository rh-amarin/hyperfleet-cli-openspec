// Package kube provides Kubernetes operations via client-go without requiring kubectl.
package kube

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/term"
)

// PortForward represents a running port-forward process tracked by a PID file.
type PortForward struct {
	Name       string
	PID        int
	LocalPort  int
	RemotePort int
}

const (
	TargetKindService = "service"
	TargetKindPod     = "pod"
)

// StartResult is returned by StartPortForward with the resolved target details.
type StartResult struct {
	Name       string
	PID        int
	LocalPort  int
	RemotePort int
	Namespace  string
	TargetKind string // TargetKindService or TargetKindPod
	TargetName string
}

// PodNotReadyError is returned when a pod is found but not in Running phase.
type PodNotReadyError struct {
	Name  string
	Phase string
}

func (e *PodNotReadyError) Error() string {
	return fmt.Sprintf("pod %s not ready (phase: %s)", e.Name, e.Phase)
}

// ResolvedContext returns the Kubernetes context name that will be used for API calls.
// If contextName is non-empty it is returned after verifying the context exists in the kubeconfig.
// If contextName is empty the kubeconfig's current-context name is returned.
func ResolvedContext(kubeconfigPath, contextName string) (string, error) {
	resolved := ResolveKubeconfig(kubeconfigPath)
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: resolved}
	rawCfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, &clientcmd.ConfigOverrides{},
	).RawConfig()
	if err != nil {
		return "", err
	}
	if contextName != "" {
		if _, ok := rawCfg.Contexts[contextName]; !ok {
			return "", fmt.Errorf("context %q not found in kubeconfig", contextName)
		}
		return contextName, nil
	}
	if rawCfg.CurrentContext == "" {
		return "", fmt.Errorf("no current context in kubeconfig")
	}
	return rawCfg.CurrentContext, nil
}

// BuildConfig builds a REST config from kubeconfig.
// Resolution order: kubeconfigPath arg → KUBECONFIG env → ~/.kube/config.
// contextName selects a specific context; empty string uses the kubeconfig's current-context.
func BuildConfig(kubeconfigPath, contextName string) (*rest.Config, error) {
	resolved := ResolveKubeconfig(kubeconfigPath)
	if _, err := os.Stat(resolved); os.IsNotExist(err) {
		return nil, fmt.Errorf("[ERROR] kubeconfig not found at %s. Set kubernetes.kubeconfig, KUBECONFIG, or use --kubeconfig.", resolved)
	}
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: resolved}
	overrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		overrides.CurrentContext = contextName
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return nil, err
	}
	if token := os.Getenv("HF_KUBE_TOKEN"); token != "" {
		config.BearerToken = token
		config.BearerTokenFile = ""
	}
	return config, nil
}

// NewClientset creates a Kubernetes clientset from kubeconfig.
// contextName selects a specific context; empty string uses the kubeconfig's current-context.
func NewClientset(kubeconfigPath, contextName string) (kubernetes.Interface, error) {
	config, err := BuildConfig(kubeconfigPath, contextName)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// IsProcessAlive returns true if a process with the given PID is alive.
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// IsPortListening reports whether a process is accepting TCP connections on
// the given local port. It attempts a connection to 127.0.0.1:<port> with a
// short timeout, so it works without any external tools on both macOS and Linux.
func IsPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// PIDForPort returns the PID of the process listening on the given TCP port.
// It runs lsof, which is available on both macOS and Linux, and parses its
// field-mode output (-F p) to extract just the PID line.
func PIDForPort(port int) (int, error) {
	out, err := exec.Command("lsof",
		"-i", fmt.Sprintf("TCP:%d", port),
		"-sTCP:LISTEN",
		"-n", "-P",
		"-F", "p",
	).Output()
	if err != nil {
		return 0, fmt.Errorf("lsof: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "p") {
			pid, err := strconv.Atoi(line[1:])
			if err == nil && pid > 0 {
				return pid, nil
			}
		}
	}
	return 0, fmt.Errorf("no process listening on TCP port %d", port)
}

// serviceHasPort reports whether svc exposes remotePort on spec.ports[].port.
func serviceHasPort(svc *corev1.Service, remotePort int) bool {
	for _, p := range svc.Spec.Ports {
		if int(p.Port) == remotePort {
			return true
		}
	}
	return false
}

// PortForwardResolution holds the pod to forward to and the display target for output.
// When resolved via a Service, DisplayKind is TargetKindService; the tunnel still uses the pod API.
type PortForwardResolution struct {
	PodName     string
	DisplayKind string
	DisplayName string
}

// podForService returns a backing pod name from EndpointSlices for remotePort.
func podForService(ctx context.Context, cs kubernetes.Interface, namespace, serviceName string, remotePort int) (string, error) {
	slices, err := cs.DiscoveryV1().EndpointSlices(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: discoveryv1.LabelServiceName + "=" + serviceName,
	})
	if err != nil {
		return "", err
	}
	for _, slice := range slices.Items {
		portOK := len(slice.Ports) == 0
		for _, p := range slice.Ports {
			if p.Port != nil && int(*p.Port) == remotePort {
				portOK = true
				break
			}
		}
		if !portOK {
			continue
		}
		for _, ep := range slice.Endpoints {
			if ep.TargetRef != nil && ep.TargetRef.Kind == "Pod" && ep.TargetRef.Name != "" {
				return ep.TargetRef.Name, nil
			}
		}
	}
	return "", fmt.Errorf("no ready pod found for service %q port %d in namespace %q", serviceName, remotePort, namespace)
}

// ResolvePortForwardTarget prefers a Kubernetes Service when available, else pod pattern.
// Returns PodNotReadyError as warn when falling back to a pod that is not Running.
func ResolvePortForwardTarget(ctx context.Context, cs kubernetes.Interface, namespace, serviceName, podPattern string, remotePort int) (PortForwardResolution, error) {
	if serviceName != "" {
		svc, err := cs.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err == nil && serviceHasPort(svc, remotePort) {
			if podName, err := podForService(ctx, cs, namespace, serviceName, remotePort); err == nil {
				return PortForwardResolution{
					PodName: podName, DisplayKind: TargetKindService, DisplayName: serviceName,
				}, nil
			}
		}
	}
	podName, err := FindRunningPod(ctx, cs, namespace, podPattern)
	if err != nil {
		var notReady *PodNotReadyError
		if errors.As(err, &notReady) {
			return PortForwardResolution{
				PodName: podName, DisplayKind: TargetKindPod, DisplayName: podName,
			}, err
		}
		return PortForwardResolution{}, err
	}
	return PortForwardResolution{
		PodName: podName, DisplayKind: TargetKindPod, DisplayName: podName,
	}, nil
}

// portForwardURL builds the REST URL for a pod port-forward subresource.
func portForwardURL(cs kubernetes.Interface, namespace, podName string) string {
	return cs.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward").
		URL().String()
}

// FindRunningPod returns the name of the first pod whose name contains pattern.
// Returns PodNotReadyError (non-nil, with pod name) when found but not Running.
func FindRunningPod(ctx context.Context, cs kubernetes.Interface, namespace, pattern string) (string, error) {
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, pattern) {
			if pod.Status.Phase == corev1.PodRunning {
				return pod.Name, nil
			}
			return pod.Name, &PodNotReadyError{Name: pod.Name, Phase: string(pod.Status.Phase)}
		}
	}
	return "", fmt.Errorf("no pod found matching %q in namespace %q", pattern, namespace)
}

// StartPortForward resolves a port-forward target (service-first), spawns a detached daemon
// subprocess, writes a PID file, and returns immediately.
// contextName selects a specific Kubernetes context; empty string uses the kubeconfig's current-context.
func StartPortForward(kubeconfigPath, namespace, name, serviceName, podPattern string, localPort, remotePort int, contextName string) (StartResult, error) {
	cs, err := NewClientset(kubeconfigPath, contextName)
	if err != nil {
		return StartResult{}, err
	}
	res, warnErr := ResolvePortForwardTarget(context.Background(), cs, namespace, serviceName, podPattern, remotePort)
	if warnErr != nil {
		var notReady *PodNotReadyError
		if errors.As(warnErr, &notReady) {
			fmt.Fprintf(os.Stderr, "[WARN] %s: pod not ready (phase: %s). Port-forward may not succeed.\n",
				name, notReady.Phase)
		} else {
			return StartResult{}, warnErr
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return StartResult{}, fmt.Errorf("cannot find executable: %w", err)
	}
	resolved := ResolveKubeconfig(kubeconfigPath)
	cmd := exec.Command(exe, "kube", "port-forward", "_daemon",
		resolved, namespace, res.PodName,
		strconv.Itoa(localPort), strconv.Itoa(remotePort), contextName)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return StartResult{}, fmt.Errorf("starting port-forward daemon: %w", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Release()

	if err := writePIDFile(name, pid, localPort, remotePort); err != nil {
		return StartResult{}, fmt.Errorf("writing PID file: %w", err)
	}
	return StartResult{
		Name: name, PID: pid, LocalPort: localPort, RemotePort: remotePort,
		Namespace: namespace, TargetKind: res.DisplayKind, TargetName: res.DisplayName,
	}, nil
}

// StopPortForward terminates the named port-forward and removes its tracking file.
// It prefers to kill the process actually listening on the local port (found via
// PIDForPort) so that externally-restarted tunnels are handled correctly. If the
// port is not currently bound it falls back to the PID recorded at start time.
func StopPortForward(name string) error {
	storedPID, localPort, _, err := readPIDFile(name)
	if err != nil {
		return fmt.Errorf("no port-forward found for %q", name)
	}

	pid := storedPID
	if activePID, pidErr := PIDForPort(localPort); pidErr == nil {
		pid = activePID
	}

	if pid > 0 {
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Signal(syscall.SIGTERM)
		}
	}
	return os.Remove(pidFilePath(name))
}

// ListPortForwards returns all port-forwards tracked by PID files.
func ListPortForwards() ([]PortForward, error) {
	entries, err := os.ReadDir(pidDir())
	if err != nil {
		return nil, nil
	}
	var result []PortForward
	for _, e := range entries {
		n := e.Name()
		if !strings.HasPrefix(n, "pf-") || !strings.HasSuffix(n, ".pid") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(n, "pf-"), ".pid")
		pid, localPort, remotePort, err := readPIDFile(name)
		if err != nil {
			continue
		}
		result = append(result, PortForward{Name: name, PID: pid, LocalPort: localPort, RemotePort: remotePort})
	}
	return result, nil
}

// RunPortForwardDaemon runs a blocking client-go port-forward for podName.
// Called by the hidden _daemon subcommand in a detached subprocess.
// contextName selects a specific Kubernetes context; empty string uses the kubeconfig's current-context.
func RunPortForwardDaemon(kubeconfigPath, namespace, podName string, localPort, remotePort int, contextName string) error {
	config, err := BuildConfig(kubeconfigPath, contextName)
	if err != nil {
		return err
	}
	cs, err := NewClientset(kubeconfigPath, contextName)
	if err != nil {
		return err
	}
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return fmt.Errorf("SPDY transport: %w", err)
	}
	pfURL, err := url.Parse(portForwardURL(cs, namespace, podName))
	if err != nil {
		return fmt.Errorf("port-forward URL: %w", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, pfURL)
	stopCh := make(chan struct{})
	readyCh := make(chan struct{})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, remotePort)},
		stopCh, readyCh, nil, nil)
	if err != nil {
		return fmt.Errorf("creating port-forward: %w", err)
	}
	return fw.ForwardPorts()
}

// StreamLogs fans out log streaming across all pods matching podPattern.
// Each line is prefixed with [pod-name].
func StreamLogs(ctx context.Context, cs kubernetes.Interface, namespace, podPattern string, w io.Writer) error {
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	var matching []string
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, podPattern) {
			matching = append(matching, pod.Name)
		}
	}
	if len(matching) == 0 {
		return fmt.Errorf("no pods found matching %q in namespace %q", podPattern, namespace)
	}
	var wg sync.WaitGroup
	for _, podName := range matching {
		wg.Add(1)
		go func(pn string) {
			defer wg.Done()
			_ = streamPodLogs(ctx, cs, namespace, pn, w)
		}(podName)
	}
	wg.Wait()
	return nil
}

// StreamLogsFiltered fans out log streaming, filtering lines to those containing
// cluster_id=clusterID in logfmt format. JSON lines (starting with '{') are skipped.
func StreamLogsFiltered(ctx context.Context, cs kubernetes.Interface, namespace, podPattern, clusterID string, w io.Writer) error {
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	var matching []string
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, podPattern) {
			matching = append(matching, pod.Name)
		}
	}
	if len(matching) == 0 {
		return fmt.Errorf("no pods found matching %q in namespace %q", podPattern, namespace)
	}
	var wg sync.WaitGroup
	for _, podName := range matching {
		wg.Add(1)
		go func(pn string) {
			defer wg.Done()
			_ = streamPodLogsFiltered(ctx, cs, namespace, pn, clusterID, w)
		}(podName)
	}
	wg.Wait()
	return nil
}

// CollectLogs fetches log lines from all pods whose name contains podPattern,
// going back sinceSeconds seconds, and returns all lines as a flat slice.
// Per-pod errors are silently skipped; only the pod list call can return an error.
func CollectLogs(ctx context.Context, cs kubernetes.Interface, namespace, podPattern string, sinceSeconds int64) ([]string, error) {
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var lines []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, pod := range pods.Items {
		if !strings.Contains(pod.Name, podPattern) {
			continue
		}
		wg.Add(1)
		go func(podName string) {
			defer wg.Done()
			req := cs.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
				SinceSeconds: &sinceSeconds,
			})
			rc, err := req.Stream(ctx)
			if err != nil {
				return
			}
			defer rc.Close()
			scanner := bufio.NewScanner(rc)
			var podLines []string
			for scanner.Scan() {
				podLines = append(podLines, scanner.Text())
			}
			mu.Lock()
			lines = append(lines, podLines...)
			mu.Unlock()
		}(pod.Name)
	}
	wg.Wait()
	return lines, nil
}

// RunCurlPod creates or reuses an hf-curl pod and runs curl inside it via SPDY exec.
// contextName selects a specific Kubernetes context; empty string uses the kubeconfig's current-context.
func RunCurlPod(ctx context.Context, kubeconfigPath, namespace, contextName string, curlArgs []string, w io.Writer) error {
	cs, err := NewClientset(kubeconfigPath, contextName)
	if err != nil {
		return err
	}
	config, err := BuildConfig(kubeconfigPath, contextName)
	if err != nil {
		return err
	}
	if err := ensureCurlPod(ctx, cs, namespace); err != nil {
		return err
	}
	return execInPod(ctx, cs, config, namespace, "hf-curl", append([]string{"curl"}, curlArgs...), nil, w, w)
}

// CreateDebugPod finds a deployment matching pattern, creates a debug pod from its spec
// (liveness/readiness probes removed, restartPolicy: Never), waits up to 3 minutes for Running,
// and returns the pod name.
func CreateDebugPod(ctx context.Context, cs kubernetes.Interface, namespace, pattern string) (string, error) {
	deps, err := cs.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("listing deployments: %w", err)
	}
	var dep *appsv1.Deployment
	for i := range deps.Items {
		if strings.Contains(deps.Items[i].Name, pattern) {
			dep = &deps.Items[i]
			break
		}
	}
	if dep == nil {
		return "", fmt.Errorf("no deployment found matching %q in namespace %q", pattern, namespace)
	}

	podSpec := dep.Spec.Template.Spec.DeepCopy()
	podSpec.RestartPolicy = corev1.RestartPolicyNever
	for i := range podSpec.Containers {
		podSpec.Containers[i].LivenessProbe = nil
		podSpec.Containers[i].ReadinessProbe = nil
	}
	podName := fmt.Sprintf("hf-debug-%s-%d", dep.Name, time.Now().Unix())
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    map[string]string{"app": "hf-debug"},
		},
		Spec: *podSpec,
	}
	if _, err := cs.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return "", fmt.Errorf("creating debug pod: %w", err)
	}

	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		p, err := cs.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return podName, err
		}
		if p.Status.Phase == corev1.PodRunning {
			return podName, nil
		}
		time.Sleep(time.Second)
	}
	return podName, fmt.Errorf("timed out waiting for debug pod %s to reach Running phase", podName)
}

// ExecShell execs an interactive /bin/sh in the given pod with TTY and terminal resize support.
func ExecShell(ctx context.Context, cs kubernetes.Interface, config *rest.Config, namespace, podName string) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("setting raw terminal: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState) //nolint:errcheck

	tsq := newTermSizeQueue()
	defer tsq.close()

	u := cs.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		URL()
	q := u.Query()
	for _, c := range []string{"/bin/sh"} {
		q.Add("command", c)
	}
	q.Set("stdin", "true")
	q.Set("stdout", "true")
	q.Set("stderr", "true")
	q.Set("tty", "true")
	u.RawQuery = q.Encode()

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", u)
	if err != nil {
		return fmt.Errorf("creating SPDY executor: %w", err)
	}
	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
		Tty:               true,
		TerminalSizeQueue: tsq,
	})
}

// ListHyperfleetNamespaces returns the names of all namespaces that have at least one
// label whose key or value contains the substring "hyperfleet". Results are sorted.
func ListHyperfleetNamespaces(ctx context.Context, cs kubernetes.Interface) ([]string, error) {
	nsList, err := cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, ns := range nsList.Items {
		for k, v := range ns.Labels {
			if strings.Contains(k, "hyperfleet") || strings.Contains(v, "hyperfleet") {
				names = append(names, ns.Name)
				break
			}
		}
	}
	sort.Strings(names)
	return names, nil
}

// DeleteNamespace deletes a Kubernetes namespace by name.
func DeleteNamespace(ctx context.Context, cs kubernetes.Interface, name string) error {
	return cs.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}


// findFreePort binds a listener on 127.0.0.1:0 to obtain a free local TCP port,
// closes it immediately, and returns the port number. There is a brief TOCTOU
// window but it is acceptable for ephemeral port-forward use.
func findFreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}

// EphemeralPortForward starts an in-process SPDY port-forward to the resolved target
// (service-first) in namespace for remotePort, allocating a random local port.
// It returns the local port, a stop func (safe to call multiple times), and any setup error.
// The stop func must be called when the caller is done to release the connection.
func EphemeralPortForward(kubeconfigPath, namespace, serviceName, podPattern string, remotePort int, contextName string) (localPort int, stop func(), err error) {
	config, err := BuildConfig(kubeconfigPath, contextName)
	if err != nil {
		return 0, nil, err
	}
	cs, err := NewClientset(kubeconfigPath, contextName)
	if err != nil {
		return 0, nil, err
	}
	res, resolveErr := ResolvePortForwardTarget(context.Background(), cs, namespace, serviceName, podPattern, remotePort)
	if resolveErr != nil {
		return 0, nil, resolveErr
	}
	localPort, err = findFreePort()
	if err != nil {
		return 0, nil, fmt.Errorf("allocating local port: %w", err)
	}
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return 0, nil, fmt.Errorf("SPDY transport: %w", err)
	}
	pfURL, err := url.Parse(portForwardURL(cs, namespace, res.PodName))
	if err != nil {
		return 0, nil, fmt.Errorf("port-forward URL: %w", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, pfURL)
	stopCh := make(chan struct{})
	readyCh := make(chan struct{})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, remotePort)},
		stopCh, readyCh, nil, nil)
	if err != nil {
		close(stopCh)
		return 0, nil, fmt.Errorf("creating port-forward: %w", err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- fw.ForwardPorts()
	}()
	select {
	case <-readyCh:
		// port-forward is ready
	case pfErr := <-errCh:
		return 0, nil, fmt.Errorf("port-forward failed: %w", pfErr)
	case <-time.After(30 * time.Second):
		close(stopCh)
		return 0, nil, fmt.Errorf("timed out waiting for port-forward to %s:%d", res.PodName, remotePort)
	}
	var once sync.Once
	stop = func() {
		once.Do(func() { close(stopCh) })
	}
	return localPort, stop, nil
}

// CheckAPIConnectivity verifies connectivity to the HyperFleet API via HTTP GET.
// Returns nil if the server responds (any HTTP status), non-nil on connection failure.
func CheckAPIConnectivity(port int) error {
	return checkHTTP(fmt.Sprintf("http://localhost:%d/api/hyperfleet/v1", port))
}

// CheckMaestroHTTPConnectivity verifies connectivity to Maestro's HTTP API via HTTP GET.
// Returns nil if the server responds (any HTTP status), non-nil on connection failure.
func CheckMaestroHTTPConnectivity(port int) error {
	return checkHTTP(fmt.Sprintf("http://localhost:%d/api/maestro/v1", port))
}

// CheckRabbitMQConnectivity verifies connectivity to the RabbitMQ management API via HTTP GET.
// Returns nil if the server responds (any HTTP status), non-nil on connection failure.
func CheckRabbitMQConnectivity(port int) error {
	return checkHTTP(fmt.Sprintf("http://localhost:%d/api/overview", port))
}

// CheckPostgresConnectivity opens a pgx connection to localhost:<port> and pings it.
// Returns nil on successful ping, non-nil on any error.
func CheckPostgresConnectivity(port int, host, dbname, user, password string) error {
	dsn := fmt.Sprintf("host=localhost port=%d dbname=%s user=%s password=%s connect_timeout=2 sslmode=disable",
		port, dbname, user, password)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()
	return pool.Ping(ctx)
}

// CheckMaestroGRPCConnectivity dials localhost:<port> and calls the standard gRPC Health/Check.
// Returns nil on SERVING or UNIMPLEMENTED (server reachable but health proto not registered),
// non-nil on connection failure or other error.
func CheckMaestroGRPCConnectivity(port int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
	hc := grpc_health.NewHealthClient(conn)
	_, err = hc.Check(ctx, &grpc_health.HealthCheckRequest{Service: ""})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.Unimplemented {
			return nil
		}
		return err
	}
	return nil
}

// checkHTTP performs a GET to url with a 2-second timeout.
// Returns nil if any HTTP response is received, non-nil on connection failure.
func checkHTTP(url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ---- private helpers ----

// ResolveKubeconfig returns the kubeconfig path to use.
// When path is non-empty it is returned as-is; otherwise KUBECONFIG env or ~/.kube/config is used.
func ResolveKubeconfig(path string) string {
	if path != "" {
		return path
	}
	if env := os.Getenv("KUBECONFIG"); env != "" {
		return env
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kube", "config")
}

func pidDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "hf")
}

func pidFilePath(name string) string {
	return filepath.Join(pidDir(), "pf-"+name+".pid")
}

func writePIDFile(name string, pid, localPort, remotePort int) error {
	_ = os.MkdirAll(pidDir(), 0700)
	content := fmt.Sprintf("%d\n%d\n%d\n", pid, localPort, remotePort)
	return os.WriteFile(pidFilePath(name), []byte(content), 0600)
}

func readPIDFile(name string) (pid, localPort, remotePort int, err error) {
	data, err := os.ReadFile(pidFilePath(name))
	if err != nil {
		return
	}
	parts := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(parts) < 3 {
		err = fmt.Errorf("malformed PID file for %q", name)
		return
	}
	if pid, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	if localPort, err = strconv.Atoi(parts[1]); err != nil {
		return
	}
	remotePort, err = strconv.Atoi(parts[2])
	return
}

func streamPodLogs(ctx context.Context, cs kubernetes.Interface, namespace, podName string, w io.Writer) error {
	req := cs.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Follow: true})
	rc, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		fmt.Fprintf(w, "[%s] %s\n", podName, scanner.Text())
	}
	return scanner.Err()
}

func streamPodLogsFiltered(ctx context.Context, cs kubernetes.Interface, namespace, podName, clusterID string, w io.Writer) error {
	req := cs.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Follow: true})
	rc, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "{") {
			continue
		}
		fields := ParseLogfmt(line)
		if clusterID != "" && fields["cluster_id"] != clusterID {
			continue
		}
		fmt.Fprintf(w, "[%s] %s  %s  %s\n", podName, fields["time"], strings.ToUpper(fields["level"]), fields["msg"])
	}
	return scanner.Err()
}

// ParseLogfmt parses a logfmt-encoded log line into a key→value map.
func ParseLogfmt(line string) map[string]string {
	result := map[string]string{}
	s := line
	for len(s) > 0 {
		s = strings.TrimLeft(s, " ")
		eq := strings.IndexByte(s, '=')
		if eq < 0 {
			break
		}
		key := s[:eq]
		s = s[eq+1:]
		var value string
		if strings.HasPrefix(s, `"`) {
			end := strings.Index(s[1:], `"`)
			if end < 0 {
				value = s[1:]
				s = ""
			} else {
				value = s[1 : end+1]
				s = s[end+2:]
			}
		} else {
			sp := strings.IndexByte(s, ' ')
			if sp < 0 {
				value = s
				s = ""
			} else {
				value = s[:sp]
				s = s[sp+1:]
			}
		}
		result[key] = value
	}
	return result
}

func ensureCurlPod(ctx context.Context, cs kubernetes.Interface, namespace string) error {
	pod, err := cs.CoreV1().Pods(namespace).Get(ctx, "hf-curl", metav1.GetOptions{})
	if err == nil && pod.Status.Phase == corev1.PodRunning {
		return nil
	}
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "hf-curl", Namespace: namespace},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			Containers: []corev1.Container{{
				Name:    "curl",
				Image:   "curlimages/curl:latest",
				Command: []string{"sh", "-c", "while true; do sleep 3600; done"},
			}},
		},
	}
	if _, err := cs.CoreV1().Pods(namespace).Create(ctx, newPod, metav1.CreateOptions{}); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("creating hf-curl pod: %w", err)
		}
	}
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		p, err := cs.CoreV1().Pods(namespace).Get(ctx, "hf-curl", metav1.GetOptions{})
		if err == nil && p.Status.Phase == corev1.PodRunning {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timed out waiting for hf-curl pod to become ready")
}

func execInPod(ctx context.Context, cs kubernetes.Interface, config *rest.Config, namespace, podName string, command []string, stdin io.Reader, stdout, stderr io.Writer) error {
	u := cs.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		URL()
	q := u.Query()
	for _, c := range command {
		q.Add("command", c)
	}
	if stdin != nil {
		q.Set("stdin", "true")
	}
	q.Set("stdout", "true")
	q.Set("stderr", "true")
	u.RawQuery = q.Encode()

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", u)
	if err != nil {
		return fmt.Errorf("creating SPDY executor: %w", err)
	}
	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
}

type termSizeQueue struct {
	ch   chan remotecommand.TerminalSize
	done chan struct{}
}

func newTermSizeQueue() *termSizeQueue {
	q := &termSizeQueue{
		ch:   make(chan remotecommand.TerminalSize, 1),
		done: make(chan struct{}),
	}
	w, h, _ := term.GetSize(int(os.Stdin.Fd()))
	q.ch <- remotecommand.TerminalSize{Width: uint16(w), Height: uint16(h)}
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		defer signal.Stop(sigCh)
		for {
			select {
			case <-q.done:
				return
			case <-sigCh:
				w, h, _ := term.GetSize(int(os.Stdin.Fd()))
				select {
				case q.ch <- remotecommand.TerminalSize{Width: uint16(w), Height: uint16(h)}:
				default:
				}
			}
		}
	}()
	return q
}

func (q *termSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.ch
	if !ok {
		return nil
	}
	return &size
}

func (q *termSizeQueue) close() {
	close(q.done)
}
