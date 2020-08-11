package user

import (
	gotime "time"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
)

type HookProvider interface {
	DispatchEvent(payload event.Payload) error
}

type Commands struct {
	Raw          *RawCommands
	Hooks        HookProvider
	Verification VerificationService
}

func (c *Commands) Create(
	userID string,
	metadata map[string]interface{},
) (*User, error) {
	return c.Raw.Create(userID, metadata)
}

func (c *Commands) AfterCreate(
	user *User,
	identities []*identity.Info,
	authenticators []*authenticator.Info,
) error {
	isVerified := c.Verification.IsVerified(identities, authenticators)
	userModel := newUserModel(user, identities, isVerified)
	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}
	err := c.Hooks.DispatchEvent(&event.UserCreateEvent{
		User:       *userModel,
		Identities: identityModels,
	})
	if err != nil {
		return err
	}

	err = c.Raw.AfterCreate(userModel, identities)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) UpdateMetadata(user *model.User, metadata map[string]interface{}) error {
	// TODO(webhook): dispatch update metadata event
	return c.Raw.UpdateMetadata(user, metadata)
}

func (c *Commands) UpdateLoginTime(user *model.User, loginAt gotime.Time) error {
	return c.Raw.UpdateLoginTime(user, loginAt)
}
