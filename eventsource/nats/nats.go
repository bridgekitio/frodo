package nats

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bridgekitio/frodo/eventsource"
	"github.com/nats-io/nats.go"
)

var ErrNotConnected = fmt.Errorf("not connected")
var ErrInvalidNamespace = fmt.Errorf("key does not have valid namespace: e.g. 'usercreated' instead of 'user.created'")

func Broker(options ...Option) eventsource.Broker {
	c := client{
		uri:               "nats://127.0.0.1:4222",
		mutex:             &sync.Mutex{},
		streams:           map[string]*nats.StreamInfo{},
		retentionMaxAge:   7 * 24 * time.Hour,
		retentionMaxMsgs:  -1,
		retentionMaxBytes: -1,
	}
	for _, option := range options {
		option(&c)
	}

	if c.conn, c.err = nats.Connect(c.uri); c.err != nil {
		c.err = fmt.Errorf("nats connect error: %w", c.err)
		return &c
	}
	if c.jetstream, c.err = c.conn.JetStream(); c.err != nil {
		c.err = fmt.Errorf("nats jetstream error: %w", c.err)
		return &c
	}
	return &c
}

type client struct {
	uri   string
	err   error
	mutex *sync.Mutex

	streams       map[string]*nats.StreamInfo
	subscriptions []subscription

	retentionMaxAge   time.Duration
	retentionMaxMsgs  int64
	retentionMaxBytes int64

	conn      *nats.Conn
	jetstream nats.JetStreamContext
}

func (c *client) Publish(ctx context.Context, key string, payload []byte) error {
	if err := c.loadStream(key); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}

	if _, err := c.jetstream.Publish(key, payload, nats.Context(ctx)); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	return nil
}

func (c *client) Subscribe(key string, handlerFunc eventsource.EventHandlerFunc) (eventsource.Subscription, error) {
	if err := c.loadStream(key); err != nil {
		return nil, fmt.Errorf("nats subscribe: %w", err)
	}

	sub, err := c.jetstream.Subscribe(key, c.toMsgHandler(handlerFunc))
	if err != nil {
		return nil, fmt.Errorf("nats subscribe: %w", err)
	}
	return subscription{sub: sub}, nil
}

func (c *client) SubscribeGroup(key string, group string, handlerFunc eventsource.EventHandlerFunc) (eventsource.Subscription, error) {
	if err := c.loadStream(key); err != nil {
		return nil, fmt.Errorf("nats subscribe group: %w", err)
	}

	// NATS doesn't like periods in consumer group names, so convert them to underscores.
	group = strings.ReplaceAll(group, ".", "_")

	sub, err := c.jetstream.QueueSubscribe(key, group, c.toMsgHandler(handlerFunc))
	if err != nil {
		return nil, fmt.Errorf("nats subscribe group: %w", err)
	}
	return subscription{sub: sub}, nil
}

func (c *client) toMsgHandler(handlerFunc eventsource.EventHandlerFunc) nats.MsgHandler {
	return func(m *nats.Msg) {
		err := handlerFunc(context.Background(), &eventsource.EventMessage{
			Timestamp: time.Now(),
			Key:       m.Subject,
			Payload:   m.Data,
		})
		if err != nil {
			fmt.Printf("[WARN] error handling subscription: %v: %v\n", m.Subject, err)
		}
	}
}

func (c *client) loadStream(key string) error {
	if c.err != nil {
		return c.err
	}
	if c.jetstream == nil {
		return ErrNotConnected
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Keys are usually things like "FooService.SaveBar", so we want just
	// the "FooService" bit as the service name.
	namespace := eventsource.Namespace(key)
	if namespace == "" {
		return ErrInvalidNamespace
	}

	// You are probably going to call Publish/Subscribe(Group) a bunch of times for the same
	// stream, so only round-trip all the way to NATS the first time. Also, if the info is
	// in this map, we've already applied any changes to the
	if _, ok := c.streams[namespace]; ok {
		return nil
	}

	switch info, err := c.jetstream.StreamInfo(namespace); err {
	// The stream already exists in NATS. Now we need to see if you updated any desired
	// settings on the stream such as the retention policy. If so, we need to push those
	// changes to NATS before moving on.
	case nil:
		if info, err = c.updateStream(info); err != nil {
			return fmt.Errorf("load stream: %w", err)
		}
		c.streams[namespace] = info
		return nil

	// We connected and NATS responded just fine. There's just no stream by this name,
	// so let's make one. We will name the stream after the service (e.g. "UserService")
	// and make sure that it accepts subjects matching any of the method
	// names (e.g. "UserService.*").
	case nats.ErrStreamNotFound:
		info, err = c.jetstream.AddStream(&nats.StreamConfig{
			Name:      namespace,
			Subjects:  []string{namespace + ".*"},
			Retention: nats.LimitsPolicy,
			MaxAge:    c.retentionMaxAge,
			MaxBytes:  c.retentionMaxBytes,
			MaxMsgs:   c.retentionMaxMsgs,
		})
		c.streams[namespace] = info
		return err

	// Shit got fucky. You're going to have a bad time.
	default:
		return fmt.Errorf("load stream: %w", err)
	}
}

func (c *client) updateStream(info *nats.StreamInfo) (*nats.StreamInfo, error) {
	// Doesn't look like you've changed the configuration since you last started
	// up the service. Just leave everything alone and move on.
	if c.retentionMaxAge == info.Config.MaxAge &&
		c.retentionMaxBytes == info.Config.MaxBytes &&
		c.retentionMaxMsgs == info.Config.MaxMsgs {
		return info, nil
	}

	config := info.Config
	config.MaxAge = c.retentionMaxAge
	config.MaxBytes = c.retentionMaxBytes
	config.MaxMsgs = c.retentionMaxMsgs
	return c.jetstream.UpdateStream(&config)
}

type subscription struct {
	sub *nats.Subscription
}

func (s subscription) Unsubscribe() error {
	return s.sub.Unsubscribe()
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
