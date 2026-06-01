// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
	"github.com/spf13/cobra"
)

// gcpFactory builds a GCPPublisher from Application Default Credentials.
// Replaced in tests with a mock factory.
var gcpFactory = func(ctx context.Context) (pubsub.GCPPublisher, error) {
	return pubsub.NewGCPClient(ctx)
}

// pubsubCmd is the top-level group for GCP Pub/Sub operations.
var pubsubCmd = &cobra.Command{
	Use:   "pubsub",
	Short: "Interact with GCP Pub/Sub topics",
	Long: `Interact with GCP Pub/Sub topics.

Subcommands: list, publish.`,
}

// pubsubListCmd lists all topics (and subscriptions) in the configured GCP project.
var pubsubListCmd = &cobra.Command{
	Use:   "list [filter]",
	Short: "List GCP Pub/Sub topics and subscriptions",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		ctx := context.Background()
		gcp, err := gcpFactory(ctx)
		if err != nil {
			return err
		}

		projectID := s.Get("hyperfleet", "gcp-project")
		fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Listing topics in project: %s\n", projectID)

		filter := ""
		if len(args) > 0 {
			filter = args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Filtering by: %s\n", filter)
		}

		groups, err := gcp.ListTopics(ctx, projectID)
		if err != nil {
			return fmt.Errorf("[ERROR] %w", err)
		}

		found := false
		for _, g := range groups {
			if filter != "" && !strings.Contains(g.Topic, filter) {
				hasSub := false
				for _, sub := range g.Subscriptions {
					if strings.Contains(sub, filter) {
						hasSub = true
						break
					}
				}
				if !hasSub {
					continue
				}
			}
			found = true
			fmt.Fprintln(cmd.OutOrStdout(), g.Topic)
			for _, sub := range g.Subscriptions {
				if filter == "" || strings.Contains(sub, filter) {
					fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", sub)
				}
			}
		}
		if !found {
			fmt.Fprintln(cmd.OutOrStdout(), "No topics found.")
		}
		return nil
	},
}

// pubsubPublishCmd is the parent for pubsub publish subcommands.
var pubsubPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a CloudEvent to a GCP Pub/Sub topic",
}

// pubsubPublishClusterCmd publishes a cluster reconcile event.
var pubsubPublishClusterCmd = &cobra.Command{
	Use:   "cluster <topic>",
	Short: "Publish a cluster reconcile event to a Pub/Sub topic",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		topic := args[0]
		s, err := loadConfig()
		if err != nil {
			return err
		}

		clusterID := s.GetState("clusters")
		if clusterID == "" {
			return fmt.Errorf("[ERROR] No clusters set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
		}

		apiURL := s.Get("hyperfleet", "api-url")
		apiVersion := s.Get("hyperfleet", "api-version")
		data, err := pubsub.BuildClusterEvent(clusterID, apiURL, apiVersion)
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to build event: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))

		ctx := context.Background()
		gcp, err := gcpFactory(ctx)
		if err != nil {
			return err
		}

		projectID := s.Get("hyperfleet", "gcp-project")
		msgID, err := gcp.Publish(ctx, projectID, topic, data)
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to publish: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Published cluster %s to topic %s (msg-id: %s)\n", clusterID, topic, msgID)
		return nil
	},
}

// pubsubPublishNodePoolCmd publishes a nodepool reconcile event.
var pubsubPublishNodePoolCmd = &cobra.Command{
	Use:   "nodepool <topic>",
	Short: "Publish a nodepool reconcile event to a Pub/Sub topic",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		topic := args[0]
		s, err := loadConfig()
		if err != nil {
			return err
		}

		clusterID := s.GetState("clusters")
		if clusterID == "" {
			return fmt.Errorf("[ERROR] No clusters set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.")
		}
		nodepoolID := s.GetState("nodepools")
		if nodepoolID == "" {
			return fmt.Errorf("[ERROR] No nodepools set in state. Run 'hf nodepool create' or 'hf nodepool use <id>' first.")
		}

		apiURL := s.Get("hyperfleet", "api-url")
		apiVersion := s.Get("hyperfleet", "api-version")
		data, err := pubsub.BuildNodePoolEvent(clusterID, nodepoolID, apiURL, apiVersion)
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to build event: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))

		ctx := context.Background()
		gcp, err := gcpFactory(ctx)
		if err != nil {
			return err
		}

		projectID := s.Get("hyperfleet", "gcp-project")
		msgID, err := gcp.Publish(ctx, projectID, topic, data)
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to publish: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Published nodepool %s to topic %s (msg-id: %s)\n", nodepoolID, topic, msgID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pubsubCmd)
	pubsubCmd.AddCommand(pubsubListCmd)
	pubsubCmd.AddCommand(pubsubPublishCmd)
	pubsubPublishCmd.AddCommand(pubsubPublishClusterCmd)
	pubsubPublishCmd.AddCommand(pubsubPublishNodePoolCmd)
}
