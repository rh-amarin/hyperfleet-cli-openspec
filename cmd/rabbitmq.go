// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"fmt"
	"strings"

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

// rabbitmqPublishCmd is the parent for rabbitmq publish subcommands.
var rabbitmqPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a CloudEvent to a RabbitMQ exchange",
}

// rabbitmqPublishClusterCmd publishes a cluster reconcile event to RabbitMQ.
var rabbitmqPublishClusterCmd = &cobra.Command{
	Use:   "cluster <exchange> [routing-key]",
	Short: "Publish a cluster reconcile event to a RabbitMQ exchange",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		exchange := args[0]
		routingKey := ""
		if len(args) > 1 {
			routingKey = args[1]
		}

		s, err := loadConfig()
		if err != nil {
			return err
		}

		clusterID := s.GetState("cluster-id")
		if clusterID == "" {
			return fmt.Errorf("[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
		}

		apiURL := s.Get("hyperfleet", "api-url")
		apiVersion := s.Get("hyperfleet", "api-version")
		data, err := pubsub.BuildClusterEvent(clusterID, apiURL, apiVersion)
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
		fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Published cluster %s to exchange %s\n", clusterID, exchange)
		return nil
	},
}

// rabbitmqPublishNodePoolCmd publishes a nodepool reconcile event to RabbitMQ.
var rabbitmqPublishNodePoolCmd = &cobra.Command{
	Use:   "nodepool <exchange> [routing-key]",
	Short: "Publish a nodepool reconcile event to a RabbitMQ exchange",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		exchange := args[0]
		routingKey := ""
		if len(args) > 1 {
			routingKey = args[1]
		}

		s, err := loadConfig()
		if err != nil {
			return err
		}

		clusterID := s.GetState("cluster-id")
		if clusterID == "" {
			return fmt.Errorf("[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
		}
		nodepoolID := s.GetState("nodepool-id")
		if nodepoolID == "" {
			return fmt.Errorf("[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool use <id>' first.")
		}

		apiURL := s.Get("hyperfleet", "api-url")
		apiVersion := s.Get("hyperfleet", "api-version")
		data, err := pubsub.BuildNodePoolEvent(clusterID, nodepoolID, apiURL, apiVersion)
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
		fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Published nodepool %s to exchange %s\n", nodepoolID, exchange)
		return nil
	},
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
	rabbitmqPublishCmd.AddCommand(rabbitmqPublishClusterCmd)
	rabbitmqPublishCmd.AddCommand(rabbitmqPublishNodePoolCmd)
}
