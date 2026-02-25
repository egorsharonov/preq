package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/config"
)

func TestGetConnectionString(t *testing.T) {
	t.Run("valid connection string", func(t *testing.T) {
		cfg := config.PostgresConfig{
			Host:              "localhost",
			Port:              "5432",
			DBName:            "portin",
			MigrationUsername: "migration_admin",
			MigrationPassword: "migration_pass",
			Username:          "admin_app",
			Password:          "admin",
			SSLMode:           "disable",
		}

		expected := "host=localhost port=5432 dbname=portin user=migration_admin password='migration_pass' sslmode=disable"
		actual := cfg.GetMigrationConnectionString()
		require.Equal(t, expected, actual)

		expected = "host=localhost port=5432 dbname=portin user=admin_app password='admin' sslmode=disable"
		actual = cfg.GetAppConnectionString()
		require.Equal(t, expected, actual)

		cfg.Schema = "mnpportin"
		//nolint:lll // full expected string
		expectedWithSchema := "host=localhost port=5432 dbname=portin user=admin_app password='admin' sslmode=disable search_path=mnpportin,public"
		actualWithSchema := cfg.GetAppConnectionString()

		require.Equal(t, expectedWithSchema, actualWithSchema)
	})
}

func TestPortInOrdersConfig(t *testing.T) {
	t.Run("applying default values", func(t *testing.T) {
		expectedDefaultAllowedStatusCodes := []int{-1, -2, -3, -4, 4, 6, 22}

		cfg := config.PortInOrdersConfig{}
		cfg.ApplyDefaults()

		require.False(t, cfg.DisableOpenOrdersCheck)
		assert.Equal(t, expectedDefaultAllowedStatusCodes, cfg.AllowedStatusesForNewOrder)
	})
}
