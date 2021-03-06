// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/drone/drone/cmd/drone-server/config"
	"github.com/drone/drone/handler/api"
	"github.com/drone/drone/handler/web"
	"github.com/drone/drone/livelog"
	"github.com/drone/drone/metric"
	"github.com/drone/drone/operator/manager"
	"github.com/drone/drone/pubsub"
	"github.com/drone/drone/service/commit"
	"github.com/drone/drone/service/hook/parser"
	"github.com/drone/drone/service/license"
	"github.com/drone/drone/service/org"
	"github.com/drone/drone/service/repo"
	"github.com/drone/drone/service/token"
	"github.com/drone/drone/service/user"
	"github.com/drone/drone/store/batch"
	"github.com/drone/drone/store/cron"
	"github.com/drone/drone/store/perm"
	"github.com/drone/drone/store/secret"
	"github.com/drone/drone/store/step"
	"github.com/drone/drone/trigger"
	cron2 "github.com/drone/drone/trigger/cron"
)

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Injectors from wire.go:

func InitializeApplication(config2 config.Config) (application, error) {
	client := provideClient(config2)
	refresher := provideRefresher(config2)
	db, err := provideDatabase(config2)
	if err != nil {
		return application{}, err
	}
	userStore := provideUserStore(db)
	renewer := token.Renewer(refresher, userStore)
	commitService := commit.New(client, renewer)
	cronStore := cron.New(db)
	repositoryStore := provideRepoStore(db)
	fileService := provideContentService(client, renewer)
	configService := provideConfigPlugin(client, fileService, config2)
	statusService := provideStatusService(client, renewer, config2)
	buildStore := provideBuildStore(db)
	stageStore := provideStageStore(db)
	scheduler := provideScheduler(stageStore, config2)
	webhookSender := provideWebhookPlugin(config2)
	triggerer := trigger.New(configService, commitService, statusService, buildStore, scheduler, repositoryStore, userStore, webhookSender)
	cronScheduler := cron2.New(commitService, cronStore, repositoryStore, userStore, triggerer)
	system := provideSystem(config2)
	coreLicense := provideLicense(client, config2)
	datadog := provideDatadog(userStore, repositoryStore, buildStore, system, coreLicense, config2)
	corePubsub := pubsub.New()
	logStore := provideLogStore(db, config2)
	logStream := livelog.New()
	netrcService := provideNetrcService(client, renewer, config2)
	encrypter, err := provideEncrypter(config2)
	if err != nil {
		return application{}, err
	}
	secretStore := secret.New(db, encrypter)
	stepStore := step.New(db)
	buildManager := manager.New(buildStore, configService, corePubsub, logStore, logStream, netrcService, repositoryStore, scheduler, secretStore, statusService, stageStore, stepStore, system, userStore, webhookSender)
	secretService := provideSecretPlugin(config2)
	registryService := provideRegistryPlugin(config2)
	runner := provideRunner(buildManager, secretService, registryService, config2)
	hookService := provideHookService(client, renewer, config2)
	licenseService := license.NewService(userStore, repositoryStore, buildStore, coreLicense)
	permStore := perm.New(db)
	repositoryService := repo.New(client, renewer)
	session := provideSession(userStore, config2)
	batcher := batch.New(db)
	syncer := provideSyncer(repositoryService, repositoryStore, userStore, batcher, config2)
	server := api.New(buildStore, cronStore, corePubsub, hookService, logStore, coreLicense, licenseService, permStore, repositoryStore, repositoryService, scheduler, secretStore, stageStore, stepStore, statusService, session, logStream, syncer, system, triggerer, userStore, webhookSender)
	organizationService := orgs.New(client, renewer)
	userService := user.New(client)
	admissionService := provideAdmissionPlugin(client, organizationService, userService, config2)
	hookParser := parser.New(client)
	middleware := provideLogin(config2)
	options := provideServerOptions(config2)
	webServer := web.New(admissionService, buildStore, client, hookParser, coreLicense, licenseService, middleware, repositoryStore, session, syncer, triggerer, userStore, userService, webhookSender, options, system)
	handler := provideRPC(buildManager, config2)
	metricServer := metric.NewServer(session)
	mux := provideRouter(server, webServer, handler, metricServer)
	serverServer := provideServer(mux, config2)
	mainApplication := newApplication(cronScheduler, datadog, runner, serverServer, userStore)
	return mainApplication, nil
}
