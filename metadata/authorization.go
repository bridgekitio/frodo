package metadata

import "context"

type contextKeyAuthorization struct{}

// Authorization extracts the original caller's metadata authorization credentials.
func Authorization(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if auth, ok := ctx.Value(contextKeyAuthorization{}).(string); ok {
		return auth
	}
	return ""
}

// WithAuthorization stores the user's/caller's auth credentials on the request context. Typically,
// you will not need to call this yourself as the framework will take care of this for you. You just
// need to call metadata.Authorization() when you need to check that value.
func WithAuthorization(ctx context.Context, auth string) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, contextKeyAuthorization{}, auth)
}
