package main

import (
	"context"

	"log"
	"os/signal"
	"syscall"

	appConfig "gitlab.services.mts.ru/salsa/go-base/application/config"
	"gitlab.services.mts.ru/salsa/go-base/migration/migrate"

	"gitlab.services.mts.ru/salsa/mnp-hub/mnp-datamart/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := appConfig.NewConfig[config.Config](ctx)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	if migrate.HandleMigrationFlag(
		ctx,
		stop,
		cfg.PortInDB.GetMigrationConnectionString(),
		cfg.MigrationsPath,
		cfg.MigrationsVersionTable,
		10) {
		return
	}
}
