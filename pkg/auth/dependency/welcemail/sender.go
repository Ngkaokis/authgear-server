package welcemail

import (
	"errors"

	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type Sender interface {
	Send(email string, user model.User) error
}

type DefaultSender struct {
	AppName        string
	Config         config.WelcomeEmailConfiguration
	Dialer         *gomail.Dialer
	TemplateEngine *template.Engine
}

func NewDefaultSender(
	config config.TenantConfiguration,
	dialer *gomail.Dialer,
	templateEngine *template.Engine,
) Sender {
	return &DefaultSender{
		AppName:        config.AppName,
		Config:         config.UserConfig.WelcomeEmail,
		Dialer:         dialer,
		TemplateEngine: templateEngine,
	}
}

func (d *DefaultSender) Send(email string, user model.User) (err error) {
	if d.Config.TextURL == "" {
		return errors.New("welcome email text template url is empty")
	}

	context := map[string]interface{}{
		"appname":    d.AppName,
		"email":      email,
		"user":       user,
		"url_prefix": d.Config.URLPrefix,
	}

	var textBody string
	if textBody, err = d.TemplateEngine.ParseTextTemplate(
		authTemplate.TemplateNameWelcomeEmailText,
		context,
		template.ParseOption{Required: true},
	); err != nil {
		return
	}

	var htmlBody string
	if htmlBody, err = d.TemplateEngine.ParseHTMLTemplate(
		authTemplate.TemplateNameWelcomeEmailHTML,
		context,
		template.ParseOption{Required: false},
	); err != nil {
		return
	}

	sendReq := mail.SendRequest{
		Dialer:    d.Dialer,
		Sender:    d.Config.Sender,
		Recipient: email,
		Subject:   d.Config.Subject,
		ReplyTo:   d.Config.ReplyTo,
		TextBody:  textBody,
		HTMLBody:  htmlBody,
	}

	err = sendReq.Execute()
	return
}
