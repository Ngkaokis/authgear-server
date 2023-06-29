package nodes

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationWhatsapp{})
}

type InputAuthenticationWhatsapp interface {
	GetWhatsappOTP() string
}

type EdgeAuthenticationWhatsapp struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeAuthenticationWhatsapp) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationWhatsapp
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	phone := e.Authenticator.OOBOTP.Phone
	userID := e.Authenticator.UserID
	code := input.GetWhatsappOTP()
	err := ctx.OTPCodeService.VerifyOTP(
		otp.KindOOBOTP(ctx.Config, model.AuthenticatorOOBChannelWhatsapp),
		phone,
		code,
		&otp.VerifyOptions{
			UserID: userID,
		},
	)
	if err != nil {
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, errorutil.WithDetails(api.ErrInvalidCredentials, errorutil.Details{
				"AuthenticationType": apierrors.APIErrorDetail.Value(e.Authenticator.Type),
			})
		}
		return nil, err
	}

	return &NodeAuthenticationWhatsapp{Stage: e.Stage, Authenticator: e.Authenticator}, nil
}

type NodeAuthenticationWhatsapp struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeAuthenticationWhatsapp) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationWhatsapp) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationWhatsapp) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			AuthenticationType:    authn.AuthenticationTypeOOBOTPSMS,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}
