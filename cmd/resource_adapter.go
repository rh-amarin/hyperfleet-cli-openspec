package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

const genericAdapterReportMsg = "Status reported via hf rs adapter-report"

func runGenericAdapterReport(cmd *cobra.Command, typeName string, args []string) error {
	if err := errCurlInteractive(genericInteractive); err != nil {
		return err
	}

	adapterName := args[0]
	status := args[1]
	gen, err := parseAdapterStatusGeneration(args[2])
	if err != nil {
		return err
	}
	if err := validateAdapterStatusValue(status); err != nil {
		return err
	}

	s, err := loadConfig()
	if err != nil {
		return err
	}

	explicit := ""
	if len(args) == 4 {
		explicit = args[3]
	}
	if genericInteractive && explicit == "" {
		explicit, err = pickGenericInteractive(cmd, s, typeName)
		if err != nil || explicit == "" {
			return err
		}
	}

	resourceID, err := s.ResourceID(typeName, explicit)
	if err != nil {
		return err
	}
	statusPath, err := s.ResolveResourceStatusPath(typeName, resourceID)
	if err != nil {
		return err
	}

	client := newAPIClient(s)
	p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

	body := adapterStatusCreateBody(adapterName, status, gen)
	result, err := api.Put[resource.AdapterStatus](context.Background(), client, statusPath, body)
	if err != nil {
		return handleAPIError(p, err)
	}

	if result.Adapter == "" {
		p.Info(fmt.Sprintf("Reported adapter status for %s on %s %s (no-op: status unchanged)", adapterName, typeName, resourceID))
		return nil
	}
	p.Info(fmt.Sprintf("Reported adapter status for %s on %s %s", adapterName, typeName, resourceID))
	return p.Print(result)
}

func validateAdapterStatusValue(status string) error {
	if status != "True" && status != "False" && status != "Unknown" {
		return fmt.Errorf("[ERROR] Invalid status value '%s'. Must be one of: True, False, Unknown.", status)
	}
	return nil
}

func parseAdapterStatusGeneration(genStr string) (int32, error) {
	gen, err := strconv.Atoi(genStr)
	if err != nil {
		return 0, fmt.Errorf("[ERROR] Invalid generation '%s': must be an integer", genStr)
	}
	return int32(gen), nil
}

func adapterStatusCreateBody(adapterName, status string, gen int32) resource.AdapterStatusCreateRequest {
	return resource.AdapterStatusCreateRequest{
		Adapter:            adapterName,
		ObservedGeneration: gen,
		ObservedTime:       time.Now().UTC().Format(time.RFC3339),
		Conditions: []resource.ConditionRequest{
			{Type: "Available", Status: status, Reason: "ManualStatusPost", Message: genericAdapterReportMsg},
			{Type: "Applied", Status: status, Reason: "ManualStatusPost", Message: genericAdapterReportMsg},
			{Type: "Health", Status: status, Reason: "ManualStatusPost", Message: genericAdapterReportMsg},
			{Type: "Finalized", Status: status, Reason: "ManualStatusPost", Message: genericAdapterReportMsg},
		},
	}
}
