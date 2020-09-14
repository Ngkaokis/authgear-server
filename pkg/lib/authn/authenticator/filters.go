package authenticator

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Filter interface {
	Keep(ai *Info) bool
}

type FilterFunc func(ai *Info) bool

func (f FilterFunc) Keep(ai *Info) bool {
	return f(ai)
}

var KeepDefault FilterFunc = func(ai *Info) bool {
	return ai.IsDefault
}

func KeepKind(kind Kind) Filter {
	return FilterFunc(func(ai *Info) bool {
		return ai.Kind == kind
	})
}

func KeepType(typ authn.AuthenticatorType) Filter {
	return FilterFunc(func(ai *Info) bool {
		return ai.Type == typ
	})
}

func KeepPrimaryAuthenticatorOfIdentity(ii *identity.Info) Filter {
	return FilterFunc(func(ai *Info) bool {
		types := ii.Type.PrimaryAuthenticatorTypes()

		for _, typ := range types {
			if ai.Type == typ {
				switch {
				case ii.Type == authn.IdentityTypeLoginID && ai.Type == authn.AuthenticatorTypeOOB:
					loginID := ii.Claims[identity.IdentityClaimLoginIDValue]
					email, _ := ai.Claims[AuthenticatorClaimOOBOTPEmail].(string)
					phone, _ := ai.Claims[AuthenticatorClaimOOBOTPPhone].(string)
					if loginID == email || loginID == phone {
						return true
					}
				default:
					return true
				}
			}
		}

		return false
	})
}

func KeepMatchingAuthenticatorOfIdentity(ii *identity.Info) Filter {
	return FilterFunc(func(ai *Info) bool {
		types := ii.Type.MatchingAuthenticatorTypes()

		for _, typ := range types {
			if ai.Type == typ {
				switch {
				case ii.Type == authn.IdentityTypeLoginID && ai.Type == authn.AuthenticatorTypeOOB:
					loginIDType, _ := ii.Claims[identity.IdentityClaimLoginIDType].(string)
					loginID, _ := ii.Claims[identity.IdentityClaimLoginIDValue].(string)
					email, _ := ai.Claims[AuthenticatorClaimOOBOTPEmail].(string)
					phone, _ := ai.Claims[AuthenticatorClaimOOBOTPPhone].(string)
					switch loginIDType {
					case string(config.LoginIDKeyTypeEmail):
						return loginID == email
					case string(config.LoginIDKeyTypePhone):
						return loginID == phone
					}
					return false
				default:
					return true
				}
			}
		}

		return false
	})
}
