// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package portal

import (
	"github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/global"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/endpoint"
	"github.com/authgear/authgear-server/pkg/portal/graphql"
	"github.com/authgear/authgear-server/pkg/portal/loader"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/portal/task"
	"github.com/authgear/authgear-server/pkg/portal/task/tasks"
	"github.com/authgear/authgear-server/pkg/portal/transport"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/template"
	"net/http"
)

import (
	_ "github.com/authgear/authgear-server/pkg/auth"
)

// Injectors from wire.go:

func newPanicEndMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panicEndMiddleware := &middleware.PanicEndMiddleware{}
	return panicEndMiddleware
}

func newPanicLogMiddleware(p *deps.RequestProvider) httproute.Middleware {
	rootProvider := p.RootProvider
	factory := rootProvider.LoggerFactory
	panicLogMiddlewareLogger := middleware.NewPanicLogMiddlewareLogger(factory)
	panicLogMiddleware := &middleware.PanicLogMiddleware{
		Logger: panicLogMiddlewareLogger,
	}
	return panicLogMiddleware
}

func newPanicWriteEmptyResponseMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panicWriteEmptyResponseMiddleware := &middleware.PanicWriteEmptyResponseMiddleware{}
	return panicWriteEmptyResponseMiddleware
}

func newBodyLimitMiddleware(p *deps.RequestProvider) httproute.Middleware {
	bodyLimitMiddleware := &middleware.BodyLimitMiddleware{}
	return bodyLimitMiddleware
}

func newSentryMiddleware(p *deps.RequestProvider) httproute.Middleware {
	rootProvider := p.RootProvider
	hub := rootProvider.SentryHub
	environmentConfig := rootProvider.EnvironmentConfig
	trustProxy := environmentConfig.TrustProxy
	sentryMiddleware := &middleware.SentryMiddleware{
		SentryHub:  hub,
		TrustProxy: trustProxy,
	}
	return sentryMiddleware
}

func newSessionInfoMiddleware(p *deps.RequestProvider) httproute.Middleware {
	sessionInfoMiddleware := &session.SessionInfoMiddleware{}
	return sessionInfoMiddleware
}

func newSessionRequiredMiddleware(p *deps.RequestProvider) httproute.Middleware {
	sessionRequiredMiddleware := &session.SessionRequiredMiddleware{}
	return sessionRequiredMiddleware
}

func newGraphQLHandler(p *deps.RequestProvider) http.Handler {
	rootProvider := p.RootProvider
	environmentConfig := rootProvider.EnvironmentConfig
	devMode := environmentConfig.DevMode
	factory := rootProvider.LoggerFactory
	logger := graphql.NewLogger(factory)
	authgearConfig := rootProvider.AuthgearConfig
	adminAPIConfig := rootProvider.AdminAPIConfig
	controller := rootProvider.ConfigSourceController
	configSource := deps.ProvideConfigSource(controller)
	clock := _wireSystemClockValue
	adder := &authz.Adder{
		Clock: clock,
	}
	adminAPIService := &service.AdminAPIService{
		AuthgearConfig: authgearConfig,
		AdminAPIConfig: adminAPIConfig,
		ConfigSource:   configSource,
		AuthzAdder:     adder,
	}
	userLoader := loader.NewUserLoader(adminAPIService)
	appServiceLogger := service.NewAppServiceLogger(factory)
	databaseEnvironmentConfig := rootProvider.DatabaseConfig
	sqlBuilder := global.NewSQLBuilder(databaseEnvironmentConfig)
	request := p.Request
	context := deps.ProvideRequestContext(request)
	pool := rootProvider.Database
	handle := global.NewHandle(context, pool, factory)
	sqlExecutor := global.NewSQLExecutor(context, handle)
	appConfig := rootProvider.AppConfig
	secretKeyAllowlist := rootProvider.SecretKeyAllowlist
	configServiceLogger := service.NewConfigServiceLogger(factory)
	domainImplementationType := rootProvider.DomainImplementation
	kubernetesConfig := rootProvider.KubernetesConfig
	kubernetesLogger := service.NewKubernetesLogger(factory)
	kubernetes := &service.Kubernetes{
		KubernetesConfig: kubernetesConfig,
		AppConfig:        appConfig,
		Logger:           kubernetesLogger,
	}
	configService := &service.ConfigService{
		Context:              context,
		Logger:               configServiceLogger,
		AppConfig:            appConfig,
		Controller:           controller,
		ConfigSource:         configSource,
		DomainImplementation: domainImplementationType,
		Kubernetes:           kubernetes,
	}
	mailConfig := rootProvider.MailConfig
	inProcessExecutorLogger := task.NewInProcessExecutorLogger(factory)
	mailLogger := mail.NewLogger(factory)
	smtpConfig := rootProvider.SMTPConfig
	smtpServerCredentials := deps.ProvideSMTPServerCredentials(smtpConfig)
	dialer := mail.NewGomailDialer(smtpServerCredentials)
	sender := &mail.Sender{
		Logger:       mailLogger,
		DevMode:      devMode,
		GomailDialer: dialer,
	}
	sendMessagesLogger := tasks.NewSendMessagesLogger(factory)
	sendMessagesTask := &tasks.SendMessagesTask{
		EmailSender: sender,
		Logger:      sendMessagesLogger,
	}
	inProcessExecutor := task.NewExecutor(inProcessExecutorLogger, sendMessagesTask)
	inProcessQueue := &task.InProcessQueue{
		Executor: inProcessExecutor,
	}
	trustProxy := environmentConfig.TrustProxy
	requestOriginProvider := &endpoint.RequestOriginProvider{
		Request:    request,
		TrustProxy: trustProxy,
	}
	endpointsProvider := &endpoint.EndpointsProvider{
		OriginProvider: requestOriginProvider,
	}
	manager := rootProvider.Resources
	defaultLanguageTag := _wireDefaultLanguageTagValue
	supportedLanguageTags := _wireSupportedLanguageTagsValue
	resolver := &template.Resolver{
		Resources:             manager,
		DefaultLanguageTag:    defaultLanguageTag,
		SupportedLanguageTags: supportedLanguageTags,
	}
	engine := &template.Engine{
		Resolver: resolver,
	}
	collaboratorService := &service.CollaboratorService{
		Context:        context,
		Clock:          clock,
		SQLBuilder:     sqlBuilder,
		SQLExecutor:    sqlExecutor,
		MailConfig:     mailConfig,
		TaskQueue:      inProcessQueue,
		Endpoints:      endpointsProvider,
		TemplateEngine: engine,
		AdminAPI:       adminAPIService,
	}
	authzService := &service.AuthzService{
		Context:       context,
		Configs:       configService,
		Collaborators: collaboratorService,
	}
	domainService := &service.DomainService{
		Context:      context,
		Clock:        clock,
		DomainConfig: configService,
		SQLBuilder:   sqlBuilder,
		SQLExecutor:  sqlExecutor,
	}
	appBaseResources := deps.ProvideAppBaseResources(rootProvider)
	appService := &service.AppService{
		Logger:             appServiceLogger,
		SQLBuilder:         sqlBuilder,
		SQLExecutor:        sqlExecutor,
		AppConfig:          appConfig,
		SecretKeyAllowlist: secretKeyAllowlist,
		AppConfigs:         configService,
		AppAuthz:           authzService,
		AppAdminAPI:        adminAPIService,
		AppDomains:         domainService,
		Resources:          manager,
		AppBaseResources:   appBaseResources,
	}
	appLoader := loader.NewAppLoader(appService, authzService)
	domainLoader := loader.NewDomainLoader(domainService, authzService)
	collaboratorLoader := loader.NewCollaboratorLoader(collaboratorService, authzService)
	collaboratorInvitationLoader := loader.NewCollaboratorInvitationLoader(collaboratorService, authzService)
	graphqlContext := &graphql.Context{
		GQLLogger:               logger,
		Users:                   userLoader,
		Apps:                    appLoader,
		Domains:                 domainLoader,
		Collaborators:           collaboratorLoader,
		CollaboratorInvitations: collaboratorInvitationLoader,
		AuthzService:            authzService,
		AppService:              appService,
		DomainService:           domainService,
		CollaboratorService:     collaboratorService,
		SecretKeyAllowlist:      secretKeyAllowlist,
	}
	graphQLHandler := &transport.GraphQLHandler{
		DevMode:        devMode,
		GraphQLContext: graphqlContext,
		Database:       handle,
	}
	return graphQLHandler
}

var (
	_wireSystemClockValue           = clock.NewSystemClock()
	_wireDefaultLanguageTagValue    = template.DefaultLanguageTag(intl.DefaultLanguage)
	_wireSupportedLanguageTagsValue = template.SupportedLanguageTags([]string{intl.DefaultLanguage})
)

func newSystemConfigHandler(p *deps.RequestProvider) http.Handler {
	rootProvider := p.RootProvider
	authgearConfig := rootProvider.AuthgearConfig
	appConfig := rootProvider.AppConfig
	manager := rootProvider.Resources
	systemConfigProvider := &service.SystemConfigProvider{
		AuthgearConfig: authgearConfig,
		AppConfig:      appConfig,
		Resources:      manager,
	}
	systemConfigHandler := &transport.SystemConfigHandler{
		SystemConfig: systemConfigProvider,
	}
	return systemConfigHandler
}

func newAdminAPIHandler(p *deps.RequestProvider) http.Handler {
	request := p.Request
	context := deps.ProvideRequestContext(request)
	rootProvider := p.RootProvider
	pool := rootProvider.Database
	factory := rootProvider.LoggerFactory
	handle := global.NewHandle(context, pool, factory)
	configServiceLogger := service.NewConfigServiceLogger(factory)
	appConfig := rootProvider.AppConfig
	controller := rootProvider.ConfigSourceController
	configSource := deps.ProvideConfigSource(controller)
	domainImplementationType := rootProvider.DomainImplementation
	kubernetesConfig := rootProvider.KubernetesConfig
	kubernetesLogger := service.NewKubernetesLogger(factory)
	kubernetes := &service.Kubernetes{
		KubernetesConfig: kubernetesConfig,
		AppConfig:        appConfig,
		Logger:           kubernetesLogger,
	}
	configService := &service.ConfigService{
		Context:              context,
		Logger:               configServiceLogger,
		AppConfig:            appConfig,
		Controller:           controller,
		ConfigSource:         configSource,
		DomainImplementation: domainImplementationType,
		Kubernetes:           kubernetes,
	}
	clockClock := _wireSystemClockValue
	databaseEnvironmentConfig := rootProvider.DatabaseConfig
	sqlBuilder := global.NewSQLBuilder(databaseEnvironmentConfig)
	sqlExecutor := global.NewSQLExecutor(context, handle)
	mailConfig := rootProvider.MailConfig
	inProcessExecutorLogger := task.NewInProcessExecutorLogger(factory)
	logger := mail.NewLogger(factory)
	environmentConfig := rootProvider.EnvironmentConfig
	devMode := environmentConfig.DevMode
	smtpConfig := rootProvider.SMTPConfig
	smtpServerCredentials := deps.ProvideSMTPServerCredentials(smtpConfig)
	dialer := mail.NewGomailDialer(smtpServerCredentials)
	sender := &mail.Sender{
		Logger:       logger,
		DevMode:      devMode,
		GomailDialer: dialer,
	}
	sendMessagesLogger := tasks.NewSendMessagesLogger(factory)
	sendMessagesTask := &tasks.SendMessagesTask{
		EmailSender: sender,
		Logger:      sendMessagesLogger,
	}
	inProcessExecutor := task.NewExecutor(inProcessExecutorLogger, sendMessagesTask)
	inProcessQueue := &task.InProcessQueue{
		Executor: inProcessExecutor,
	}
	trustProxy := environmentConfig.TrustProxy
	requestOriginProvider := &endpoint.RequestOriginProvider{
		Request:    request,
		TrustProxy: trustProxy,
	}
	endpointsProvider := &endpoint.EndpointsProvider{
		OriginProvider: requestOriginProvider,
	}
	manager := rootProvider.Resources
	defaultLanguageTag := _wireDefaultLanguageTagValue
	supportedLanguageTags := _wireSupportedLanguageTagsValue
	resolver := &template.Resolver{
		Resources:             manager,
		DefaultLanguageTag:    defaultLanguageTag,
		SupportedLanguageTags: supportedLanguageTags,
	}
	engine := &template.Engine{
		Resolver: resolver,
	}
	authgearConfig := rootProvider.AuthgearConfig
	adminAPIConfig := rootProvider.AdminAPIConfig
	adder := &authz.Adder{
		Clock: clockClock,
	}
	adminAPIService := &service.AdminAPIService{
		AuthgearConfig: authgearConfig,
		AdminAPIConfig: adminAPIConfig,
		ConfigSource:   configSource,
		AuthzAdder:     adder,
	}
	collaboratorService := &service.CollaboratorService{
		Context:        context,
		Clock:          clockClock,
		SQLBuilder:     sqlBuilder,
		SQLExecutor:    sqlExecutor,
		MailConfig:     mailConfig,
		TaskQueue:      inProcessQueue,
		Endpoints:      endpointsProvider,
		TemplateEngine: engine,
		AdminAPI:       adminAPIService,
	}
	authzService := &service.AuthzService{
		Context:       context,
		Configs:       configService,
		Collaborators: collaboratorService,
	}
	adminAPILogger := transport.NewAdminAPILogger(factory)
	adminAPIHandler := &transport.AdminAPIHandler{
		Database: handle,
		Authz:    authzService,
		AdminAPI: adminAPIService,
		Logger:   adminAPILogger,
	}
	return adminAPIHandler
}

func newStaticAssetsHandler(p *deps.RequestProvider) http.Handler {
	rootProvider := p.RootProvider
	manager := rootProvider.Resources
	staticAssetsHandler := &transport.StaticAssetsHandler{
		Resources: manager,
	}
	return staticAssetsHandler
}
