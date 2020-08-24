package auth

import (
	"github.com/google/wire"

	handleroauth "github.com/authgear/authgear-server/pkg/auth/handler/oauth"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	viewmodelswebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	identityanonymous "github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	oidchandler "github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ProvideOAuthMetadataProviders(oauth *oauth.MetadataProvider, oidc *oidc.MetadataProvider) []handleroauth.MetadataProvider {
	return []handleroauth.MetadataProvider{oauth, oidc}
}

var DependencySet = wire.NewSet(
	deps.RequestDependencySet,
	deps.CommonDependencySet,

	wire.Bind(new(webapp.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
	wire.Bind(new(webapp.GraphService), new(*interaction.Service)),
	wire.Bind(new(webapp.CookieFactory), new(*httputil.CookieFactory)),

	wire.NewSet(
		wire.Struct(new(MainOriginProvider), "*"),
		wire.Bind(new(OriginProvider), new(*MainOriginProvider)),
		wire.Struct(new(EndpointsProvider), "*"),

		wire.Bind(new(oauth.EndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(webapp.EndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(handlerwebapp.SetupTOTPEndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(oidc.EndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(sso.EndpointsProvider), new(*EndpointsProvider)),
		wire.Bind(new(otp.EndpointsProvider), new(*EndpointsProvider)),
	),

	webapp.DependencySet,
	wire.Bind(new(oauthhandler.WebAppAuthenticateURLProvider), new(*webapp.AuthenticateURLProvider)),
	wire.Bind(new(oidchandler.WebAppURLsProvider), new(*webapp.URLProvider)),
	wire.Bind(new(sso.RedirectURLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(forgotpassword.URLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(verification.WebAppURLProvider), new(*webapp.URLProvider)),

	middleware.DependencySet,

	handleroauth.DependencySet,
	wire.Bind(new(handleroauth.ProtocolAuthorizeHandler), new(*oauthhandler.AuthorizationHandler)),
	wire.Bind(new(handleroauth.ProtocolTokenHandler), new(*oauthhandler.TokenHandler)),
	wire.Bind(new(handleroauth.ProtocolRevokeHandler), new(*oauthhandler.RevokeHandler)),
	wire.Bind(new(handleroauth.ProtocolEndSessionHandler), new(*oidchandler.EndSessionHandler)),
	wire.Bind(new(handleroauth.ProtocolUserInfoProvider), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.JWSSource), new(*oidc.IDTokenIssuer)),
	wire.Bind(new(handleroauth.ChallengeProvider), new(*challenge.Provider)),
	ProvideOAuthMetadataProviders,

	viewmodelswebapp.DependencySet,

	handlerwebapp.DependencySet,
	wire.Bind(new(handlerwebapp.SettingsIdentityService), new(*identityservice.Service)),
	wire.Bind(new(handlerwebapp.SettingsVerificationService), new(*verification.Service)),
	wire.Bind(new(handlerwebapp.PasswordPolicy), new(*password.Checker)),
	wire.Bind(new(handlerwebapp.LogoutSessionManager), new(*session.Manager)),
	wire.Bind(new(handlerwebapp.WebAppService), new(*webapp.Service)),
)
