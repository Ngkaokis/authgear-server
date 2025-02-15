package webapp

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

var phoneRegexp = regexp.MustCompile(`^\+[0-9]*$`)

func GetIdentificationOptions(f *authflow.FlowResponse) []declarative.IdentificationOption {
	var options []declarative.IdentificationOption
	switch data := f.Action.Data.(type) {
	case declarative.IntentLoginFlowStepIdentifyData:
		options = data.Options
	case declarative.IntentSignupFlowStepIdentifyData:
		options = data.Options
	case declarative.IntentPromoteFlowStepIdentifyData:
		options = data.Options
	case declarative.IntentSignupLoginFlowStepIdentifyData:
		options = data.Options
	default:
		panic(fmt.Errorf("unexpected type of data: %T", f.Action.Data))
	}
	return options
}

func GetMostAppropriateIdentification(f *authflow.FlowResponse, loginID string, loginIDInputType string) config.AuthenticationFlowIdentification {
	// If loginIDInputType already tell us the login id type, return the corresponding type
	switch loginIDInputType {
	case "email":
		return config.AuthenticationFlowIdentificationEmail
	case "phone":
		return config.AuthenticationFlowIdentificationPhone
	}

	// Else, guess the type

	lookLikeAPhoneNumber := func(loginID string) bool {
		err := phone.EnsureE164(loginID)
		if err == nil {
			return true
		}

		if phoneRegexp.MatchString(loginID) {
			return true
		}

		return false
	}

	lookLikeAnEmailAddress := func(loginID string) bool {
		_, err := mail.ParseAddress(loginID)
		if err == nil {
			return true
		}

		if strings.Contains(loginID, "@") {
			return true
		}

		return false
	}

	isPhoneLike := lookLikeAPhoneNumber(loginID)
	isEmailLike := lookLikeAnEmailAddress(loginID)

	options := GetIdentificationOptions(f)
	var iden config.AuthenticationFlowIdentification
	for _, o := range options {
		switch o.Identification {
		case config.AuthenticationFlowIdentificationEmail:
			// If it is a email like login id, and there is an email option, it must be email
			if isEmailLike {
				iden = config.AuthenticationFlowIdentificationEmail
				break
			}
		case config.AuthenticationFlowIdentificationPhone:
			// If it is a phone like login id, and there is an phone option, it must be phone
			if isPhoneLike {
				iden = config.AuthenticationFlowIdentificationPhone
				break
			}
		case config.AuthenticationFlowIdentificationUsername:
			// The login id is not phone or email, then it can only be username
			if !isPhoneLike && !isEmailLike {
				iden = config.AuthenticationFlowIdentificationUsername
				break
			}
			// If it is like a email or phone, it can be username,
			// but we should continue the loop to see if there are better options
			if iden == "" {
				iden = config.AuthenticationFlowIdentificationUsername
			}
		}
	}

	if iden == "" {
		panic(fmt.Errorf("expected the authflow to allow login ID as identification"))
	}

	return iden
}
