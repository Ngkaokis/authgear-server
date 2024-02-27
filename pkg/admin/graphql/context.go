package graphql

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api/event"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserLoader interface {
	graphqlutil.DataLoaderInterface
}

type IdentityLoader interface {
	graphqlutil.DataLoaderInterface
}

type AuthenticatorLoader interface {
	graphqlutil.DataLoaderInterface
}

type RoleLoader interface {
	graphqlutil.DataLoaderInterface
}

type GroupLoader interface {
	graphqlutil.DataLoaderInterface
}

type AuditLogLoader interface {
	graphqlutil.DataLoaderInterface
}

type AuditLogFacade interface {
	QueryPage(opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
}

type UserFacade interface {
	ListPage(sortOption libuser.SortOption, args graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	SearchPage(searchKeyword string, sortOption libuser.SortOption, args graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	Create(identityDef model.IdentityDef, password string) (string, error)
	ResetPassword(id string, password string) error
	SetDisabled(id string, isDisabled bool, reason *string) error
	ScheduleDeletion(id string) error
	UnscheduleDeletion(id string) error
	Delete(id string) error
	ScheduleAnonymization(id string) error
	UnscheduleAnonymization(id string) error
	Anonymize(id string) error
}

type RolesGroupsFacade interface {
	CreateRole(options *rolesgroups.NewRoleOptions) (string, error)
	UpdateRole(options *rolesgroups.UpdateRoleOptions) error
	DeleteRole(id string) error
	ListGroupsByRoleID(roleID string) ([]*apimodel.Group, error)
	ListRoles(options *rolesgroups.ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)

	CreateGroup(options *rolesgroups.NewGroupOptions) (string, error)
	UpdateGroup(options *rolesgroups.UpdateGroupOptions) error
	DeleteGroup(id string) error
	ListRolesByGroupID(groupID string) ([]*apimodel.Role, error)
	ListGroups(options *rolesgroups.ListGroupsOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)

	AddRoleToGroups(options *rolesgroups.AddRoleToGroupsOptions) (string, error)
	RemoveRoleFromGroups(options *rolesgroups.RemoveRoleFromGroupsOptions) (string, error)

	AddRoleToUsers(options *rolesgroups.AddRoleToUsersOptions) (string, error)
	RemoveRoleFromUsers(options *rolesgroups.RemoveRoleFromUsersOptions) (string, error)
}

type IdentityFacade interface {
	Get(id string) (*identity.Info, error)
	List(userID string, identityType *apimodel.IdentityType) ([]*apimodel.IdentityRef, error)
	Remove(identityInfo *identity.Info) error
	Create(userID string, identityDef model.IdentityDef, password string) (*apimodel.IdentityRef, error)
	Update(identityID string, userID string, identityDef model.IdentityDef) (*apimodel.IdentityRef, error)
}

type AuthenticatorFacade interface {
	Get(id string) (*authenticator.Info, error)
	List(userID string, authenticatorType *apimodel.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error)
	Remove(authenticatorInfo *authenticator.Info) error
}

type VerificationFacade interface {
	Get(userID string) ([]model.Claim, error)
	SetVerified(userID string, claimName string, claimValue string, isVerified bool) error
}

type UserProfileFacade interface {
	DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
	ReadCustomAttributesInStorageForm(role accesscontrol.Role, userID string, storageForm map[string]interface{}) (map[string]interface{}, error)
	UpdateUserProfile(
		role accesscontrol.Role,
		userID string,
		stdAttrs map[string]interface{},
		customAttrs map[string]interface{},
	) error
}

type SessionFacade interface {
	List(userID string) ([]session.Session, error)
	Get(id string) (session.Session, error)
	Revoke(id string) error
	RevokeAll(userID string) error
}

type AuthorizationFacade interface {
	Get(id string) (*oauth.Authorization, error)
	List(userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error)
	Delete(a *oauth.Authorization) error
}

type OAuthFacade interface {
	CreateSession(clientID string, userID string) (session.Session, protocol.TokenResponse, error)
}

type SessionListingService interface {
	FilterForDisplay(sessions []session.Session, currentSession session.Session) ([]*sessionlisting.Session, error)
}

type OTPCodeService interface {
	GenerateOTP(kind otp.Kind, target string, form otp.Form, opts *otp.GenerateOptions) (string, error)
}

type ForgotPasswordService interface {
	SendCode(loginID string, options *forgotpassword.CodeOptions) error
}

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("admin-graphql")} }

type Context struct {
	GQLLogger Logger

	Config                *config.AppConfig
	OAuthConfig           *config.OAuthConfig
	AdminAPIFeatureConfig *config.AdminAPIFeatureConfig

	Users          UserLoader
	Identities     IdentityLoader
	Authenticators AuthenticatorLoader
	Roles          RoleLoader
	Groups         GroupLoader
	AuditLogs      AuditLogLoader

	UserFacade          UserFacade
	RolesGroupsFacade   RolesGroupsFacade
	AuditLogFacade      AuditLogFacade
	IdentityFacade      IdentityFacade
	AuthenticatorFacade AuthenticatorFacade
	VerificationFacade  VerificationFacade
	SessionFacade       SessionFacade
	UserProfileFacade   UserProfileFacade
	AuthorizationFacade AuthorizationFacade
	OAuthFacade         OAuthFacade
	SessionListing      SessionListingService
	OTPCode             OTPCodeService
	ForgotPassword      ForgotPasswordService
	Events              EventService
}

func (c *Context) Logger() *log.Logger {
	return c.GQLLogger.Logger
}

func WithContext(ctx context.Context, gqlContext *Context) context.Context {
	return graphqlutil.WithContext(ctx, gqlContext)
}

func GQLContext(ctx context.Context) *Context {
	return graphqlutil.GQLContext(ctx).(*Context)
}
