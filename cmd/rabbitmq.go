// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
	"github.com/spf13/cobra"
)

// rabbitFactory builds the RabbitPublisher used by rabbitmq commands.
// Replaced in tests with a mock factory.
var rabbitFactory = func() pubsub.RabbitPublisher {
	return pubsub.NewRabbitClient()
}

// rabbitmqCmd is the top-level group for RabbitMQ operations.
var rabbitmqCmd = &cobra.Command{
	Use:   "rabbitmq",
	Short: "Publish events to RabbitMQ exchanges",
	Long: `Publish events to RabbitMQ exchanges.

Subcommands: publish.`,
}

// rabbitmqPublishCmd publishes a reconcile CloudEvent for any configured resource type.
var rabbitmqPublishCmd = &cobra.Command{
	Use:   "publish <resource-type> <exchange> [routing-key]",
	Short: "Publish a reconcile CloudEvent to a RabbitMQ exchange",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		typeName := args[0]
		exchange := args[1]
		routingKey := ""
		if len(args) > 2 {
			routingKey = args[2]
		}

		s, err := loadConfig()
		if err != nil {
			return err
		}

		def, err := s.ResourceType(typeName)
		if err != nil {
			return err
		}

		// Check ancestor state before own state so missing-parent errors surface first.
		ancestors, err := buildRabbitAncestors(s, typeName)
		if err != nil {
			return err
		}

		resourceID := s.GetState(def.StateKey)
		if resourceID == "" {
			return fmt.Errorf("[ERROR] No %s set in state. Run 'hf resource %s search <name>' first.", def.StateKey, typeName)
		}

		resourcePath, err := s.ResolveResourcePath(typeName)
		if err != nil {
			return err
		}

		apiURL := s.Get("hyperfleet", "api-url")
		apiVersion := s.Get("hyperfleet", "api-version")
		data, err := pubsub.BuildGenericReconcileEvent(typeName, resourceID, ancestors, resourcePath, apiURL, apiVersion)
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to build event: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))

		baseURL := rabbitBaseURL(s)
		rabbit := rabbitFactory()
		if err := rabbit.Publish(
			context.Background(),
			baseURL,
			s.Get("rabbitmq", "user"),
			s.Get("rabbitmq", "password"),
			s.Get("rabbitmq", "vhost"),
			exchange,
			routingKey,
			data,
		); err != nil {
			return fmt.Errorf("[ERROR] Failed to publish: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Published %s %s to exchange %s\n", typeName, resourceID, exchange)
		return nil
	},
}

// buildRabbitAncestors walks the parent chain for typeName and returns AncestorID entries
// ordered root→immediate-parent, each with their state-resolved IDs and paths.
func buildRabbitAncestors(s *config.Store, typeName string) ([]pubsub.AncestorID, error) {
	def, err := s.ResourceType(typeName)
	if err != nil {
		return nil, err
	}
	if def.Parent == "" {
		return nil, nil
	}

	var ancestors []pubsub.AncestorID
	current := def.Parent
	for current != "" {
		parentDef, err := s.ResourceType(current)
		if err != nil {
			return nil, err
		}
		id := s.GetState(parentDef.StateKey)
		if id == "" {
			return nil, fmt.Errorf("[ERROR] No %s set in state. Run 'hf resource %s search <name>' first.", parentDef.StateKey, current)
		}
		path, err := s.ResolveResourcePath(current)
		if err != nil {
			return nil, err
		}
		ancestors = append([]pubsub.AncestorID{{TypeName: current, ID: id, Path: path}}, ancestors...)
		current = parentDef.Parent
	}
	return ancestors, nil
}

// rabbitBaseURL constructs the HTTP management API base URL.
func rabbitBaseURL(s interface{ Get(string, string) string }) string {
	host := s.Get("rabbitmq", "host")
	port := s.Get("rabbitmq", "mgmt-port")
	return strings.TrimRight("http://"+host+":"+port, "/")
}

func init() {
	rootCmd.AddCommand(rabbitmqCmd)
	rabbitmqCmd.AddCommand(rabbitmqPublishCmd)
}
