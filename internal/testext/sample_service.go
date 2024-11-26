package testext

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bridgekit-io/frodo/services"
)

//go:generate ../../out/frodo server  $GOFILE --force
//go:generate ../../out/frodo client  $GOFILE --force
//go:generate ../../out/frodo client  $GOFILE --force --language=js
//go:generate ../../out/frodo client  $GOFILE --force --language=dart
//go:generate ../../out/frodo mock    $GOFILE --force
//go:generate ../../out/frodo docs    $GOFILE --force

// SampleService is a mix of different options, parameter setups, and responses so that we can
// run integration tests using our code-generated clients. Each method is nothing special, but
// they each do something a little differently than the rest to flex different parts of the framework.
//
// PREFIX /v2
type SampleService interface {
	// Defaults simply utilizes all of the framework's default behaviors.
	Defaults(context.Context, *SampleRequest) (*SampleResponse, error)

	// ComplexValues flexes our ability to encode/decode non-flat structs.
	ComplexValues(context.Context, *SampleComplexRequest) (*SampleComplexResponse, error)

	// ComplexValuesPath flexes our ability to encode/decode non-flat structs while
	// specifying them via path and query string.
	//
	// GET /complex/values/{InUser.ID}/{InUser.Name}/woot
	ComplexValuesPath(context.Context, *SampleComplexRequest) (*SampleComplexResponse, error)

	// Fail4XX always returns a non-nil 400-series error.
	Fail4XX(context.Context, *SampleRequest) (*SampleResponse, error)

	// Fail5XX always returns a non-nil 500-series error.
	Fail5XX(context.Context, *SampleRequest) (*SampleResponse, error)

	// CustomRoute performs a service operation where you override default behavior
	// by providing routing-related Doc Options.
	//
	// HTTP 202
	// GET /custom/route/1/{ID}/{Text}
	CustomRoute(context.Context, *SampleRequest) (*SampleResponse, error)

	// CustomRouteQuery performs a service operation where you override default behavior
	// by providing routing-related Doc Options. The input data relies on the path
	//
	// HTTP 202
	// GET /custom/route/2/{ID}
	CustomRouteQuery(context.Context, *SampleRequest) (*SampleResponse, error)

	// CustomRouteBody performs a service operation where you override default behavior
	// by providing routing-related Doc Options, but rely on body encoding rather than path.
	//
	// HTTP 201
	// PUT /custom/route/3/{ID}
	CustomRouteBody(context.Context, *SampleRequest) (*SampleResponse, error)

	// OmitMe exists in the service, but should be excluded from the public API.
	//
	// HTTP OMIT
	OmitMe(ctx context.Context, request *SampleRequest) (*SampleResponse, error)

	// Download results in a raw stream of data rather than relying on auto-encoding
	// the response value.
	//
	// GET /download
	Download(context.Context, *SampleDownloadRequest) (*SampleDownloadResponse, error)

	// DownloadResumable results in a raw stream of data rather than relying on auto-encoding
	// the response value. The stream includes Content-Range info as though you could resume
	// your stream/download progress later.
	//
	// GET /download/resumable
	DownloadResumable(context.Context, *SampleDownloadRequest) (*SampleDownloadResponse, error)

	// Redirect results in a 307-style redirect to the Download endpoint.
	//
	// GET /redirect
	Redirect(context.Context, *SampleRedirectRequest) (*SampleRedirectResponse, error)

	// Authorization regurgitates the "Authorization" metadata/header.
	Authorization(context.Context, *SampleRequest) (*SampleResponse, error)

	// Sleep successfully responds, but it will sleep for 5 seconds before doing so. Use this
	// for test cases where you want to try out timeouts.
	Sleep(context.Context, *SampleRequest) (*SampleResponse, error)

	/*
	 Event based endpoints
	*/

	// TriggerUpperCase ensures that events still fire as "SampleService.TriggerUpperCase" even though
	// we are going to set a different HTTP path.
	//
	// GET /Upper/Case/WootyAndTheBlowfish
	TriggerUpperCase(context.Context, *SampleRequest) (*SampleResponse, error)
	TriggerLowerCase(context.Context, *SampleRequest) (*SampleResponse, error)
	TriggerFailure(context.Context, *SampleRequest) (*SampleResponse, error)

	// ListenerA fires on only one of the triggers.
	//
	// GET /ListenerA/Woot
	// ON SampleService.TriggerUpperCase
	ListenerA(context.Context, *SampleRequest) (*SampleResponse, error)

	// ListenerB fires on multiple triggers... including another event-based endpoint. We also
	// listen for the TriggerFailure event which should never fire properly.
	//
	// HTTP OMIT
	// ON SampleService.TriggerUpperCase
	// ON SampleService.TriggerLowerCase
	// ON SampleService.TriggerFailure
	// ON SampleService.ListenerA
	// ON OtherService.SpaceOut
	ListenerB(context.Context, *SampleRequest) (*SampleResponse, error)

	// FailAlways will return an error no matter what. It's only goal in life is to trigger OnFailAlways.
	FailAlways(ctx context.Context, request *FailAlwaysRequest) (*FailAlwaysResponse, error)

	// OnFailAlways should trigger after FailAlways inevitably shits the bed.
	//
	// ON SampleService.FailAlways:Error
	OnFailAlways(ctx context.Context, request *FailAlwaysErrorRequest) (*FailAlwaysErrorResponse, error)

	// Chain1 kicks off the Chain1/Chain2/Chain3 event chain, but we expect that it's going to stop after
	Chain1(ctx context.Context, request *SampleRequest) (*SampleResponse, error)

	// Chain2 ALWAYS FAILS, SO CHAIN3 NEVER FIRES!!!
	//
	// ON SampleService.Chain1
	Chain2(ctx context.Context, request *SampleRequest) (*SampleResponse, error)

	// Chain2OnSuccess never fires. It listens for the success of Chain2, but since that always fails, this
	// should never be triggered, so tests should never have this in its output.
	//
	// HTTP OMIT
	// ON SampleService.Chain2
	Chain2OnSuccess(ctx context.Context, request *SampleRequest) (*SampleResponse, error)

	// Chain2OnError listens for errors that occur on calls to Chain2
	//
	// HTTP OMIT
	// ON SampleService.Chain2:Error
	Chain2OnError(ctx context.Context, request *FailAlwaysErrorRequest) (*FailAlwaysErrorResponse, error)

	// Chain1GroupStar listens for calls to Chain1, but rather than being part of a consumer group that only lets
	// one instance of the service run it, it should define its own group that lets EVERY instance of this service
	// react to this event.
	//
	// ON SampleService.Chain1 GROUP *
	Chain1GroupStar(ctx context.Context, request *SampleRequest) (*SampleResponse, error)

	// Chain1GroupFooBar listens for calls to Chain1, but rather than being part of a consumer group that only lets
	// one instance of the service run it, it should define its own shared group name.
	//
	// ON SampleService.Chain1 GROUP FooBar
	Chain1GroupFooBar(ctx context.Context, request *SampleRequest) (*SampleResponse, error)

	// SecureWithRoles lets us test role based security by looking at the 'roles' doc option.
	//
	// ROLES admin.write,user.{ID}.write ,   user.{User.ID}.admin, junk.{NotReal}.crap
	SecureWithRoles(context.Context, *SampleSecurityRequest) (*SampleSecurityResponse, error)

	// SecureWithRolesAliased lets us test role based security by looking at the 'roles' doc option. Specifically,
	// we make sure we can resolve role segments with string alias types, not just strings.
	//
	// ROLES admin.write,user.{FancyID}.write ,   user.{User.FancyID}.admin, junk.{NotReal}.crap
	SecureWithRolesAliased(context.Context, *SampleSecurityRequest) (*SampleSecurityResponse, error)

	// Panic um... panics. It never succeeds. It always behaves like me when I'm on a high place looking down.
	Panic(context.Context, *SampleRequest) (*SampleResponse, error)
}

type SampleRequest struct {
	ID   string
	Text string
}

type SampleResponse SampleRequest

// SampleUser contains an array of different fields that we support sending to/from clients
// in all of our supported languages.
type SampleUser struct {
	// ID is a string value that will likely have no whitespace.
	ID string
	// FancyID makes sure that we can use aliases properly rather than just the raw primitive types.
	FancyID StringLike
	// Name is a string value that will likely have spaces.
	Name string
	// Age is a numeric value that we should support.
	Age int
	// Attention is a duration to ensure that we use epoch nanos as the format, NOT the string.
	Attention time.Duration
	// AttentionString is a custom duration alias that overrides MarshalJSON/UnmarshalJSON to use strings for transport.
	AttentionString CustomDuration
	// PhoneNumber exercises the notion that clients should refer to this field as Digits, not PhoneNumber.
	PhoneNumber string `json:"Digits"`
	// MarshalToString makes sure that we can use strings as an alternate JSON format for structs.
	MarshalToString MarshalToString
	// MarshalToString makes sure that we can use custom marshaling of struct values.
	// This is NOT globally supported in all client languages - just Go for now.
	MarshalToObject MarshalToObject
}

type SampleSecurityRequest struct {
	ID      string
	User    SampleUser
	FancyID StringLike
}

type StringLike string

type SampleSecurityResponse struct {
	Roles []string
}

// MarshalToString implements MarshalJSON/UnmarshalJSON to show that you can convert a struct
// type into some primitive like a string and have that work in your clients. Instead of using
// the standard object-based JSON this would normally marshal to, this uses a string
// formatted like "Home,Work".
//
// This SHOULD be supported by external clients like JS/Dart/etc.
type MarshalToString struct {
	// Home is supposed to be a home email address.
	Home string
	// Work is supposed to be a home email address.
	Work string
}

func (m *MarshalToString) UnmarshalJSON(jsonBytes []byte) error {
	jsonString := strings.Trim(string(jsonBytes), `"`)
	if jsonString == "" {
		return nil
	}

	switch emails := strings.Split(jsonString, ","); {
	case len(emails) >= 2:
		m.Home = emails[0]
		m.Work = emails[1]
	case len(emails) == 1:
		m.Home = emails[0]
	}
	return nil
}

func (m MarshalToString) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s,%s"`, m.escape(m.Home), m.escape(m.Work))), nil
}

func (MarshalToString) escape(value string) string {
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", "\\n")
	return value
}

// MarshalToObject is a struct that implements MarshalJSON/UnmarshalJSON in order to
// remap the structure of this from {Home:"", Work:""} to {H:"", W:""}. Ideally, you
// should just do this using struct attributes - it will work better.
//
// This is NOT supported in non-Go language clients because we have no way to convey
// to the request builder code the correct structure it should submit. I include this
// so that we can have a test codifying that this behavior is not supported. If you want
// different fields, use `json:""` tags.
type MarshalToObject struct {
	// Home is supposed to be a home email address.
	Home string
	// Work is supposed to be a home email address.
	Work string
}

func (m *MarshalToObject) UnmarshalJSON(jsonBytes []byte) error {
	out := map[string]string{}
	if err := json.Unmarshal(jsonBytes, &out); err != nil {
		return nil
	}
	m.Home = out["H"]
	m.Work = out["W"]
	return nil
}

func (m MarshalToObject) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"H":"%s", "W":"%s"}`, m.Home, m.Work)), nil
}

// CustomDuration is a standard Duration alias that uses duration strings for JSON
// transport as opposed to epoch nanos.
type CustomDuration time.Duration

func (duration *CustomDuration) UnmarshalJSON(jsonBytes []byte) error {
	durationString := strings.Trim(string(jsonBytes), `"`)
	d, err := time.ParseDuration(durationString)
	if err != nil {
		return fmt.Errorf("custom duration: %w", err)
	}
	*duration = CustomDuration(d)
	return nil
}

func (duration CustomDuration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Duration(duration).String() + `"`), nil
}

type SampleComplexRequest struct {
	InUser    SampleUser
	InFlag    bool
	InFloat   float64
	InTime    time.Time
	InTimePtr *time.Time
}

type SampleComplexResponse struct {
	OutFlag    bool
	OutFloat   float64
	OutUser    *SampleUser
	OutTime    time.Time
	OutTimePtr *time.Time
}

type SampleDownloadRequest struct {
	Format string
}

type SampleDownloadResponse struct {
	services.StreamResponse
}

type SampleRedirectRequest struct{}

type SampleRedirectResponse struct {
	URI string
	services.StreamResponse
}

func (res SampleRedirectResponse) Redirect() string {
	return res.URI
}

type FailAlwaysRequest struct {
	RequestValue string
}

type FailAlwaysResponse struct {
	ResponseValue string
}

// EventError captures the various ways you can bind the error message and its status codes
type EventError struct {
	Message        string
	Error          string
	Code           int
	Status         int
	StatusCode     int
	HTTPStatusCode int
}

type FailAlwaysErrorRequest struct {
	Error         EventError
	RequestValue  string
	ResponseValue string
	Text          string
}

type FailAlwaysErrorResponse struct {
}
