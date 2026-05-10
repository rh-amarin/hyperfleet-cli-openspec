// Package pubsub provides GCP Pub/Sub and RabbitMQ CloudEvent publishing.
package pubsub

import "context"

// GCPPublisher abstracts GCP Pub/Sub operations for testability.
type GCPPublisher interface {
	// Publish publishes data to the given topic in projectID and returns the server message ID.
	Publish(ctx context.Context, projectID, topicID string, data []byte) (string, error)
	// ListTopics returns all topics (and their subscriptions) in projectID.
	ListTopics(ctx context.Context, projectID string) ([]TopicGroup, error)
}

// RabbitPublisher abstracts RabbitMQ HTTP management API publishing.
type RabbitPublisher interface {
	// Publish sends payload to exchange via the RabbitMQ HTTP management API.
	Publish(ctx context.Context, baseURL, user, password, vhost, exchange, routingKey string, payload []byte) error
}

// TopicGroup pairs a Pub/Sub topic name with its subscription names.
type TopicGroup struct {
	Topic         string
	Subscriptions []string
}
