# Frodo

*That's just, like, your opinion, man!*

Frod is an opinionated code generator and runtime library that helps you
write (micro) services/APIs that supports, both, RPC/HTTP
and Event-Driven invocation. It parses the
interfaces/structs/comments in your code service
code to generate all of the client, server, gateway, and pub-sub
communication code automatically.

You write business logic. Frodo generates the annoying
copy/paste boilerplate needed to expose your service as an
HTTP API as well as Pub/Sub code to create event-driven
workflows across your services.

This is the spiritual successor to my original project of the same name in a different org: [Frodo](https://github.com/monadicstack/frodo).
It supports the RPC/HTTP related features of the original, but it addresses
many shortcomings in the architecture/approach and adds
Event-Driven communication with almost no extra code on your part.

## Getting Started

```shell
go install github.com/bridgekitio/frodo@latest
go get -u github.com/bridgekitio/frodo
```
This will fetch the `frodo` code generation executable then add the
runtime libraries that allow your services and clients to
communicate with each other as a dependency to your project.

## Basic Example

We're going to write a simple `CalculatorService` that
lets you either add or subtract two numbers.

### Step 1: Describe Your Service

Your first step is to write a .go file that just defines
the contract for your service; the interface as well as the
inputs/outputs.

```go
// calc/calculator_service.go

package calc

import (
    "context"
)

type CalculatorService interface {
    Add(context.Context, *AddRequest) (*AddResponse, error)
    Sub(context.Context, *SubRequest) (*SubResponse, error)
}

type AddRequest struct {
    A int
    B int
}

type AddResponse struct {
    Result int
}

type SubRequest struct {
    A int
    B int
}

type SubResponse struct {
    Result int
}
```

One important detail is that the interface name ends with
the suffix "Service". This tells Frodo that this is an
actual service interface and not just some random abstraction
in your code.

At this point you haven't actually defined *how* this service gets
this work done; just which operations are available.

### Step 2: Implement Your Service Logic

We actually have enough for Frodo to
generate your RPC/API/Event code already, but we'll hold off
for a moment. Frodo frees you up to focus on building
features, so let's actually implement service - no networking,
no marshaling, no status stuff, no pub/sub - just business logic to make your
service behave properly.

```go
// calc/calculator_service_handler.go

package calc

import (
    "context"
)

type CalculatorServiceHandler struct {}

func (svc CalculatorServiceHandler) Add(ctx context.Context, req *AddRequest) (*AddResponse, error) {
    result := req.A + req.B
    return &AddResponse{Result: result}, nil
}

func (svc CalculatorServiceHandler) Sub(ctx context.Context, req *SubRequest) (*SubResponse, error) {
    result := req.A - req.B
    return &SubResponse{Result: result}, nil
}
```

### Step 3: Generate Your RPC Client and Server Code

At this point, you've just written the same code that you (hopefully)
would have written even if you weren't using Frodo. Next,
we want to auto-generate two things:

* The "server" bits that allow an instance of your CalculatorService
  to listen for incoming requests from an either HTTP API or
  a published event. (We'll look at events later...)
* A "client" struct that communicates with that API to get work done.

Just run these two commands in a terminal:

```shell
# Feed it the service interface file, not the handler.
frodo server calculator_service.go
frodo client calculator_service.go
```

### Step 4: Run Your Calculator API

Let's fire up an HTTP server on port 9000 that makes your service
available for consumption.

```go
package main

import (
    "github.com/bridgekitio/frodo/services"
    "github.com/bridgekitio/frodo/services/gateways/apis"

    "github.com/your/project/calc"
    calcgen "github.com/your/project/calc/gen"
)

func main() {
    // Create your logic-only handler, then wrap it in service
    // communication bits that let it interact with the Frodo runtime.
    calcHandler := calc.CalculatorServiceHandler{}
    calcService := calcgen.CalculatorServiceServer(calcHandler)
	
    // Fire up a server that will manage our service and listen to 
    // API calls on port 9000.
    server := services.NewServer(
        services.Listen(apis.NewGateway(":9000")),
        services.Register(calcService),
    )
    server.Run()
}
```

Seriously. That's the whole program.

Compile and run it, and your service/API is now ready
to be consumed. We'll use the Go client we generated in just
a moment, but you can try this out right now by simply
using curl:

```shell
curl -d '{"A":5, "B":2}' http://localhost:9000/CalculatorService.Add
# {"Result":7}
curl -d '{"A":5, "B":2}' http://localhost:9000/CalculatorService.Sub
# {"Result":3}
```

### Step 5: Interact With Your Calculator Service

While you can use raw HTTP to communicate with the service,
let's use our auto-generated client to hide the gory
details of JSON marshaling, status code translation, and
other noise.

The client actually implements CalculatorService
just like the server/handler does. As a result the RPC-style
call will "feel" like you're executing the service work
locally, when in reality the client is actually making API
calls to the server running on port 9000.

```go
package main

import (
    "context"
    "fmt"

    "github.com/your/project/calc"
    calcgen "github.com/your/project/calc/gen"
)

func main() {
    ctx := context.Background()
    client := calcgen.CalculatorServiceClient("http://localhost:9000")

    add, err := client.Add(ctx, &calc.AddRequest{A:5, B:2})
    if err != nil {
        // handle error	
    }
    fmt.Println("5 + 2 =", add.Result)

    sub, err := client.Sub(ctx, &calc.SubRequest{A:5, B:2})
    if err != nil {
        // handle error	
    }
    fmt.Println("5 - 2 =", sub.Result)
}
```

Compile/run this program, and you should see the following output:

```
5 + 2 = 7
5 - 2 = 3
```
That's it!

## Creating a JavaScript Client

The `frodo` tool can actually generate a JS client that you
can add to your frontend code (or React Native mobile code)
to hide the complexity of making API calls to your backend
service. Without any plugins or fuss, we can create a JS client of the same
CalculatorService from earlier...

```shell
frodo client calc/calculator_service.go --language=js
    or
frodo client calc/calculator_service.go --language=node
```

This will create the file `calculator_service.gen.client.js`
which you can include with your frontend codebase. Using it
should look similar to the Go client we saw earlier:

```js
import {CalculatorService} from 'lib/calculator_service.gen.client';

// The service client is a class that exposes all of the
// operations as 'async' functions that resolve with the
// result of the service call.
//
// All of the operations on the Go service are exposed here
// as well. The arguments to these functions are the same as
// the request struct in the Go code, and the return value
// will match the response struct from your Go service.
const service = new CalculatorService('http://localhost:9000');
const add = await service.Add({A:5, B:2});
const sub = await service.Sub({A:5, B:2});

// Should print:
// Add(5, 2) = 7
// Sub(5, 2) = 3
console.info('Add(5, 2) = ' + add.Result)
console.info('Sub(5, 2) = ' + sub.Result)
```

Another subtle benefit of using the generated client is that your
service/method documentation follows you in the generated code.
It's included in the file as JSDoc comments so your
documentation should be available to your IDE even when writing
your frontend code.

#### Node Support

Frodo uses the `fetch` function to make the actual HTTP requests,
so if you are using Node 18+, you shouldn't need to do anything
special as `fetch` is now in the global scope. If that's the
case, ignore the next paragraph and subsequent sample code.

If you're using an older version of node or just really prefer
to use the classic `node-fetch` package, you can supply the
fetch implementation to use when constructing your client:

```js
const fetch = require('node-fetch');

const service = new CalculatorService('http://localhost:9000', {fetch});
const add = await service.Add({A:5, B:2});
const sub = await service.Sub({A:5, B:2});
```

## Creating a Dart/Flutter Client

Just like the JS client, Frodo can create a Dart client that you can embed
in your Flutter apps so mobile frontends can consume your service.

```shell
frodo client calc/calculator_service.go --language=dart
  or
frodo client calc/calculator_service.go --language=flutter
```

This will create the file `calculator_service.gen.client.dart`. Add it
to your Flutter codebase, and it behaves very similarly to the JS client.

> The `HttpClient` from the standard `dart:io` package is NOT supported
> in Flutter web applications. To support Flutter mobile as well as web,
> Frodo clients uses the [http](https://pub.dev/packages/http) package to
> make requests to the backend API. You'll need to add that to your
> pubspec for the following code to work:

```dart
import 'lib/calculator_service.gen.client.dart';

var service = CalculatorServiceClient("http://localhost:9000");
var add = await service.Add(AddRequest(A:5, B:2));
var sub = await service.Sub(SubRequest(A:5, B:2));

// Should print:
// Add(5, 2) = 7
// Sub(5, 2) = 3
print('Add(5, 2) = ${add.Result}');
print('Sub(5, 2) = ${sub.Result}');
```

For more examples of how to write services that let Frodo take
care of the RPC/API boilerplate, take a look in the [example/](https://github.com/bridgekitio/frodo/tree/main/example)
directory of this repo.

## Adding Event-Driven Methods

RPC-style communication works for lots of scenarios, but sometimes
you want loosely-coupled workflows that fire when other operations
in the system complete. For instance, let's say that after a user
places an order, you want the system to send them an order confirmation
email as well as send them a coupon for a future order.

Frodo makes it super easy to wire these events up. Here is what
your service interface would look like. And yes, you'd probably
put email-related operations in a different service, but we
just want to see how to wire up event-driven service invocation.
This setup works equally well if you split this up, but we'll
look at multiple service setups later.

```go
type OrderService interface {
    // PlaceOrder... places an order.
    //
    // HTTP 202
    // POST /orders
    PlaceOrder(context.Context, *PlaceOrderRequest) (*PlaceOrderResponse, error)

    // SendConfirmation sends an email confirmation to the user.
    //
    // ON OrderService.PlaceOrder
    SendConfirmation(context.Context, *SendConfirmationRequest) (*SendConfirmationResponse, error)

    // SendCoupon sends a 5% off similar item coupon to the user based on the order.
    //
    // ON OrderService.PlaceOrder
    SendCoupon(context.Context, *SendCouponRequest) (*SendCouponResponse, error)
}

type PlaceOrderResponse struct {
    OrderID  string
    UserID   string
    ItemIDs  string[]
    DateTime time.Time
}

type SendConfirmationRequest struct {
    OrderID string
    UserID  string
}

type SendCouponRequest struct {
    OrderID string
    UserID  string
}
```

The next section will go over Doc Options in more detail, but
just looking at the code, it should be somewhat obvious what
we're going for. We will invoke both "send" methods automatically be
any time there's a successful call to `PlaceOrder`.

When `PlaceOrder` finishes, Frodo automatically publishes an
`OrderService.PlaceOrder` event with the response value.
Both `SendXXX` methods receive that event and build
their request structs automatically. They'll fill in `OrderID`
and `UserID`, but they'll just ignore `ItemIDs` and `DateTime`
because they don't have equivalent fields for those.

If that makes sense, notice that the only thing you did
differently than before was adding that line in the comments.
That's all the info that Frodo needs to wire that behavior up for you!

There is one more one-line change we need to make in order for
this to work. That's in `main()` when we set up our server.
Before we only told the server to listen for requests via an API
Gateway. Now we need to tell it to also listen for requests via
an Event Gateway:

```go
func main() {
    // Create the handler and service, exact same as before...
    orderHandler := orders.OrderServiceHandler{}
    orderService := ordersgen.OrderServiceServer(orderHandler)

    // Now, the service can accept requests via the HTTP API
    // OR from events wired up using the 'ON' doc option.
    server := services.NewServer(
        services.Listen(apis.NewGateway(":9000")),
        services.Listen(events.NewGateway()), // <--- This is the only difference.
        services.Register(orderService),
    )
    server.Run()
}
```

Now you'd expect something like this when running this code:

```shell
curl -d '{...}' http://localhost:9000/orders
# curl response 
{
  "OrderID": "123",
  "UserID": "456",
  "ItemIDs": ["789"],
  "DateTime": "2022-12-17T17:00:23+00:00"
}
# and you should have 2 emails in your inbox
```

### Distributed Events Using NATS JetStream

The order example above works great if you're running everything
in one process as a monolith. By default, the Event Gateway uses an in-memory
event broker to publish and react to events fired by your services. 


If you want to write this as a
distributed system with multiple remote instances and services, however, you
will need some third party event broker to manage this. Frodo ships with support
for using [NATS JetStream](https://docs.nats.io/nats-concepts/jetstream) out-of-the-box.

```go
import (
    // ... other imports ...
    "github.com/bridgekitio/frodo/eventsource/nats"
)

func main() {
    // Create the handler and service, exact same as before...
    orderHandler := orders.OrderServiceHandler{}
    orderService := ordersgen.OrderServiceServer(orderHandler)
	
    // Configure a NATS client to distribute events.
    natsBroker := nats.Broker(
        nats.WithAddress("nats://127.0.0.1:4222"),
        nats.WithMaxAge(24 * time.Hours),
    )
	
    // Tell the event gateway to use NATS instead of local queues. Notice that
    // we're still listening for HTTP/RPC requests as well.
    server := services.NewServer(
        services.Listen(apis.NewGateway(":9000")),
        services.Listen(events.NewGateway(events.WithBroker(natsBroker))),
        services.Register(orderService),
    )
    server.Run()
}
```

Now, you can run 20 different instances if you like, and the
events will be spread around to all of them rather than always being
handled by the instance that placed the order.

### A Word About "Consumer Groups"

If you were to run 20 instances of the `OrderService`, you're not going to
blast of 20 copies of each email. The NATS broker will create two Queue Groups
(consumer groups to use the Kafka-style terminology), one named "OrderService.SendConfirmation"
and another named "OrderService.SendCoupon". This means that when the place order event
fires, both groups will receive a copy of the event - BUT - only one of the 20 instances
will handle the event for the confirmation group and only one of the 20 instances
will handle the event for the coupon group. As a result, you can have as many loosely
coupled units of work fire while still scaling out your infrastructure.

## Doc Options: Custom URLs, Status, etc

Frodo gives you a service/API that "just works" out of the
box. By default, endpoints follow a similar RPC/POST style used by lots of
other service libraries/frameworks.

You can, however customize the API routes for individual operations,
set a prefix for all routes in a service, and more using "Doc Options"...
worst Spider-Man villain ever.

Here's an example with most of the available options. They are all
independent, so you can specify a custom status without specifying
a custom route and so on.

```go
// CalculatorService provides some basic arithmetic operations.
//
// VERSION 0.1.3
// PATH /v1
type CalculatorService interface {
    // Add calculates the sum of A + B.
    //
    // HTTP 202
    // GET /sum/{A}/{B}
    Add(context.Context, *AddRequest) (*AddResponse, error)

    // Sub calculates the difference of A - B.
    //
    // GET /difference/{A}/{B}
    Sub(context.Context, *SubRequest) (*SubResponse, error)
	
    // CountCalls is NOT exposed in your HTTP API. It is, however
    // called after every single successful call to either Add or Sub
    // in this service. It will even fire when FixWord is called in
    // a completely different service!
    //
    // ROLES admin.read
    // HTTP OMIT
    // ON CalculatorService.Add
    // ON CalculatorService.Sub
    // ON SpellingService.FixWord
    CountCalls(context.Context, *CountCallsRequest) (*CountCallsResponse, error)
}
```

#### Service: PATH

This prepends your custom value on every route in the API. It applies
to the standard `ServiceName.FunctionName` routes as well as custom routes
as we'll cover in a moment.

Your generated API and RPC clients will be auto-wired to use the prefix "v1" under the
hood, so you don't need to change your code any further. If you want
to hit the raw HTTP endpoints, however, here's how they look now:

```shell
curl -d '{"A":5, "B":2}' http://localhost:9000/v1/CalculatorService.Add
# {"Result":7}

curl -d '{"A":5, "B":2}' http://localhost:9000/v1/CalculatorService.Sub
# {"Result":3}
```

#### Function: GET/POST/PUT/PATCH/DELETE

You can replace the default `POST ServiceName.FunctionName` route for any
operation with the route of your choice. In the example, the path parameters `{A}` and `{B}`
will be bound to the equivalent A and B attributes on the request struct.

Here are the updated curl calls after we generate the new
gateway code. Notice it's also taking into account the service's PATH
prefix as well:

```shell
curl http://localhost:9000/v1/sum/5/2
# {"Result":7}
curl http://localhost:9000/v1/difference/5/2
# {"Result":3}
```

Use these options to your heart's content if you want your API
to feel more REST-ful instead of RPC-ful.

#### Method: HTTP {StatusCode}

This lets you have the API return a non-200 status code on success.
For instance, the Add function's route will return a `202 Accepted`
status when it responds with the answer instead of `200 OK`.

Since we didn't specify anything special for the Sub method, it
will continue to respond with `200 OK`, same as before.

#### Method: HTTP OMIT

Sometimes you want your service to be able to perform operations
that you don't want to expose to the outside world. Perhaps this
only fires asynchronously when some event fires (next section)
or it's just some private code you want to manually execute but not
allow external access.

If the operation has `HTTP OMIT`, Frodo will not create an API
route for it. It will not appear in your OpenAPI docs or external
language clients (like JS and Dart). The method *will* still
appear in your Go client because we need to satisfy the service
interface, but you'll receive a 404 error if you attempt to
invoke it.

#### Method: ON {ServiceName.MethodName}

This is what we used in the previous section to allow services
to trigger workflow events. The format is always `ON ServiceName.MethodName`.
This is true even if you provide a custom HTTP API route. The
event name is ALWAYS the same no matter what.

As you can see in the example above, you can have as many `ON`
triggers as you want on a single method, and they do not even
need to be from the same service!

#### Method: ROLES roleA,roleB,roleC

Similar to the version number on your service, this option doesn't alter the
behavior of your service at all. What it does provide, however, is a way to apply
some dynamic metadata to your request context that you can use to make your authorization
layer easier to implement.

Let's say you had a service operation that can only be run by system admins or
users with rights to edit the group you're modifying:

```go
// RenameGroup changes a group's display name.
//
// PUT /group/{ID}/name
// HTTP 200
// ROLES admin.write, group.{ID}.write
func (svc *GroupService) RenameGroup(ctx context.Context, req *RenameGroupRequest) (*RenameGroupResponse, error) {
    // This slice will have 2 elements in it. The values will depend on the URL used
    // to hit this API endpoint. If you `PUT /group/123/name` then the slice will contain
    // the values "admin.write" and "group.123.write". If you `PUT /group/789/name` then
    // you'll have "group.789.write" as the second value.
    //
    // Frodo will automatically bind the "{ID}" value in the role just like it did
    // when populating your request struct. It follows the exact same binding rules, so
    // if you know how to customize URL routes, you can have exact roles to check for.
    requiredRoles := metadata.Route(ctx).Roles

    // Not Frodo stuff... this is your black box that gets info about the authenticated caller
    // and checks to see that it contains any of the 'requiredRoles'
    if !authorization.HasAnyRole(ctx, requiredRoles) {
        return nil, fail.Forbidden('no soup for you!')
    }

    // ... do the normal renaming work you'd normally do ...
}
```

Because Frodo automatically populates path variables in your role names, you can
very easily make a single `AuthorizationMiddleware` that works for all endpoints in
all your services. This way you're not copy/pasting these 3 lines to high hell. Yay!

At some point Frodo might get even more opinionated and provide ways to carry this info
around, but for now that's an exercise for the user.

## Error Handling

By default, if your service call returns a non-nil error, the
resulting RPC/HTTP request will have a 500 status code. You
can, however, customize that status code to correspond to the type
of failure (e.g. 404 when something was not found).

The easiest way to do this is to just use Frodo's `fail`
package when you encounter a failure case:

```go
import (
    "github.com/bridgekitio/frodo/fail"
)

func (svc UserService) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
    if req.ID == "" {
        return nil, fail.BadRequest("id is required")
    }
    user, err := svc.Repo.GetByID(req.ID)
    if err != nil {
    	return nil, err
    }
    if user == nil {
        return nil, fail.NotFound("user not found: %s", req.ID)
    }
    return &GetResponse{User: user}, nil
}
```

In this case, the caller will receive an HTTP 400 if they
didn't provide an id, a 404 if there is no user with that
id, and a 500 if any other type of error occurs.

#### Customizing Errors

While the error categories in the `fail` package are
probably good enough for most people, you can build your own
custom status-bound errors by simply having them implement the
`StatusCode() int` function:

```go
type RateLimitError struct {
    Limit int
}

func (err RateLimitError) StatusCode() int {
    return 429
}

func (err RateLimitError) Error() string {
    return fmt.Sprintf("limit of %d/sec exceeded", err.Limit)
}
```

Now when you implement a handler or middleware, you can simply
return your custom error type and have your API respond w/
a 429 instead of a generic 500 error:

```go
func (svc UserService) CreateToken(ctx context.Context, req *CreateTokenRequest) (*CreateTokenResponse, error) {
    if (svc.exceededLimit(ctx, 5)) {
        return nil, RateLimitError{Limit: 5}
    }
    return &CreateTokenResponse{Token: "Hello"}, nil
}

// Sample call:
// curl -XPOST http://localhost:9000/UserService.CreateToken
// {
//    "StatusCode": 429,
//    "Message": "limit of 5/sec exceeded"
// }
```

### Errors In Async Event Handlers

Handling errors in RPC calls is fairly easy. The clients that
Frodo generate return the error. Simple.

When using the `ON Service.Method` option to trigger calls
based on events, you don't really have control over that code, so
we need to do something a little different to handle errors
that might occur during those asynchronous flows.

You can give the Event Gateway a callback function that Frodo will
invoke any time an error occurs processing event-based service operations.

```go
func main() {
    // ...
    server.Listen(events.NewGateway(
        events.WithBroker(natsBroker),
        events.WithErrorHandler(handleEventError),
    ))
    server.Run()
}

func handleEventError(err error) {
    // Don't panic...
}
```

## Middleware

You'll find that you frequently have work that you want to execute
before/after every single service invocation regardless of whether
it came from the API or some event. Frodo uses continuation passing
functions similar to what you see in the most popular Go HTTP middleware
libraries.

```go
func main() {
    // Every service call to the CalculatorService will write to
    // the log and track how long it took.
    calcHandler := calc.CalculatorServiceHandler{}
    calcService := calcgen.CalculatorServiceServer(calcHandler,
        LogRequest,
        CollectTiming,
    )

    // No changes here...
    server := services.NewServer(
        services.Listen(apis.NewGateway(":9000")),
        services.Listen(events.NewGateway()),
        services.Register(calcService),
    )
    server.Run()
}

func LogRequest(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
    route := metadata.Route(ctx)
    fmt.Printf("Invoking %s.%s\n", route.ServiceName, route.Name)
    res, err := next(ctx, req)
    fmt.Printf(" > failed %v\n", err != nil)
    return res, err
}

func CollectTiming(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
    start := time.Now()
    res, err := next(ctx, req)
    elapsed := time.Now().Sub(start)

    route := metadata.Route(ctx)
    metricsCollector.StoreTiming(route.QualifiedName(), elapsed)
    return res, err
}
```

#### HTTP Middleware

Most of your middleware should be done at the service level like
we have seen above: authorization, logging, observability, etc. They're
all things that are important regardless of whether we're servicing
an API call or an event.

One of Frodo's primary goals is to make it so that you never have
to think about HTTP or transport code, but there are still times
when there's no getting around it. If you want to consume your
service in a web application, you're going to need to set up
CORS and that has to be done at the HTTP level.

Luckily, you can provide HTTP-level middleware when calling
`apis.NewGateway()`. The middleware function it expects are
compatible with [Negroni](https://github.com/urfave/negroni), so
you have an entire ecosystem of off-the-shelf handlers to plug in.

```go
func main() {
    // We'll still log and capture metrics on every call.
    calcHandler := calc.CalculatorServiceHandler{}
    calcService := calcgen.CalculatorServiceServer(calcHandler,
        LogRequest,
        CollectTiming,
    )

    server := services.NewServer(
        services.Listen(apis.NewGateway(":9000",
            apis.WithMiddleware(
                negroni.NewLogger().ServeHTTP,
                cors.New().ServeHTTP,
                gzip.New().ServeHTTP,
            ),
        )),
        services.Listen(events.NewGateway()),
        services.Register(calcService),
    )
    server.Run()
}
```

## Metadata

When you make an RPC call from Service A to Service B, values
stored on the `context.Context` will NOT be available to you when
are in Service B's handler. There are
instances, however, where it's useful to have data follow
every hop from service to service; trace ids, authorization, etc.

Frodo uses the `metadata` package to store all manner of values for
the entire request; even if that request hits multiple services.

### Metadata: Authorization

You probably want your services to have some level of access control,
so incoming HTTP calls will likely have the "Authorization" header set.
Your middleware and handler functions don't work at the HTTP level, so the
`metadata` package captures that and makes it available for you.

```go
func CheckAdminMiddleware(ctx context.Context, req any, next services.Handler) (any, error) {
    auth := metadata.Authorization(ctx)
    if auth != "Bearer 12345" {
        return nil, fail.PermissionDenied("admins only")
    }
    return next(ctx, req)
}
```

Ignore the horrifyingly bad security - ultimately the service
should behave like this:

```shell
curl -H "Authorization: Guest" http://localhost:9000/AdminService.DropTables
{"StatusCode":403, "Message": "admins only"}

curl -H "Authorization: Bearer 12345" http://localhost:9000/AdminService.DropTables
{"TablesDropped":42}
```

#### Supplying Authorization Credentials

In the previous example we assumed that there was some primordial HTTP request with
an Authorization header. Since authorization is just a value stored
on the context, you can supply them fairly easily when using
the auto-generated service clients - again, the goal is to avoid
worrying about the transport layer:

```go
// Go: Using the generated service client
client := admingen.AdminServiceClient("http://localhost:9000")
ctx = metadata.WithAuthorization(ctx, "Bearer 12345")
client.DropTables(ctx, &admin.DropTablesRequest{Name:"*"})
```

```js
// JS: Using the generated service client
const client = new AdminClient('...');
const req = {Name: '*'};
client.DropTables(req, {authorization: 'Bearer 12345'});
```

```dart
// Dart: Using the generated service client
var client = AdminClient('...');
var req = DropTablesRequest(Name: '*');
client.DropTables(req, authorization: 'Token 12345');
```

### Metadata: Trace ID

If you ever want to be able to debug/observe behaviors in your
system, you'll need a consistent request/trace id to tie back
to every operation. For instance, if you place an order and then
that triggers 4 other operations (emails, analytics, etc.), Frodo
manages a trace id that will be the same across all of those
related operations.

> Frodo will honor any X-Request-ID header it receives,
so if your service is behind a load balancer or some proxy that
generates that HTTP header, that is the Trace ID that Frodo will use.
If not, Frodo will generate a unique value for you so that you always
have a meaningful Trace ID.

```go
func (svc FooService) Foo(ctx context.Context, req *FooRequest) (*FooResponse, error) {
    traceID := metadata.TraceID(ctx)
    fmt.Printf(">> Foo: %s\n", traceID)

    // By using the same context, the trace id is passed along.
    barServiceClient.Bar(ctx, &BarRequest{})
    // ...
}

func (svc BarService) Bar(ctx context.Context, req *BarRequest) (*BarResponse, error) {
    traceID := metadata.TraceID(ctx)
    fmt.Printf(">> Bar: %s\n", traceID)
    // ...
}

// The interface for BazService.Bar had this:
// ON BarService.Bar
func (svc BazService) Bar(ctx context.Context, req *BarRequest) (*BarResponse, error) {
    traceID := metadata.TraceID(ctx)
    fmt.Printf(">> Baz: %s\n", traceID)
    // ...
}
```
Here's the output of our console when we make that initial
call to the Foo operation:

```shell
# When your service receives an explicit request id:
curl -H "X-Request-ID: Hello12345" -XPOST http://localhost:9000/FooService.Foo
# Console output
>> Foo: Hello12345
>> Bar: Hello12345
>> Baz: Hello12345

# There's no explicit id, so we'll just create one:
curl -XPOST http://localhost:9000/FooService.Foo
# Console output
>> Foo: dGhpcyBpcyBhIHJlYWxseSBs
>> Bar: dGhpcyBpcyBhIHJlYWxseSBs
>> Baz: dGhpcyBpcyBhIHJlYWxseSBs
```

It doesn't matter how many hops your request takes or whether
they were RPC calls or event-based calls. Your trace id follows you.

### Metadata: Values

Although Frodo manages some very specific fields with very specific
purposes, the `metadata` package lets you store a general purpose
map of values that you deem as important. Just like
authorization or trace ids, these values will be accessible by
subsequent service calls for the same request.

```go
func (svc ServiceA) Foo(ctx context.Context, req *FooRequest) (*FooResponse, error) {
    // "Hello" will NOT follow you when you call Bar(),
    // but "DontPanic" will. Notice that the metadata
    // value does not need to be a string like in gRPC.
    ctx = context.WithValue(ctx, "Hello", "World")
    ctx = metadata.WithValue(ctx, "DontPanic", Answer{Value: 42})

    serviceB.Bar(ctx, &BarRequest{})
}

func (b ServiceB) Bar(ctx context.Context, req *BarRequest) (*BarResponse, error) {
    valueA, okA := ctx.Value("Hello").(string)

    valueB := Answer{}
    okB = metadata.Value(ctx, "DontPanic", &b)
    
    // valueA  == ""               okA == false
    // valueB == Answer{Value:42}  okB == true
}

// Pretend that your ServiceC interface had this option on Baz:
// ON ServiceA.Foo
func (c ServiceC) Baz(ctx context.Context, req *BazRequest) (*BazResponse, error) {
    valueA, okA := ctx.Value("Hello").(string)

    valueB := Answer{}
    okB = metadata.Value(ctx, "DontPanic", &b)

    // valueA  == ""               okA == false
    // valueB == Answer{Value:42}  okB == true
}
```

If you're wondering why `metadata.Value()` looks more like
`json.Unarmsahl()` than `context.Value()`, it has to
do with a limitation of reflection in Go. When the values
are sent over the network from Service A to Service B/C, we
lose all type information. We need the type info `&b` gives
us in order to properly restore the original value, so Frodo
follows the idiom established by many
of the decoders in the standard library.

## Returning Raw File Data

Let's say that you're writing ProfilePictureService. One of the operations
you might want is the ability to return the raw JPG data for a user's profile
picture. You do this the same way that you handle JSON-based responses; just
implement some specialized interfaces so that Frodo knows to treat it a little different:

```go
type ServeResponse struct {
    file *io.File
}

// By implementing services.ContentGetter, the response tells Frodo to
// respond w/ raw data rather than JSON. Instead of turning the struct into
// JSON, grab bytes from this reader and deliver them in the response.
func (res ServeResponse) Content() io.ReadCloser {
    return res.file
}

// By implementing services.ContentTypeGetter, this lets you dictate the
// underlying HTTP Content-Type header. Without this Frodo will have
// nothing to go on and assume "application/octet-stream".
func (res ServeResponse) ContentType() string {
    return "image/jpeg"
}

// --- and now in your service ---

func (svc *ProfilePictureService) Serve(ctx context.Context, req *ServeRequest) (*ServeResponse, error) {
    // Ignore the fact that you probably don't store profile pictures on the
    // hard drive of your service boxes...
    f, err := os.Open("./pictures/" + req.UserID + ".jpg")
    if err != nil {
        return nil, fail.NotFound("no profile picture for user %s", req.UserID)
    }
    return &ServeResponse{file: f}, nil
}
```

## HTTP Redirects

It's fairly common to have a service call that does some work to locate a
resource, authorize it, and then redirect to S3, CloudFront, or some other
CDN to actually serve up the raw asset.

With Frodo, it's pretty simple. If your XxxResponse struct implements the
`services.Redirector` interface then the API gateway will respond with a
307-style redirect to the URL of your choice:

```go
// In video_service.go, this implements the services.Redirector interface.
type DownloadResponse struct {
    Bucket string
    Key    string	
}

// Redirect returns the raw HTTP URL that the client should redirect to once it
// receives this response. This is the implementation of services.Redirector.
func (res DownloadResponse) Redirect() string {
    return fmt.Sprintf("https://%s.s3.amazonaws.com/%s",
        res.Bucket,
        res.Key)
}

// ...

// This triggers a 307-style redirect to the URL returned by response.Redirect()
func (svc VideoServiceHandler) Download(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
    file := svc.Repo.Get(req.FileID)
    return &DownloadResponse{Bucket: file.Bucket, Key: file.Key}, nil
}
```

## Running Multiple Services

One of the core ideas behind Frodo is that you should build your services in an isolated,
decoupled manner regardless of how you intend to deploy them. Frodo gives you the 
flexibility to write your services once, and you can choose to either run them separately
as micro/mini services. Alternately, you can take all of the services and run them in
a single process as a monolith.

### To Run Them As a Monolith

```go
// Initialize your raw service handlers just like you normally would.
userService := userGen.UserServiceServer(userHandler)
groupService := groupGen.GroupServiceServer(groupHandler)
mailService := mailGen.MailServiceServer(mailHandler)
orderService := orderGen.MailServiceServer(orderHandler)

// All 4 services will listen on on port 9000 for incoming HTTP requests, and
// they will all listen for each others' events and react apporpriately.
server := services.NewServer(
    services.Listen(apis.NewGateway(":9000")),
    services.Listen(events.NewGateway()),
    services.Register(userService, groupService, mailService, orderService),
)
server.Run()
```

### To Run Them Is Micro/Mini Services

```go
// In users/cmd/main.go
userService := userGen.UserServiceServer(userHandler)
server := services.NewServer(
    services.Listen(apis.NewGateway(":9001")),
    services.Listen(events.NewGateway(events.WithBroker(natsBroker))),
    services.Register(userService),
)
server.Run()

// In groups/cmd/main.go
groupService := groupGen.GroupServiceServer(groupHandler)
server := services.NewServer(
    services.Listen(apis.NewGateway(":9002")),
    services.Listen(events.NewGateway(events.WithBroker(natsBroker))),
    services.Register(groupService),
)
server.Run()

// And follow the pattern for the other 2 services...
```

This flexibility requires no fundamental change to the way you write your services - only how
you start them up in `main()`. It's great when you're starting up a new project/business
where you want to start small/cheap and validate your idea. Run it as a monolith in the
beginning; then as you need to scale, you can peel off services and run them by themselves
or have process A run two services and process B run the other two services.

Ultimately you get to decouple your deployment from your software design which leads to
better, more resilient code.

> One small note about the micro/mini service example. The events your services
> publish will need to cross process boundaries because they're not running together.
> The might be separate process on the same machine or different machines entirely.
> As a result, if you plan to use the `events` gateway, you'll need to use the NATS
> broker since the default broker only communicates with services in the same process.

## Go Generate Support

If you prefer to stick to the standard Go toolchain for generating code, you can use
`//go:generate` comments to hook the Frodo code generator into your build process. 

For example, this generates the server/gateway, mock service, Go client, JS client,
and Flutter/Dart client, and OpenAPI documentation just by marking up your service
definition a bit:

```go
import (
   ...
)

//go:generate frodo server  $GOFILE
//go:generate frodo client  $GOFILE
//go:generate frodo client  $GOFILE --language=js
//go:generate frodo client  $GOFILE --language=flutter
//go:generate frodo docs    $GOFILE
//go:generate frodo mock    $GOFILE

// CalculatorService provides basic arithmetic operations.
//
// VERSION 1.0.0
// PREFIX  v1
type CalculatorService interface {
    ...
}
```

## Mocking Services

Using mocks is a divisive topic, and I'm not here to tell you the right/wrong way to
test your code. If you prefer mocks, Frodo can generate helpful mock implementations
of your services to use in your tests. Using a similar command that we used for 
generating our server and clients, you can do the following:

```shell
frodo mock calculator_service.go
```

That creates `gen/calculator_service.gen.mock.go` which you can use in your test
suites like so:

```go
import (
    "context"
    "fmt"

    "github.com/example/calc"
    mocks "github.com/example/calc/gen"
)

func TestSomethingThatDependsOnAddFailure(t *testing.T) {
    // You can program behaviors for Add(). If the test code calls Sub()
    // it will panic since you didn't define a behavior for that operation.
    svc := mocks.MockCalculatorService{
        AddFunc: func(ctx context.Context, req *calc.AddRequest) (*calc.AddResponse, error) {
            return nil, fmt.Errorf("barf...")
        },	
    }

    // Feed your mock service to the thing you're testing
    something := NewSomething(svc)
    _, err := something.BlahBlah(100)
    assertError(err)
    ...

    // You can also verify invocations on your service:
    assertEquals(0, svc.Calls.Sub.Times)
    assertEquals(5, svc.Calls.Add.Times)
    assertEquals(1, svc.Calls.Add.TimesFor(calc.Request{A: 4, B: 2}))
    assertEquals(2, svc.Calls.Add.TimesMatching(func(r calc.Request) bool {
        return r.A > 2
    }))
}
```

## Generate OpenAPI/Swagger Documentation (Experimental)

Definitely a work in progress, but in addition to generating your backend and
frontend assets, Frodo can generate OpenAPI 3.0 YAML files to describe your API.
It uses the name/type information from your Go code as well as the GoDoc comments
that you (hopefully) write. Document your code in Go and you can get online API docs
for free:

```shell
frodo docs calculator_service.go
```

Now you can feed the file gen/calculator_service.gen.swagger.yaml to your favorite 
Swagger tools. You can try it out by just pasting the output on
https://editor.swagger.io.

OpenAPI docs let you specify the current version of your service. You can specify
that value by including the VERSION doc option on your service interface.

```go
// FooService is a magical service that does awesome things.
//
// VERSION 1.2.1
type FooService interface {
    // ...
}
```

Now, when you generate your docs the version badge will display "1.2.1".

Not gonna lie... this whole feature is still a work in progress. I've still
got some issues to work out with nested request/response structs. It spits out enough
good stuff that it should describe your services better than no documentation at all,
though.

## FAQs

### Why a separate repo/project? Why not do a Frodo version 2?

I hate Go the major versioning scheme for Go modules. I never understood the
widespread dislike of that choice until I ran into it myself. It's silly
having to make sure that people `go install` the URL that ends in `/v2` instead
of the more natural root package.

While this project solves the same core issue that the original Frodo did, I had to change the API
significantly support event driven workflows and to better handle the multi-service case in the section above.
Try as I might, every attempt to fit event
driven stuff into Frodo's original runtime code felt hacky and wrong. I needed a
version 2 but Go did us all dirty with versioning. So, you can continue to use the old one
to your heart's content. I have no intention of taking it down, but this is the one I'm
going to maintain moving forward.

### Why does Frodo only support NATS for event-driven flows?

Well, I had to start somewhere. NATS is written in Go, it's stupid simple to
set up, and satisfies most use cases, so it seemed like the natural way to go.
I may add support for Redis and (maybe) Kafka in the future, but for now NATS is
the only officially supported event driven client.
