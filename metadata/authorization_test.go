//go:build unit

package metadata_test

import (
	"context"
	"testing"

	"github.com/bridgekit-io/frodo/metadata"
	"github.com/stretchr/testify/suite"
)

func TestAuthorizationSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationSuite))
}

type AuthorizationSuite struct {
	suite.Suite
}

func (suite *AuthorizationSuite) TestDefaults() {
	suite.Equal("", metadata.Authorization(nil))
	suite.Equal("", metadata.Authorization(context.Background()))
	suite.Nil(metadata.WithAuthorization(nil, ""))
}

func (suite *AuthorizationSuite) TestWithAuthorization() {
	ctx := context.Background()

	ctx = metadata.WithAuthorization(ctx, "Abide")
	suite.Equal("Abide", metadata.Authorization(ctx))

	ctx = metadata.WithAuthorization(ctx, "Abide Harder")
	suite.Equal("Abide Harder", metadata.Authorization(ctx))
}
