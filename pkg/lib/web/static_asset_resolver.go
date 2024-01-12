package web

import (
	"context"

	// nolint:gosec
	"crypto/md5"
	"fmt"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

const AppAssetsURLDirname = "static"
const GeneratedAssetsURLDirname = "shared-assets"

var StaticAssetResources = map[string]resource.Descriptor{
	"app-logo":      AppLogo,
	"app-logo-dark": AppLogoDark,
	"favicon":       Favicon,

	"authgear-light-theme.css":            AuthgearLightThemeCSS,
	"authgear-dark-theme.css":             AuthgearDarkThemeCSS,
	"authgear-authflowv2-light-theme.css": AuthgearAuthflowV2LightThemeCSS,
	"authgear-authflowv2-dark-theme.css":  AuthgearAuthflowV2DarkThemeCSS,
}

type ResourceManager interface {
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
}

type EmbeddedResourceManager interface {
	AssetName(key string) (name string, err error)
}

type StaticAssetResolver struct {
	Context           context.Context
	Localization      *config.LocalizationConfig
	HTTPOrigin        httputil.HTTPOrigin
	HTTPProto         httputil.HTTPProto
	WebAppCDNHost     config.WebAppCDNHost
	Resources         ResourceManager
	EmbeddedResources EmbeddedResourceManager
}

func (r *StaticAssetResolver) HasAppSpecificAsset(id string) bool {
	desc, ok := StaticAssetResources[id]
	if !ok {
		return false
	}

	css, ok := desc.(CSSDescriptor)
	if !ok {
		return false
	}

	_, err := r.Resources.Read(desc, resource.AppFile{
		Path: css.Path,
	})

	return err == nil
}

func (r *StaticAssetResolver) StaticAssetURL(id string) (string, error) {
	desc, ok := StaticAssetResources[id]
	if !ok {
		return "", fmt.Errorf("unknown static asset: %s", id)
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(r.Context)
	result, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: r.Localization.SupportedLanguages,
		DefaultTag:    *r.Localization.FallbackLanguage,
		PreferredTags: preferredLanguageTags,
	})
	if err != nil {
		return "", err
	}

	asset := result.(*StaticAsset)

	// md5 is used to compute the hash in the filename for caching purpose only
	// nolint:gosec
	hash := md5.Sum(asset.Data)

	hashPath := filepathutil.MakeHashedPath(asset.Path, fmt.Sprintf("%x", hash))
	return staticAssetURL(string(r.HTTPOrigin), "", hashPath)
}

func (r *StaticAssetResolver) GeneratedStaticAssetURL(key string) (string, error) {
	name, err := r.EmbeddedResources.AssetName(key)
	if err != nil {
		return "", err
	}

	origin := string(r.HTTPOrigin)
	if r.WebAppCDNHost != "" {
		origin = fmt.Sprintf("%s://%s", r.HTTPProto, r.WebAppCDNHost)
	}

	return staticAssetURL(origin, GeneratedAssetsURLDirname, name)
}

func staticAssetURL(origin string, prefix string, assetPath string) (string, error) {
	o, err := url.Parse(origin)
	if err != nil {
		return "", err
	}
	u, err := o.Parse(prefix)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, assetPath)
	return u.String(), nil
}

func LookLikeAHash(s string) bool {
	// hash that generated by parcel should be in length of 8
	return len(s) == 8
}
