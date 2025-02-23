// Package pubsub wraps nats calls
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.infratographer.com/x/gidx"
	"go.infratographer.com/x/pubsubx"
	"go.uber.org/zap"
)

const (
	// CreateEventType is the create event type string
	CreateEventType = "create"
	// DeleteEventType is the delete event type string
	DeleteEventType = "delete"
	// UpdateEventType is the update event type string
	UpdateEventType = "update"
)

// May be a config option later
var prefix = "com.infratographer.events"

func newMessage(actorID, subjectID gidx.PrefixedID, additionalSubjectIDs ...gidx.PrefixedID) *pubsubx.ChangeMessage {
	return &pubsubx.ChangeMessage{
		SubjectID:            subjectID,
		ActorID:              actorID,
		Timestamp:            time.Now().UTC(),
		Source:               "tenantapi",
		AdditionalSubjectIDs: additionalSubjectIDs,
	}
}

// PublishCreate publishes a create event
func (c *Client) PublishCreate(ctx context.Context, actor gidx.PrefixedID, location string, data *pubsubx.ChangeMessage) error {
	data.EventType = CreateEventType

	return c.publish(ctx, CreateEventType, actor, location, data)
}

// PublishUpdate publishes an update event
func (c *Client) PublishUpdate(ctx context.Context, actor gidx.PrefixedID, location string, data *pubsubx.ChangeMessage) error {
	data.EventType = UpdateEventType

	return c.publish(ctx, UpdateEventType, actor, location, data)
}

// PublishDelete publishes a delete event
func (c *Client) PublishDelete(ctx context.Context, actor gidx.PrefixedID, location string, data *pubsubx.ChangeMessage) error {
	data.EventType = DeleteEventType
	return c.publish(ctx, DeleteEventType, actor, location, data)
}

// publish publishes an event
func (c *Client) publish(ctx context.Context, action, actor gidx.PrefixedID, location string, data interface{}) error {
	subject := fmt.Sprintf("%s.%s.%s.%s", prefix, actor, action, location)

	b, err := json.Marshal(data)
	if err != nil {
		c.logger.Debug("failed to marshal message", zap.String("nats.subject", subject), zap.Error(err))

		return err
	}

	if _, err := c.js.Publish(subject, b); err != nil {
		c.logger.Debug("failed to publish nats message", zap.String("nats.subject", subject), zap.Error(err))

		return err
	}

	c.logger.Debug("published nats message", zap.String("nats.subject", subject))

	return nil
}

// ChanSubscribe creates a subcription and returns messages on a channel
func (c *Client) ChanSubscribe(ctx context.Context, sub string, ch chan *nats.Msg, stream string) (*nats.Subscription, error) {
	return c.js.ChanSubscribe(sub, ch, nats.BindStream(stream))
}
