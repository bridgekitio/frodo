//go:build unit

package local_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bridgekitio/frodo/eventsource"
	"github.com/bridgekitio/frodo/eventsource/local"
	"github.com/bridgekitio/frodo/internal/testext"
	"github.com/bridgekitio/frodo/internal/wait"
	"github.com/stretchr/testify/suite"
)

func TestLocalBroker(t *testing.T) {
	suite.Run(t, new(LocalBrokerSuite))
}

type LocalBrokerSuite struct {
	suite.Suite
}

func (suite *LocalBrokerSuite) TestPublish_canceledContext() {
	broker := local.Broker()

	// Canceled explicitly
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	suite.Error(broker.Publish(ctx, "Foo", []byte("Hello")))

	// Canceled due to deadline
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	suite.Error(broker.Publish(ctx, "Foo", []byte("Hello")))
}

func (suite *LocalBrokerSuite) TestPublish_noSubscribers() {
	broker := local.Broker()
	suite.NoError(broker.Publish(context.Background(), "Foo", []byte("Hello")))
	suite.NoError(broker.Publish(context.Background(), "Bar", []byte("Goodbye")))
	suite.NoError(broker.Publish(context.Background(), "Baz", []byte("Seriously, go home.")))
}

func (suite *LocalBrokerSuite) publish(broker eventsource.Broker, key string, value string) {
	msg := "Publishing with a valid context, should always succeed"
	suite.Require().NoError(broker.Publish(context.Background(), key, []byte(value)), msg)
}

func (suite *LocalBrokerSuite) subscribe(broker eventsource.Broker, sequence *testext.Sequence, key string) eventsource.Subscription {
	subs, err := broker.Subscribe(key, func(ctx context.Context, evt *eventsource.EventMessage) error {
		if string(evt.Payload) == "error" {
			return fmt.Errorf("nope")
		}

		sequence.Append(fmt.Sprintf("%s:%s", key, string(evt.Payload)))
		sequence.WaitGroup().Done()
		return nil
	})
	suite.Require().NoError(err, "There shouldn't be any issues subscribing locally... ever.")
	return subs
}

func (suite *LocalBrokerSuite) subscribeGroup(broker eventsource.Broker, sequence *testext.Sequence, key string, group string, which string) eventsource.Subscription {
	subs, err := broker.SubscribeGroup(key, group, func(ctx context.Context, evt *eventsource.EventMessage) error {
		if string(evt.Payload) == "error" {
			return fmt.Errorf("nope")
		}

		sequence.Append(fmt.Sprintf("%s:%s:%s:%s", key, group, which, string(evt.Payload)))
		sequence.WaitGroup().Done()
		return nil
	})
	suite.Require().NoError(err, "There shouldn't be any issues subscribing locally... ever.")
	return subs
}

func (suite *LocalBrokerSuite) assertFired(sequence *testext.Sequence, expected []string) {
	// This sucks, but there's no other way for us to determine if all of the handlers have
	// finished their work. It's all small, in-memory lists, so this should be more than
	// enough time to be sure that the sequence contains the handler values.
	// time.Sleep(25 * time.Millisecond)

	wait.WithTimeout(sequence.WaitGroup(), 5*time.Second)
	suite.ElementsMatch(expected, sequence.Values())
}

func (suite *LocalBrokerSuite) TestPublish_noMatching() {
	results := &testext.Sequence{}
	broker := local.Broker()
	suite.subscribe(broker, results, "Foo")
	suite.subscribe(broker, results, "Foo.Bar")

	suite.NoError(broker.Publish(context.Background(), "Foo.Foo", []byte("Hello")))
	suite.NoError(broker.Publish(context.Background(), "Bar", []byte("Goodbye")))
	suite.NoError(broker.Publish(context.Background(), "Baz", []byte("Seriously, go home.")))

	time.Sleep(10 * time.Millisecond)
	suite.Len(results.Values(), 0, "None of the event handlers should have fired")
}

// Mostly ensures that wildcards adhere to the fact that "*" can only match a single token in a key. For instance,
// the key "*" matches "Foo" but not "Foo.Bar" whereas "*.*" matches "Foo.Bar", but not "Foo".
func (suite *LocalBrokerSuite) TestPublish_subscription_wildcards() {
	results := &testext.Sequence{}
	broker := local.Broker()
	suite.subscribe(broker, results, "*")
	suite.subscribe(broker, results, "*.*")
	suite.subscribe(broker, results, "*.*.*")

	suite.subscribeGroup(broker, results, "*", "1", "")
	suite.subscribeGroup(broker, results, "*", "1", "")
	suite.subscribeGroup(broker, results, "*", "2", "")

	suite.subscribeGroup(broker, results, "*.*", "1", "")
	suite.subscribeGroup(broker, results, "*.*", "1", "")
	suite.subscribeGroup(broker, results, "*.*", "2", "")

	suite.subscribeGroup(broker, results, "*.*.*", "1", "")
	suite.subscribeGroup(broker, results, "*.*.*", "2", "")
	suite.subscribeGroup(broker, results, "*.*.*", "2", "")

	results.ResetWithWorkers(3)
	suite.publish(broker, "Foo", "A")
	suite.assertFired(results, []string{
		"*:A",
		"*:1::A",
		"*:2::A",
	})

	results.ResetWithWorkers(3)
	suite.publish(broker, "Foo.Bar", "A")
	suite.assertFired(results, []string{
		"*.*:A",
		"*.*:1::A",
		"*.*:2::A",
	})

	results.ResetWithWorkers(3)
	suite.publish(broker, "Foo.Bar.Baz", "A")
	suite.assertFired(results, []string{
		"*.*.*:A",
		"*.*.*:1::A",
		"*.*.*:2::A",
	})
}

// Ensures that wildcards work when used as keys for publishing messages.
func (suite *LocalBrokerSuite) TestPublish_publish_wildcards() {
	results := &testext.Sequence{}
	broker := local.Broker()
	suite.subscribe(broker, results, "Foo")
	suite.subscribe(broker, results, "Foo.Bar")
	suite.subscribe(broker, results, "Foo.Bar.Baz")

	suite.subscribeGroup(broker, results, "Foo", "1", "")
	suite.subscribeGroup(broker, results, "Foo", "1", "")
	suite.subscribeGroup(broker, results, "Foo", "2", "")

	suite.subscribeGroup(broker, results, "Foo.Bar", "1", "")
	suite.subscribeGroup(broker, results, "Foo.Bar", "1", "")
	suite.subscribeGroup(broker, results, "Foo.Bar", "2", "")

	suite.subscribeGroup(broker, results, "Foo.Bar.Baz", "1", "")
	suite.subscribeGroup(broker, results, "Foo.Bar.Baz", "2", "")
	suite.subscribeGroup(broker, results, "Foo.Bar.Baz", "2", "")

	results.ResetWithWorkers(3)
	suite.publish(broker, "*", "A")
	suite.assertFired(results, []string{
		"Foo:A",
		"Foo:1::A",
		"Foo:2::A",
	})
	results.ResetWithWorkers(3)
	suite.publish(broker, "*.*", "A")
	suite.assertFired(results, []string{
		"Foo.Bar:A",
		"Foo.Bar:1::A",
		"Foo.Bar:2::A",
	})
	results.ResetWithWorkers(3)
	suite.publish(broker, "*.*.*", "A")
	suite.assertFired(results, []string{
		"Foo.Bar.Baz:A",
		"Foo.Bar.Baz:1::A",
		"Foo.Bar.Baz:2::A",
	})
}

func (suite *LocalBrokerSuite) TestPublish_matching() {
	results := &testext.Sequence{}
	broker := local.Broker()
	suite.subscribe(broker, results, "Foo")
	suite.subscribe(broker, results, "Bar")
	suite.subscribe(broker, results, "Foo.Bar")
	suite.subscribe(broker, results, "Foo.*")
	suite.subscribe(broker, results, "*")
	suite.subscribe(broker, results, "*.*")
	suite.subscribe(broker, results, "Foo.Bar.Goo")
	suite.subscribe(broker, results, "Foo.*.Goo")

	results.ResetWithWorkers(2)
	suite.publish(broker, "Foo", "A")
	suite.assertFired(results, []string{
		"Foo:A",
		"*:A",
	})

	results.ResetWithWorkers(6)
	suite.publish(broker, "Foo.Bar", "A")
	suite.publish(broker, "Foo.Bar", "B")
	suite.assertFired(results, []string{
		"Foo.Bar:A",
		"Foo.Bar:B",
		"Foo.*:A",
		"Foo.*:B",
		"*.*:A",
		"*.*:B",
	})

	results.ResetWithWorkers(7)
	suite.publish(broker, "Bar", "A")
	suite.publish(broker, "Bar.Baz", "B")
	suite.publish(broker, "Hello.World", "C")
	suite.publish(broker, "Foo.Bar.Goo", "D")
	suite.publish(broker, "Foo.Baz.Goo", "E")
	suite.publish(broker, "Nope.Nope.Nope", "F")
	suite.assertFired(results, []string{
		"Bar:A",
		"*:A",
		"*.*:B",
		"*.*:C",
		"Foo.Bar.Goo:D",
		"Foo.*.Goo:D",
		"Foo.*.Goo:E",
	})

	// Multiple subscribers to the same event should ALL get the event.
	results.ResetWithWorkers(4)
	suite.subscribe(broker, results, "Foo")
	suite.subscribe(broker, results, "Foo")
	suite.publish(broker, "Foo", "A")
	suite.assertFired(results, []string{
		"Foo:A",
		"Foo:A",
		"Foo:A",
		"*:A", // there's three explicit Foo handlers, but only one * handler
	})
}

// Ensure that the correct subscribers and groups fire when mixed together.
func (suite *LocalBrokerSuite) TestPublish_mixedGroups() {
	results := &testext.Sequence{}
	broker := local.Broker()

	// One group handling Foo.
	suite.subscribeGroup(broker, results, "Foo", "1", "")
	suite.subscribeGroup(broker, results, "Foo", "1", "")

	// Another group handling Foo.
	suite.subscribeGroup(broker, results, "Foo", "2", "")
	suite.subscribeGroup(broker, results, "Foo", "2", "")
	suite.subscribeGroup(broker, results, "Foo", "2", "")
	suite.subscribeGroup(broker, results, "Foo", "2", "")
	suite.subscribeGroup(broker, results, "Foo", "2", "")
	suite.subscribeGroup(broker, results, "Foo", "2", "")

	// Another generic group that should get every 1-token key.
	suite.subscribeGroup(broker, results, "*", "3", "")

	// Some one-off handlers that should get everything.
	suite.subscribe(broker, results, "Foo")
	suite.subscribe(broker, results, "Foo")

	results.ResetWithWorkers(11)
	suite.publish(broker, "Foo", "A")
	suite.publish(broker, "Foo", "B")
	suite.publish(broker, "Bar", "C")
	suite.publish(broker, "Foo.Bar", "D") // nothing matches this

	suite.assertFired(results, []string{
		// Group matches
		"Foo:1::A",
		"Foo:2::A",
		"*:3::A",
		"Foo:1::B",
		"Foo:2::B",
		"*:3::B",
		"*:3::C",

		// Individual listener matches
		"Foo:A",
		"Foo:A",
		"Foo:B",
		"Foo:B",
	})
}

func (suite *LocalBrokerSuite) TestPublish_groupRoundRobin() {
	results := &testext.Sequence{}
	broker := local.Broker()
	suite.subscribeGroup(broker, results, "Foo", "1", "0")
	suite.subscribeGroup(broker, results, "Foo", "1", "1")
	suite.subscribeGroup(broker, results, "*", "1", "2")
	suite.subscribeGroup(broker, results, "Foo", "2", "3") // different group

	results.ResetWithWorkers(21)
	suite.publish(broker, "Foo", "A")
	suite.publish(broker, "Foo", "B")
	suite.publish(broker, "Foo", "C")
	suite.publish(broker, "Foo", "D")
	suite.publish(broker, "Foo", "E")
	suite.publish(broker, "Foo", "F")
	suite.publish(broker, "Foo", "G")
	suite.assertFired(results, []string{
		// The "1" group listening for the "Foo" topic should round-robin after each message.
		"Foo:1:0:A",
		"Foo:1:1:B",
		"Foo:1:0:C",
		"Foo:1:1:D",
		"Foo:1:0:E",
		"Foo:1:1:F",
		"Foo:1:0:G",

		// Even though it's the same "1" group, the "*" topic is going to treat this as a different stream
		// of results. Not sure if this is the best behavior, but that's what it is for now.
		"*:1:2:A",
		"*:1:2:B",
		"*:1:2:C",
		"*:1:2:D",
		"*:1:2:E",
		"*:1:2:F",
		"*:1:2:G",

		// The "2" group listening for the "Foo" has only one member and it receives messages independent of "1"
		"Foo:2:3:A",
		"Foo:2:3:B",
		"Foo:2:3:C",
		"Foo:2:3:D",
		"Foo:2:3:E",
		"Foo:2:3:F",
		"Foo:2:3:G",
	})
}

// Publishing should still work even if subscribers fail.
func (suite *LocalBrokerSuite) TestPublish_subscriberErrors() {
	results := &testext.Sequence{}
	broker := local.Broker(local.WithErrorHandler(func(err error) {
		results.Append("oops")
		results.WaitGroup().Done()
	}))

	suite.subscribeGroup(broker, results, "Foo", "1", "")
	suite.subscribeGroup(broker, results, "Foo", "1", "")
	suite.subscribeGroup(broker, results, "Foo", "2", "")
	suite.subscribe(broker, results, "Foo")
	suite.subscribe(broker, results, "Foo")
	suite.subscribe(broker, results, "*")

	// The value "error" is not actually added to the sequence and the handlers return a non-nil error.
	results.ResetWithWorkers(11)
	suite.publish(broker, "Foo", "error")
	suite.publish(broker, "Foo", "error")
	suite.publish(broker, "Bar", "error")
	suite.assertFired(results, []string{
		// Errors from publish #1
		"oops",
		"oops",
		"oops",
		"oops",
		"oops",
		// Errors from publish #2
		"oops",
		"oops",
		"oops",
		"oops",
		"oops",
		// Errors from publish #3 (only * matched)
		"oops",
	})
}

func (suite *LocalBrokerSuite) TestUnsubscribe() {
	results := &testext.Sequence{}
	broker := local.Broker()

	s1 := suite.subscribe(broker, results, "Foo")
	s2 := suite.subscribeGroup(broker, results, "Foo", "1", "")
	s3 := suite.subscribe(broker, results, "*")

	results.ResetWithWorkers(3)
	suite.publish(broker, "Foo", "A")
	suite.assertFired(results, []string{
		"*:A",
		"Foo:A",
		"Foo:1::A",
	})

	results.ResetWithWorkers(2)
	suite.NoError(s1.Unsubscribe())
	suite.publish(broker, "Foo", "B")
	suite.assertFired(results, []string{
		"*:B",
		"Foo:1::B",
	})

	results.ResetWithWorkers(1)
	suite.NoError(s2.Unsubscribe())
	suite.publish(broker, "Foo", "C")
	suite.assertFired(results, []string{
		"*:C",
	})

	results.ResetWithWorkers(0)
	suite.NoError(s3.Unsubscribe())
	suite.publish(broker, "Foo", "D")
	suite.assertFired(results, []string{})

	// Should be able to add more back in after the fact
	results.ResetWithWorkers(1)
	suite.subscribe(broker, results, "Foo")
	suite.publish(broker, "Foo", "I'm Back!")
	suite.assertFired(results, []string{
		"Foo:I'm Back!",
	})
}

// This test ensures that no matter what context you pass to broker.Publish(), that context does NOT propagate to
// your subscriber(s). They should each get a clean context to begin their execution.
func (suite *LocalBrokerSuite) TestSubscriberContext() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	broker := local.Broker()

	publishCtx := context.WithValue(context.Background(), "foo", "bar")

	_, _ = broker.Subscribe("test.context", func(ctx context.Context, evt *eventsource.EventMessage) error {
		foo, _ := ctx.Value("foo").(string)
		suite.Require().Equal("", foo, "Subscriber context should not be the same as publisher context")

		wg.Done()
		return nil
	})

	_, _ = broker.Subscribe("test.context", func(ctx context.Context, evt *eventsource.EventMessage) error {
		foo, _ := ctx.Value("foo").(string)
		suite.Require().Equal("", foo, "Subscriber context should not be the same as publisher context")

		wg.Done()
		return nil
	})

	err := broker.Publish(publishCtx, "test.context", []byte("Isolate Contexts"))
	suite.Require().NoError(err, "Isolating subscriber context shouldn't break publishing.")

	wg.Wait()
}
