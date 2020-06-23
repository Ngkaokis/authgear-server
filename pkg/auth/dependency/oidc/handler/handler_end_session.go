package handler

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc/protocol"
	coreurl "github.com/skygeario/skygear-server/pkg/core/url"
)

// TODO(oidc): write tests

type WebAppURLsProvider interface {
	LogoutURL(redirectURI *url.URL) *url.URL
	SettingsURL() *url.URL
}

type EndSessionHandler struct {
	Config    *config.OAuthConfig
	Endpoints oidc.EndpointsProvider
	URLs      WebAppURLsProvider
}

func (h *EndSessionHandler) Handle(s auth.AuthSession, req protocol.EndSessionRequest, r *http.Request, rw http.ResponseWriter) error {
	if s != nil {
		endSessionURL := coreurl.WithQueryParamsAdded(
			h.Endpoints.EndSessionEndpointURL(),
			req,
		)
		logoutURL := h.URLs.LogoutURL(endSessionURL)

		http.Redirect(rw, r, logoutURL.String(), http.StatusFound)
		return nil
	}

	redirectURI := req.PostLogoutRedirectURI()
	valid, client := h.validateRedirectURI(redirectURI)
	if !valid {
		// Invalid/empty redirect URI, redirect to home page/settings
		if client != nil && client.ClientURI() != "" {
			redirectURI = client.ClientURI()
		} else {
			redirectURI = h.URLs.SettingsURL().String()
		}
		http.Redirect(rw, r, redirectURI, http.StatusFound)
		return nil
	}

	if state := req.State(); state != "" {
		uri, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}
		redirectURI = coreurl.WithQueryParamsAdded(uri, map[string]string{"state": state}).String()
	}

	http.Redirect(rw, r, redirectURI, http.StatusFound)
	return nil
}

func (h *EndSessionHandler) validateRedirectURI(redirectURI string) (valid bool, client config.OAuthClientConfig) {
	for _, client := range h.Config.Clients {
		for _, uri := range client.PostLogoutRedirectURIs() {
			if uri == redirectURI {
				return true, client
			}
		}
	}
	return false, nil
}
