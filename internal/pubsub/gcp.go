package pubsub

import (
	"context"
	"errors"
	"fmt"

	gcppubsub "cloud.google.com/go/pubsub"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
)

// ErrNoCredentials is returned when GCP Application Default Credentials are not found.
var ErrNoCredentials = errors.New("[ERROR] GCP credentials not found. Run 'gcloud auth application-default login' or set GOOGLE_APPLICATION_CREDENTIALS")

// GCPClient implements GCPPublisher using the GCP Pub/Sub SDK.
type GCPClient struct{}

// NewGCPClient checks for Application Default Credentials and returns a GCPClient.
// Returns ErrNoCredentials if no credentials are found.
func NewGCPClient(ctx context.Context) (*GCPClient, error) {
	_, err := google.FindDefaultCredentials(ctx, gcppubsub.ScopePubSub)
	if err != nil {
		return nil, ErrNoCredentials
	}
	return &GCPClient{}, nil
}

// Publish publishes data to topicID in projectID and returns the server-assigned message ID.
func (g *GCPClient) Publish(ctx context.Context, projectID, topicID string, data []byte) (string, error) {
	client, err := gcppubsub.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("create pubsub client: %w", err)
	}
	defer client.Close()

	topic := client.Topic(topicID)
	result := topic.Publish(ctx, &gcppubsub.Message{Data: data})
	msgID, err := result.Get(ctx)
	if err != nil {
		return "", err
	}
	return msgID, nil
}

// ListTopics returns all topics and their subscriptions in projectID.
func (g *GCPClient) ListTopics(ctx context.Context, projectID string) ([]TopicGroup, error) {
	client, err := gcppubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}
	defer client.Close()

	var groups []TopicGroup
	it := client.Topics(ctx)
	for {
		topic, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list topics: %w", err)
		}

		group := TopicGroup{Topic: topic.ID()}
		sit := topic.Subscriptions(ctx)
		for {
			sub, serr := sit.Next()
			if serr == iterator.Done {
				break
			}
			if serr != nil {
				return nil, fmt.Errorf("list subscriptions for %s: %w", topic.ID(), serr)
			}
			group.Subscriptions = append(group.Subscriptions, sub.ID())
		}
		groups = append(groups, group)
	}
	return groups, nil
}
