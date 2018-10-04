package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// AttachLoginHandler attach login handler to server
func AttachLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/login", &LoginHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

// LoginHandlerFactory creates new handler
type LoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginHandlerFactory) NewHandler(request *http.Request) handler.Handler {
	h := &LoginHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h)
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	AuthData         map[string]interface{} `json:"auth_data"`
	Password         string                 `json:"password"`
}

// Validate request payload
func (p LoginRequestPayload) Validate() error {
	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
}

// LoginHandler handles login request
type LoginHandler struct{
	AuthDataChecker      dependency.AuthDataChecker `dependency:"AuthDataChecker"`
	TokenStore           authtoken.Store            `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
}

// ProvideAuthzPolicy provides authorization policy
func (h LoginHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// DecodeRequest decode request payload
func (h LoginHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := LoginRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

// Handle api request
func (h LoginHandler) Handle(req interface{}, ctx handler.AuthContext) (resp interface{}, err error) {
	payload := req.(LoginRequestPayload)

	if valid := h.AuthDataChecker.IsValid(payload.AuthData); !valid {
		err = skyerr.NewInvalidArgument("invalid auth data", []string{"auth_data"})
		return
	}

	principal := password.Principal{}
	err = h.PasswordAuthProvider.GetPrincipal(payload.AuthData, &principal)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return 
		}
		// TODO: more error handling here if necessary
		err = skyerr.NewResourceFetchFailureErr("auth_data", payload.AuthData)
		return
	}

	if !principal.IsSamePassword(payload.Password) {
		err = skyerr.NewError(skyerr.InvalidCredentials, "auth_data or password incorrect")
		return
	}

	fetchedAuthInfo := authinfo.AuthInfo{}
	if err = h.AuthInfoStore.GetAuth(principal.UserID, &fetchedAuthInfo); err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		// TODO: more error handling here if necessary
		err = skyerr.NewResourceFetchFailureErr("auth_data", payload.AuthData)
		return
	}

	if err = checkUserIsNotDisabled(&fetchedAuthInfo); err != nil {
		return
	}

	// generate access-token
	token, err := h.TokenStore.NewToken(fetchedAuthInfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	authContext := handler.AuthContext{}
	authContext.AuthInfo = &fetchedAuthInfo

	resp = response.NewAuthResponse(authContext, skydb.Record{}, token.AccessToken)

	// Populate the activity time to user
	now := timeNow()
	authContext.AuthInfo.LastLoginAt = &now
	authContext.AuthInfo.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(authContext.AuthInfo); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	// TODO: Audit

	return
}
