package authflowv2

import (
	"net/http"
	"net/url"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowPromoteHTML = template.RegisterHTML(
	"web/authflowv2/signup.html",
	handlerwebapp.Components...,
)

var AuthflowPromoteLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"q_login_id_key": { "type": "string" },
			"q_login_id": { "type": "string" }
		},
		"required": ["q_login_id_key", "q_login_id"]
	}
`)

func ConfigureAuthflowV2PromoteRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RoutePromote)
}

type AuthflowV2PromoteEndpointsProvider interface {
	SSOCallbackURL(alias string) *url.URL
}

type AuthflowV2PromoteHandler struct {
	Controller        *handlerwebapp.AuthflowController
	BaseViewModel     *viewmodels.BaseViewModeler
	AuthflowViewModel *viewmodels.AuthflowViewModeler
	Renderer          handlerwebapp.Renderer
	Endpoints         AuthflowV2PromoteEndpointsProvider
}

func (h *AuthflowV2PromoteHandler) GetData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	authflowViewModel := h.AuthflowViewModel.NewWithAuthflow(screen.StateTokenFlowResponse, r)
	viewmodels.Embed(data, authflowViewModel)

	data["LoginDisabled"] = true

	return data, nil
}

func (h *AuthflowV2PromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flowName := "default"
	opts := webapp.SessionOptions{
		RedirectURI: h.Controller.RedirectURI(r),
	}

	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowPromoteHTML, data)
		return nil
	})

	handlers.PostAction("oauth", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		providerAlias := r.Form.Get("x_provider_alias")
		callbackURL := h.Endpoints.SSOCallbackURL(providerAlias).String()
		input := map[string]interface{}{
			"identification": "oauth",
			"alias":          providerAlias,
			"redirect_uri":   callbackURL,
			"response_mode":  string(sso.ResponseModeFormPost),
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	handlers.PostAction("login_id", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowPromoteLoginIDSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginIDKey := r.Form.Get("q_login_id_key")
		loginID := r.Form.Get("q_login_id")
		identification := loginIDKey
		input := map[string]interface{}{
			"identification": identification,
			"login_id":       loginID,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStartOfFlow(w, r, opts, authflow.FlowReference{
		Type: authflow.FlowTypePromote,
		Name: flowName,
	}, &handlers, nil)
}
