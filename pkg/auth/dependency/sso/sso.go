package sso

// Scope parameter allows the application to express the desired scope of the access request.
type Scope []string

// Options parameter allows additional options for getting auth url
type Options map[string]interface{}

// UXMode indicates how the URL is used
type UXMode int

const (
	// Undefined for undefined uxmode
	Undefined UXMode = iota
	// WebRedirect for web url redirect
	WebRedirect
	// WebPopup for web popup window
	WebPopup
	// IOS for device iOS
	IOS
	// Android for device Android
	Android
)

func (m UXMode) String() string {
	names := [...]string{
		"web_redirect",
		"web_popup",
		"ios",
		"android",
	}

	if m < WebRedirect || m > Android {
		return "undefined"
	}

	return names[m-1]
}

// UXModeFromString converts string to UXMode
func UXModeFromString(input string) (u UXMode) {
	UXModes := [...]UXMode{WebRedirect, WebPopup, IOS, Android}
	for _, v := range UXModes {
		if input == v.String() {
			u = v
			return
		}
	}

	return
}

// GetURLParams structs parameters for GetLoginAuthURL
type GetURLParams struct {
	Scope       Scope
	Options     Options
	CallbackURL string
	UXMode      UXMode
	UserID      string
	Action      string
}

// Setting is the base settings for SSO
type Setting struct {
	URLPrefix            string
	JSSDKCDNURL          string
	StateJWTSecret       string
	AutoLinkProviderKeys []string
	AllowedCallbackURLs  []string
}

// Config is the base config of a SSO provider
type Config struct {
	Name         string
	Enabled      bool
	ClientID     string
	ClientSecret string
	Scope        Scope
}

// Provider defines SSO interface
type Provider interface {
	GetAuthURL(params GetURLParams) (url string, err error)
}

// NewProvider is the provider factory
func NewProvider(
	setting Setting,
	config Config,
) Provider {
	if !config.Enabled {
		return nil
	}
	switch config.Name {
	case "google":
		return &GoogleImpl{
			Setting: setting,
			Config:  config,
		}
	case "facebook":
		return &FacebookImpl{
			Setting: setting,
			Config:  config,
		}
	case "instagram":
		return &InstagramImpl{
			Setting: setting,
			Config:  config,
		}
	case "linkedin":
		return &LinkedInImpl{
			Setting: setting,
			Config:  config,
		}
	}
	return nil
}
