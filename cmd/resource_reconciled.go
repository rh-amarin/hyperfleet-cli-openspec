package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

var (
	genericReconciledDeleteForce   bool
	genericReconciledDeleteReason  string
	genericReconciledForceReason   string
	genericCreateReplicas          int
	genericCreateNodepoolID        string
	genericCreatePlatformType      string
	genericStatusesFilter          bool
)

func runReconciledList(cmd *cobra.Command, typeName string) error {
	profile := reconciledProfile(typeName)
	watch, watchSecs := effectiveListWatch()
	if watch && outputFmt == "table" {
		if curlMode {
			return runReconciledListOnce(cmd, profile, 0, 0)
		}
		ctx, cancel := watchContext(context.Background())
		defer cancel()
		return runWatch(ctx, cmd.OutOrStdout(), watchSecs, func(tick int) error {
			return runReconciledListOnce(cmd, profile, tick, watchSecs)
		})
	}
	return runReconciledListOnce(cmd, profile, 0, 0)
}

func runReconciledListOnce(cmd *cobra.Command, profile string, tick, frequencySecs int) error {
	switch profile {
	case "cluster":
		return fetchAndRenderReconciledClusterList(cmd, tick, frequencySecs)
	case "nodepool":
		return fetchAndRenderReconciledNodepoolList(cmd, tick, frequencySecs)
	default:
		return fmt.Errorf("unknown reconciled profile %q", profile)
	}
}

func runReconciledTable(cmd *cobra.Command, typeName string) error {
	prev := outputFmt
	outputFmt = "table"
	defer func() { outputFmt = prev }()
	return runReconciledList(cmd, typeName)
}

func syncClusterFlagsFromGeneric() {
	clusterInteractive = genericInteractive
	clusterListSearch = genericListSearch
	clusterListWatch = genericListWatch
	clusterListWatchSecs = genericListWatchSecs
	clusterCreateName = genericCreateName
	clusterCreateFile = genericCreateFile
	clusterCreateReplicas = genericCreateReplicas
	clusterCreateNPID = genericCreateNodepoolID
	clusterDeleteForce = genericReconciledDeleteForce
	clusterDeleteReason = genericReconciledDeleteReason
	clusterStatusesFilter = genericStatusesFilter
	clusterIDInteractive = genericIDInteractive
}

func syncNodepoolFlagsFromGeneric() {
	nodepoolInteractive = genericInteractive
	nodepoolListSearch = genericListSearch
	nodepoolListWatch = genericListWatch
	nodepoolListWatchSecs = genericListWatchSecs
	nodepoolCreateName = genericCreateName
	nodepoolCreateFile = genericCreateFile
	nodepoolCreateType = genericCreatePlatformType
	nodepoolCreateReplicas = genericCreateReplicas
	nodepoolDeleteForce = genericReconciledDeleteForce
	nodepoolForceDeleteReason = genericReconciledForceReason
	nodepoolStatusesFilter = genericStatusesFilter
	nodepoolIDInteractive = genericIDInteractive
}

func runReconciledClusterGet(cmd *cobra.Command, args []string) error {
	syncClusterFlagsFromGeneric()
	return clusterGetCmd.RunE(cmd, args)
}

func runReconciledClusterSearch(cmd *cobra.Command, args []string) error {
	syncClusterFlagsFromGeneric()
	return clusterSearchCmd.RunE(cmd, args)
}

func runReconciledClusterCreate(cmd *cobra.Command, args []string) error {
	return runReconciledClusterCreateFromStore(cmd, args)
}

func runReconciledClusterPatch(cmd *cobra.Command, section, explicit string) error {
	if genericPatchFile != "" {
		return runGenericPatch(cmd, "clusters", section, explicit)
	}
	syncClusterFlagsFromGeneric()
	args := []string{section}
	if explicit != "" {
		args = append(args, explicit)
	}
	return clusterPatchCmd.RunE(cmd, args)
}

func runReconciledClusterDelete(cmd *cobra.Command, explicit string) error {
	syncClusterFlagsFromGeneric()
	args := []string{}
	if explicit != "" {
		args = append(args, explicit)
	}
	return clusterDeleteCmd.RunE(cmd, args)
}

func runReconciledClusterConditions(cmd *cobra.Command, explicit string) error {
	syncClusterFlagsFromGeneric()
	args := []string{}
	if explicit != "" {
		args = append(args, explicit)
	}
	return clusterConditionsCmd.RunE(cmd, args)
}

func runReconciledClusterStatuses(cmd *cobra.Command, explicit string) error {
	syncClusterFlagsFromGeneric()
	args := []string{}
	if explicit != "" {
		args = append(args, explicit)
	}
	return clusterStatusesCmd.RunE(cmd, args)
}

func runReconciledClusterID(cmd *cobra.Command) error {
	syncClusterFlagsFromGeneric()
	return clusterIDCmd.RunE(cmd, nil)
}

func runReconciledNodepoolGet(cmd *cobra.Command, args []string) error {
	syncNodepoolFlagsFromGeneric()
	return nodepoolGetCmd.RunE(cmd, args)
}

func runReconciledNodepoolSearch(cmd *cobra.Command, args []string) error {
	syncNodepoolFlagsFromGeneric()
	return nodepoolSearchCmd.RunE(cmd, args)
}

func runReconciledNodepoolCreate(cmd *cobra.Command, args []string) error {
	return runReconciledNodepoolCreateFromStore(cmd, args)
}

func runReconciledNodepoolPatch(cmd *cobra.Command, section, explicit string) error {
	if genericPatchFile != "" {
		return runGenericPatch(cmd, "nodepools", section, explicit)
	}
	syncNodepoolFlagsFromGeneric()
	args := []string{section}
	if explicit != "" {
		args = append(args, explicit)
	}
	return nodepoolPatchCmd.RunE(cmd, args)
}

func runReconciledNodepoolDelete(cmd *cobra.Command, explicit string) error {
	syncNodepoolFlagsFromGeneric()
	args := []string{}
	if explicit != "" {
		args = append(args, explicit)
	}
	return nodepoolDeleteCmd.RunE(cmd, args)
}

func runReconciledNodepoolForceDelete(cmd *cobra.Command, explicit string) error {
	syncNodepoolFlagsFromGeneric()
	args := []string{}
	if explicit != "" {
		args = append(args, explicit)
	}
	return nodepoolForceDeleteCmd.RunE(cmd, args)
}

func runReconciledNodepoolConditions(cmd *cobra.Command, explicit string) error {
	syncNodepoolFlagsFromGeneric()
	args := []string{}
	if explicit != "" {
		args = append(args, explicit)
	}
	return nodepoolConditionsCmd.RunE(cmd, args)
}

func runReconciledNodepoolStatuses(cmd *cobra.Command, explicit string) error {
	syncNodepoolFlagsFromGeneric()
	args := []string{}
	if explicit != "" {
		args = append(args, explicit)
	}
	return nodepoolStatusesCmd.RunE(cmd, args)
}

func runReconciledNodepoolID(cmd *cobra.Command) error {
	syncNodepoolFlagsFromGeneric()
	return nodepoolIDCmd.RunE(cmd, nil)
}

// runReconciledClusterCreateWithTemplate uses config template path for hf rs clusters create.
func runReconciledClusterCreateFromStore(cmd *cobra.Command, args []string) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	def, err := s.ResourceType("clusters")
	if err != nil {
		return err
	}
	body, err := loadResourceTemplate(s, def.CreateTemplate, genericCreateFile)
	if err != nil {
		return fmt.Errorf("[ERROR] %w", err)
	}
	name := genericCreateName
	if len(args) >= 1 {
		name = args[0]
	}
	if name != "" {
		body["name"] = name
	}
	spec := ensureSpecMap(body)
	if len(args) >= 2 {
		spec["region"] = args[1]
	}
	if len(args) >= 3 {
		spec["version"] = args[2]
	}
	if genericCreateReplicas > 0 {
		spec["replicas"] = strconv.Itoa(genericCreateReplicas)
	}
	if genericCreateNodepoolID != "" {
		body["nodepool_id"] = genericCreateNodepoolID
	}
	nameStr, _ := body["name"].(string)

	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	if !curlMode && nameStr != "" {
		existing, err := api.Get[resource.ListResponse[resource.Cluster]](
			context.Background(), client,
			"clusters?search=name='"+nameStr+"'",
		)
		if err == nil && len(existing.Items) > 0 {
			p.Warn(fmt.Sprintf("Cluster '%s' already exists, skipping creation", nameStr))
			return nil
		}
	}

	cluster, err := api.Post[resource.Cluster](context.Background(), client, "clusters", body)
	if err != nil {
		return handleAPIError(p, err)
	}
	if setErr := s.SetState(def.StateKey, cluster.ID); setErr != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist %s: %v\n", def.StateKey, setErr)
	} else {
		p.Info(fmt.Sprintf("Cluster context set to '%s'", cluster.ID))
	}
	return p.Print(cluster)
}

func runReconciledNodepoolCreateFromStore(cmd *cobra.Command, args []string) error {
	s, err := loadConfig()
	if err != nil {
		return err
	}
	def, err := s.ResourceType("nodepools")
	if err != nil {
		return err
	}
	clusterID, err := s.ResourceID("clusters", "")
	if err != nil {
		return err
	}
	body, err := loadResourceTemplate(s, def.CreateTemplate, genericCreateFile)
	if err != nil {
		return fmt.Errorf("[ERROR] %w", err)
	}
	name := genericCreateName
	if len(args) >= 1 {
		name = args[0]
	}
	if name != "" {
		body["name"] = name
	}
	spec := ensureSpecMap(body)
	if genericCreatePlatformType != "" {
		platform, ok := spec["platform"].(map[string]any)
		if !ok {
			platform = map[string]any{}
			spec["platform"] = platform
		}
		platform["type"] = genericCreatePlatformType
	}
	if genericCreateReplicas > 0 {
		spec["replicas"] = genericCreateReplicas
	}
	nameStr, _ := body["name"].(string)

	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	if !curlMode && nameStr != "" {
		existing, err := api.Get[resource.ListResponse[resource.NodePool]](
			context.Background(), client,
			npBase(clusterID)+"?search=name='"+nameStr+"'",
		)
		if err == nil && len(existing.Items) > 0 {
			p.Warn(fmt.Sprintf("NodePool '%s' already exists, skipping creation", nameStr))
			return nil
		}
	}

	np, err := api.Post[resource.NodePool](context.Background(), client, npBase(clusterID), body)
	if err != nil {
		return handleAPIError(p, err)
	}
	if setErr := s.SetState(def.StateKey, np.ID); setErr != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] Failed to persist %s: %v\n", def.StateKey, setErr)
	} else {
		p.Info(fmt.Sprintf("NodePool context set to '%s'", np.ID))
	}
	return p.Print(np)
}
