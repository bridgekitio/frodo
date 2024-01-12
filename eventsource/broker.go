package eventsource

import (
	"context"
	"strings"
	"time"
)

// Broker encapsulates a connection/client that can both publish events and subscribe to them.
type Broker interface {
	Publisher
	Subscriber
}

// Publisher is an event source client that can write events to the source/stream.
type Publisher interface {
	// Publish broadcasts the given event to the broker. This is usually done asynchronously
	// in order to limit the amount of blocking you do waiting for this.
	//
	// You have no guarantee that any subscribers will receive this message even if there
	// is no error returned. Errors should only describe an inability to give the message
	// to the broker.
	Publish(ctx context.Context, key string, payload []byte) error
}

// Subscriber is a broker client/connection that lets you subscribe to asynchronous events
// that occur elsewhere in the system.
type Subscriber interface {
	// Subscribe creates a one-off listener that will fire your handler function for
	// EVERY instance of the event/key.
	Subscribe(key string, handlerFunc EventHandlerFunc) (Subscription, error)

	// SubscribeGroup creates a listener that is a member of a "Consumer Group". If there
	// are other listeners in the same 'group', only one of them should have their
	// handler function fired.
	//
	// This is akin to a "Consumer Group" if you are from the Kafka world or "Queue Group" if
	// NATS is more of your jam.
	SubscribeGroup(key string, group string, handlerFunc EventHandlerFunc) (Subscription, error)
}

// Subscription is simply a registration pointer that can allow you to stop listening at any time.
type Subscription interface {
	// Unsubscribe notifies the Broker/Subscriber that created this subscription that we
	// want to stop receiving events. Typically, this is done during process shutdown automatically.
	Unsubscribe() error
}

// EventHandlerFunc is the signature for a function that can asynchronously handle an incoming event
// that was published through a broker and subscribed to by a listener.
type EventHandlerFunc func(ctx context.Context, evt *EventMessage) error

// EventMessage is the message/envelope that brokers use to deliver events to subscribers.
type EventMessage struct {
	// Timestamp indicates when the event was fired/published.
	Timestamp time.Time
	// Key is the identifier of the event (e.g. "UserService.Created").
	Key string
	// Payload is the ALREADY-ENCODED data you want to send to any listeners/subscribers. It
	// is the job of the layer on top of this to decide on appropriate encoding/decoding practices.
	Payload []byte
}

// Namespace returns the portion of an event key that occurs before the first period. This
// only applies if there is a period in your key - if not, this function assumes that there
// is no namespace. It's just a free agent key. Not all implementations will like this.
//
//	Namespace("UserCreated")        // ""
//	Namespace("User.Created")       // "User"
//	Namespace("User.Created.Error") // "User"
func Namespace(key string) string {
	namespace, _, ok := strings.Cut(key, ".")
	if ok {
		return namespace
	}
	return ""
}
