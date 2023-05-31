package pubsub

import (
	"context"

	gcps "cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"
)

type PubSub interface {
	CreateSubscription(ctx context.Context, subscription Subscription) error
	DeleteSubscription(ctx context.Context, subscription Subscription) error
}

type pubSubImpl struct {
	PubSub

	log    *log.Entry
	client *gcps.Client
}

func NewPubSub(ctx context.Context, config *Config) (PubSub, error) {
	client, err := gcps.NewClient(ctx, config.ProjectID)

	if err != nil {
		return nil, err
	}

	return NewPubSubWithClient(client), nil
}

func NewPubSubWithClient(client *gcps.Client) PubSub {
	log := log.WithField("component", "pubsub")

	return &pubSubImpl{
		log:    log,
		client: client,
	}
}

func (ps *pubSubImpl) ensureTopic(ctx context.Context, topicName string) (*gcps.Topic, error) {
	log := ps.log.WithField("topic", topicName)

	topic := ps.client.Topic(topicName)

	exists, err := topic.Exists(ctx)

	if err != nil {
		log.WithError(err).Error("error checking if topic exists")
		return nil, err
	}

	if !exists {
		topic, err = ps.client.CreateTopic(ctx, topicName)
	}

	if err != nil {
		log.WithError(err).Error("error creating topic")
		return nil, err
	}

	return topic, nil
}

func (ps *pubSubImpl) CreateSubscription(ctx context.Context, subscription Subscription) error {
	log := ps.log.WithField("subscription_id", subscription.GetSubscriptionID()).WithField("topic", subscription.Topic).WithField("endpoint", subscription.Endpoint)

	topic, err := ps.ensureTopic(ctx, subscription.Topic)

	if err != nil {
		log.WithError(err).Error("error ensuring topic")
		return err
	}

	sub := ps.client.Subscription(subscription.GetSubscriptionID())

	exists, err := sub.Exists(ctx)

	if err != nil {
		log.WithError(err).Error("error checking if subscription exists")
		return err
	}

	// re-create subscription if it already exists.
	// TODO: evaluate, maybe we should just update the subscription instead?
	// however, updating a subscription does not update all of the properties
	if exists {
		sub.Delete(ctx)
		// return ps.updateSubscription(ctx, sub, subscription)
	}

	return ps.createSubscription(ctx, topic, subscription)
}

func (ps *pubSubImpl) createSubscription(ctx context.Context, topic *gcps.Topic, subscription Subscription) error {
	log := ps.log.WithField("subscription_id", subscription.GetSubscriptionID()).WithField("topic", subscription.Topic).WithField("endpoint", subscription.Endpoint)

	_, err := ps.client.CreateSubscription(ctx, subscription.GetSubscriptionID(), createSubscriptionConfig(topic, subscription))

	if err != nil {
		log.WithError(err).Error("error creating subscription")
		return err
	}

	log.Debug("subscription created")

	return nil
}

// func (ps *pubSubImpl) updateSubscription(ctx context.Context, sub *gcps.Subscription, subscription Subscription) error {
// 	log := ps.log.WithField("subscription_id", subscription.GetSubscriptionID()).WithField("topic", subscription.Topic).WithField("endpoint", subscription.Endpoint)

// 	_, err := sub.Update(ctx, updateSubscriptionConfig(subscription))

// 	if err != nil {
// 		log.WithError(err).Error("error updating subscription")
// 		return err
// 	}

// 	log.Debug("subscription updated")

// 	return nil
// }

func (ps *pubSubImpl) DeleteSubscription(ctx context.Context, subscription Subscription) error {
	log := ps.log.WithField("subscription_id", subscription.GetSubscriptionID()).WithField("topic", subscription.Topic).WithField("endpoint", subscription.Endpoint)

	sub := ps.client.Subscription(subscription.GetSubscriptionID())

	exists, err := sub.Exists(ctx)

	if err != nil {
		log.WithError(err).Error("error checking if subscription exists")
		return err
	}

	if !exists {
		log.Debug("skipping non-existing subscription")
		return nil
	}

	err = sub.Delete(ctx)

	if err != nil {
		log.WithError(err).Error("error removing subscription")
		return err
	}

	log.Debug("subscription removed")

	return nil
}

func createSubscriptionConfig(topic *gcps.Topic, subscription Subscription) gcps.SubscriptionConfig {
	var deadLetterPolicy *gcps.DeadLetterPolicy

	if subscription.DeadLetterTopic != "" {
		deadLetterPolicy = &gcps.DeadLetterPolicy{
			DeadLetterTopic:     subscription.DeadLetterTopic,
			MaxDeliveryAttempts: subscription.MaxDeadLetterDeliveryAttempts,
		}
	}

	retryPolicy := &gcps.RetryPolicy{}
	if subscription.RetryMinimumBackoff != nil {
		retryPolicy.MinimumBackoff = *subscription.RetryMinimumBackoff
	}
	if subscription.RetryMaximumBackoff != nil {
		retryPolicy.MaximumBackoff = *subscription.RetryMaximumBackoff
	}

	return gcps.SubscriptionConfig{
		Topic: topic,
		PushConfig: gcps.PushConfig{
			Endpoint: subscription.Endpoint,
		},
		AckDeadline:               subscription.AckDeadline,
		RetainAckedMessages:       subscription.RetainAckedMessages,
		RetentionDuration:         subscription.RetentionDuration,
		EnableMessageOrdering:     subscription.EnableOrdering,
		ExpirationPolicy:          subscription.ExpirationTTL,
		Filter:                    subscription.Filter,
		EnableExactlyOnceDelivery: subscription.DeliverExactlyOnce,
		DeadLetterPolicy:          deadLetterPolicy,
		RetryPolicy:               retryPolicy,
	}
}

// func updateSubscriptionConfig(subscription Subscription) gcps.SubscriptionConfigToUpdate {
// 	var deadLetterPolicy *gcps.DeadLetterPolicy

// 	if subscription.DeadLetterTopic != "" {
// 		deadLetterPolicy = &gcps.DeadLetterPolicy{
// 			DeadLetterTopic:     subscription.DeadLetterTopic,
// 			MaxDeliveryAttempts: subscription.MaxDeadLetterDeliveryAttempts,
// 		}
// 	}

// 	retryPolicy := &gcps.RetryPolicy{}
// 	if subscription.RetryMinimumBackoff != nil {
// 		retryPolicy.MinimumBackoff = *subscription.RetryMinimumBackoff
// 	}
// 	if subscription.RetryMaximumBackoff != nil {
// 		retryPolicy.MaximumBackoff = *subscription.RetryMaximumBackoff
// 	}

// 	return gcps.SubscriptionConfigToUpdate{
// 		PushConfig: &gcps.PushConfig{
// 			Endpoint: subscription.Endpoint,
// 		},
// 		AckDeadline:               subscription.AckDeadline,
// 		RetainAckedMessages:       subscription.RetainAckedMessages,
// 		RetentionDuration:         subscription.RetentionDuration,
// 		ExpirationPolicy:          subscription.ExpirationTTL,
// 		EnableExactlyOnceDelivery: subscription.DeliverExactlyOnce,
// 		DeadLetterPolicy:          deadLetterPolicy,
// 		RetryPolicy:               retryPolicy,
// 	}
// }
