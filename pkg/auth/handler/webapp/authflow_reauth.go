package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuthflowReauthRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(webapp.AuthflowRouteReauth)
}

type AuthflowReauthHandler struct {
	Controller *AuthflowController

	AuthflowNavigator AuthflowNavigator
}

func (h *AuthflowReauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flowName := "default"
	opts := webapp.SessionOptions{
		RedirectURI: h.Controller.RedirectURI(r),
	}

	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		// HandleStartOfFlow used to redirect to the next screen for us.
		// But that redirect was removed.
		// So we need to redirect here.
		// See https://github.com/authgear/authgear-server/issues/3470
		result := &webapp.Result{}
		screen.Navigate(h.AuthflowNavigator, r, s.ID, result)
		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStartOfFlow(w, r, opts, authflow.FlowReference{
		Type: authflow.FlowTypeReauth,
		Name: flowName,
	}, &handlers, nil)
}
