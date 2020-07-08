package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	TemplateItemTypeAuthUISettingsHTML config.TemplateItemType = "auth_ui_settings.html"
)

var TemplateAuthUISettingsHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISettingsHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="settings-form primary-txt">
  You are authenticated. To logout, please visit <a href="/logout">here</a>.
</div>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

func ConfigureSettingsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/settings")
}

type SettingsHandler struct {
	// RenderProvider webapp.RenderProvider
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// FIXME(webapp): settings
	// h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUISettingsHTML, nil)
}
