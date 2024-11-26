package nats

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bridgekit-io/frodo/eventsource"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var ErrNotConnected = fmt.Errorf("nats broker not connected")
var ErrInvalidNamespace = fmt.Errorf("key does not have valid namespace: e.g. 'usercreated' instead of 'user.created'")

// Broker creates a new event broker that distributes messages using NATS JetStream queues/groups.
func Broker(options ...Option) eventsource.Broker {
	c := client{
		uri:               nats.DefaultURL,
		mutex:             &sync.Mutex{},
		streams:           map[string]jetstream.Stream{},
		retentionMaxAge:   7 * 24 * time.Hour,
		retentionMaxMsgs:  -1,
		retentionMaxBytes: -1,
		inactiveThreshold: 1 * time.Hour,
		onError: func(err error) {
			log.Printf("[nats broker error] %v\n", err)
		},
	}
	for _, option := range options {
		option(&c)
	}

	if c.conn, c.err = nats.Connect(c.uri, nats.UserInfo(c.username, c.password)); c.err != nil {
		c.err = fmt.Errorf("nats connect error: %w", c.err)
		return &c
	}
	if c.js, c.err = jetstream.New(c.conn); c.err != nil {
		c.err = fmt.Errorf("nats jetstream error: %w", c.err)
		return &c
	}
	return &c
}

type client struct {
	uri     string
	err     error
	mutex   *sync.Mutex
	onError func(error)

	streams       map[string]jetstream.Stream
	subscriptions []subscription
	username      string
	password      string

	retentionMaxAge   time.Duration
	retentionMaxMsgs  int64
	retentionMaxBytes int64
	inactiveThreshold time.Duration

	conn *nats.Conn
	js   jetstream.JetStream
}

func (c *client) Publish(ctx context.Context, key string, payload []byte) error {
	if _, err := c.connectStream(ctx, key); err != nil {
		return fmt.Errorf("broker publish error: %w", err)
	}
	if _, err := c.js.Publish(ctx, key, payload); err != nil {
		return fmt.Errorf("broker publish error: %w", err)
	}
	return nil
}

func (c *client) Subscribe(ctx context.Context, key string, handlerFunc eventsource.EventHandlerFunc) (eventsource.Subscription, error) {
	return c.consume(ctx, key, "", handlerFunc)
}

func (c *client) SubscribeGroup(ctx context.Context, key string, group string, handlerFunc eventsource.EventHandlerFunc) (eventsource.Subscription, error) {
	return c.consume(ctx, key, group, handlerFunc)
}

func (c *client) consume(ctx context.Context, key string, group string, handlerFunc eventsource.EventHandlerFunc) (eventsource.Subscription, error) {
	stream, err := c.connectStream(ctx, key)
	if err != nil {
		return nil, err
	}

	// NATS doesn't like periods in the names of things. Message key/subject is fine, but the names of streams/consumers is not cool.
	group = strings.ReplaceAll(group, ".", "_")

	// The one-hour timeout doesn't delete messages older than an hour. It just auto-cleans up the metadata about a consumer, so
	// if you create the same group 2 hours later, NATS will just treat it like this is the first time its ever seen this group.
	// For now, we only support a DeliverLastPolicy, so it will start delivering after the most recent message anyway. We can
	// revisit this if we want to support "catch-up" style broker behavior where instances will process all messages that
	// came in while the consuming service was down. But... we don't do that yet, so this always-act-like-its-new approach
	// is good enough for now.
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:           group,
		InactiveThreshold: c.inactiveThreshold,
		DeliverPolicy:     jetstream.DeliverNewPolicy,
		FilterSubject:     key,
	})
	if err != nil {
		return nil, fmt.Errorf("broker consumer error: %w", err)
	}

	consumerContext, err := consumer.Consume(c.toJetStreamMessageHandler(handlerFunc))
	if err != nil {
		return nil, fmt.Errorf("broker consumer context error: %w", err)
	}
	return subscription{consumer: consumer, consumerContext: consumerContext}, nil
}

func (c *client) toJetStreamMessageHandler(handlerFunc eventsource.EventHandlerFunc) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		if err := msg.Ack(); err != nil {
			c.onError(fmt.Errorf("error during ack for '%s': %w", msg.Subject(), err))
			return
		}

		err := handlerFunc(context.Background(), &eventsource.EventMessage{
			Timestamp: time.Now(),
			Key:       msg.Subject(),
			Payload:   msg.Data(),
		})
		if err != nil {
			c.onError(fmt.Errorf("error handling event message '%s': %w", msg.Subject(), err))
		}
	}
}

// connectStream lazy creates the event stream where this key is supposed to go. Keys look like "FooService.BarMethod",
// so this will create/update a stream named "FooService".
func (c *client) connectStream(ctx context.Context, key string) (jetstream.Stream, error) {
	if c.js == nil {
		return nil, ErrNotConnected
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Each service gets its own "stream" of events. No matter how methods you have, they'll all go into the same
	// event stream for that service. The consumers (subscribers) will filter out just the methods they're interested in.
	// The key we receive is the fully-qualified service method like "CalculatorService.Add". The stream name will be
	// CalculatorService, so its Add, Subtract, Multiply, and other methods will all get published in there.
	serviceName := eventsource.Namespace(key)
	if serviceName == "" {
		return nil, ErrInvalidNamespace
	}

	// You are probably going to call Publish/Subscribe(Group) a bunch of times for the same
	// stream, so only round-trip all the way to NATS the first time.
	if stream, ok := c.streams[serviceName]; ok {
		return stream, nil
	}

	stream, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      serviceName,
		Subjects:  []string{serviceName + ".>"},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    c.retentionMaxAge,
		MaxBytes:  c.retentionMaxBytes,
		MaxMsgs:   c.retentionMaxMsgs,
	})
	if err != nil {
		return nil, fmt.Errorf("broker event stream error: %w", err)
	}

	c.streams[serviceName] = stream
	return stream, nil
}

type subscription struct {
	consumer        jetstream.Consumer
	consumerContext jetstream.ConsumeContext
}

func (s subscription) Close() error {
	s.consumerContext.Drain()
	return nil
}

type Option func(c *client)

func WithAddress(address string) Option {
	return func(c *client) {
		// Allow us to accept addresses that contain or don't contain the "nats://" protocol prefix.
		switch before, after, ok := strings.Cut(address, "://"); ok {
		case true:
			c.uri = "nats://" + after
		case false:
			c.uri = "nats://" + before
		}
	}
}

func WithMaxAge(ttl time.Duration) Option {
	return func(c *client) {
		c.retentionMaxAge = ttl
	}
}

func WithMaxBytes(maxBytes int64) Option {
	return func(c *client) {
		c.retentionMaxBytes = maxBytes
	}
}

func WithMaxMsgs(maxMsgs int64) Option {
	return func(c *client) {
		c.retentionMaxMsgs = maxMsgs
	}
}

func WithInactiveThreshold(threshold time.Duration) Option {
	return func(c *client) {
		c.inactiveThreshold = threshold
	}
}

func WithUserInfo(username, password string) Option {
	return func(c *client) {
		c.username = username
		c.password = password
	}
}
