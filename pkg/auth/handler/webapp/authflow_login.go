package webapp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureAuthflowLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/authflow/login")
}

type AuthflowLoginHandler struct {
	Controller              *AuthflowController
	BaseViewModel           *viewmodels.BaseViewModeler
	AuthenticationViewModel *viewmodels.AuthenticationViewModeler
	FormPrefiller           *FormPrefiller
	Renderer                Renderer
	MeterService            MeterService
	TutorialCookie          TutorialCookie
	ErrorCookie             ErrorCookie
}

func (h *AuthflowLoginHandler) GetData(w http.ResponseWriter, r *http.Request, f *authflow.FlowResponse, allowLoginOnly bool) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	if h.TutorialCookie.Pop(r, w, httputil.SignupLoginTutorialCookieName) {
		baseViewModel.SetTutorial(httputil.SignupLoginTutorialCookieName)
	}
	viewmodels.Embed(data, baseViewModel)
	authenticationViewModel := h.AuthenticationViewModel.NewWithAuthflow(f, r.Form)
	viewmodels.Embed(data, authenticationViewModel)
	viewmodels.Embed(data, NewLoginViewModel(allowLoginOnly, r))
	return data, nil
}

func (h *AuthflowLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.FormPrefiller.Prefill(r.Form)

	var handlers AuthflowControllerHandlers
	defer h.Controller.MakeHTTPHandler(&handlers).ServeHTTP(w, r)

	opts := webapp.SessionOptions{
		RedirectURI: h.Controller.RedirectURI(r),
	}
	s, err := h.Controller.GetOrCreateWebSession(w, r, opts)
	if err != nil {
		// FIXME(authflow): log the error.
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	flowName := "default"
	flowReference := authflow.FlowReference{
		Type: authflow.FlowTypeLogin,
		Name: flowName,
	}
	checkFn := func(f *authflow.FlowResponse) bool {
		return f.Type == authflow.FlowTypeLogin && f.Name == flowName && f.Action.Type == authflow.FlowActionType(config.AuthenticationFlowStepTypeIdentify)
	}
	f, err := h.Controller.GetOrCreateAuthflow(r, s, flowReference, checkFn)
	if err != nil {
		// FIXME(authflow): log the error.
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	oauthProviderAlias := s.OAuthProviderAlias
	allowLoginOnly := s.UserIDHint != ""

	//oauthPostAction := func(providerAlias string) error {
	//	result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
	//		input = &InputUseOAuth{
	//			ProviderAlias:    providerAlias,
	//			ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
	//		}
	//		return
	//	})
	//	if err != nil {
	//		return err
	//	}

	//	result.WriteResponse(w, r)
	//	return nil
	//}

	handlers.Get(func() error {
		visitorID := webapp.GetVisitorID(r.Context())
		if visitorID == "" {
			// visitor id should be generated by VisitorIDMiddleware
			return fmt.Errorf("webapp: missing visitor id")
		}

		err := h.MeterService.TrackPageView(visitorID, meter.PageTypeLogin)
		if err != nil {
			return err
		}

		// FIXME(authflow): support oauthProviderAlias
		_ = oauthProviderAlias
		//_, hasErr := h.ErrorCookie.GetError(r)
		// If x_oauth_provider_alias is provided via authz endpoint
		// redirect the user to the oauth provider
		// If there is error in the ErrorCookie, the user will stay in the login
		// page to see the error message and the redirection won't be performed
		//if !hasErr && oauthProviderAlias != "" {
		//	return oauthPostAction(oauthProviderAlias)
		//}

		data, err := h.GetData(w, r, f, allowLoginOnly)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebLoginHTML, data)
		return nil
	})

	//handlers.PostAction("oauth", func() error {
	//	providerAlias := r.Form.Get("x_provider_alias")
	//	return oauthPostAction(providerAlias)
	//})

	handlers.PostAction("login_id", func() error {
		err = LoginWithLoginIDSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginID := r.Form.Get("q_login_id")
		identification := webapp.GetMostAppropriateIdentification(f, loginID)
		input := map[string]interface{}{
			"identification": identification,
			"login_id":       loginID,
		}

		result, err := h.Controller.FeedInput(r, s, f, input)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	//handlers.PostAction("passkey", func() error {
	//	result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
	//		err = PasskeyAutofillSchema.Validator().ValidateValue(FormToJSON(r.Form))
	//		if err != nil {
	//			return
	//		}

	//		assertionResponseStr := r.Form.Get("x_assertion_response")
	//		assertionResponse := []byte(assertionResponseStr)
	//		stage := string(authn.AuthenticationStagePrimary)

	//		input = &InputPasskeyAssertionResponse{
	//			Stage:             stage,
	//			AssertionResponse: assertionResponse,
	//		}
	//		return
	//	})
	//	if err != nil {
	//		return err
	//	}

	//	result.WriteResponse(w, r)
	//	return nil
	//})
}
