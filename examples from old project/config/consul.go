package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/consul/api"
	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/datastructures"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

const (
	consulKeyLogTag = "consul.key"
)

type consulSettings struct {
	DisableOpenOrdersCheck             bool                                        `json:"disableOpenOrdersCheck"`
	AllowedStatusesForNewOrder         []string                                    `json:"allowedStatusesForNewOrder"`
	PortationNumbersStatesAggregateMap map[string]portationNumberStateAgregateFlat `json:"portationNumberStatesAggragateMap"`
}

type portationNumberStateAgregateFlat struct {
	TargerNumbersStates  []string         `json:"targetNumbersStates"`
	AggregatedOrderState model.OrderState `json:"aggregatedOrderState"`
}

func NewConsulClient(cfg *ConsulConfig) (*api.Client, error) {
	consulCfg := api.DefaultConfig()
	consulCfg.Address = cfg.Address
	consulCfg.Token = cfg.Token

	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init Consul client: %w", err)
	}

	return client, nil
}

func EnrichConfig(ctx context.Context, currentCfg *Config) (*Config, error) {
	enrichedCfg := currentCfg

	err := enrichedCfg.applyConsulConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich config with consul settings: %w", err)
	}

	enrichedCfg.ApplyDefaults()

	return enrichedCfg, nil
}

func (cfg *Config) applyConsulConfig(ctx context.Context) error {
	consulSettings, err := getConsulSettings(ctx, &cfg.Consul)
	if err != nil {
		return fmt.Errorf("failed to get settings from consul: %w", err)
	}

	log := diagnostics.LoggerFromContext(ctx)

	portInOrdCfg, err := consulSettings.validateAndBuildPortInOrdersCfg()
	if err != nil {
		return fmt.Errorf("failed to build PortInOrders configuration: %w", err)
	}

	cfg.PortInOrders = *portInOrdCfg

	log.Info(fmt.Sprintf("DisableOpenOrdersCheck: %t", cfg.PortInOrders.DisableOpenOrdersCheck))
	log.Info(fmt.Sprintf("AllowedStatusesForNewOrder: %v", cfg.PortInOrders.AllowedStatusesForNewOrder))
	log.Info(fmt.Sprintf("PortationNumbersStatesAgregateMap: %v", cfg.PortInOrders.PortationNumbersStatesAgregateMap))

	return nil
}

func (s *consulSettings) validateAndBuildPortInOrdersCfg() (*PortInOrdersConfig, error) {
	if s == nil {
		return nil, fmt.Errorf("unexpected nil consul settings during enrichemnt")
	}

	cfg := &PortInOrdersConfig{
		DisableOpenOrdersCheck: s.DisableOpenOrdersCheck,
	}

	allowedStates, err := model.StateCodeNamesToInts(s.AllowedStatusesForNewOrder)
	if err != nil {
		return nil, err
	}

	cfg.AllowedStatusesForNewOrder = allowedStates

	portNumStatesAgMap, err := buildAndValidatePortationNumberStateAgregate(s.PortationNumbersStatesAggregateMap)
	if err != nil {
		return nil, err
	}

	cfg.PortationNumbersStatesAgregateMap = portNumStatesAgMap

	return cfg, nil
}

func buildAndValidatePortationNumberStateAgregate(
	flatCfg map[string]portationNumberStateAgregateFlat) (map[string]model.PortationNumberStateAgregate, error) {
	portNumStatesAgMap := make(map[string]model.PortationNumberStateAgregate, len(flatCfg))

	for state, stateAgg := range flatCfg {
		if err := model.ValidCodeName(state); err != nil && state != "transfered" { //nolint:misspell // "transfered" от БДПН
			return nil, err
		}

		if len(stateAgg.TargerNumbersStates) == 0 {
			return nil, fmt.Errorf("portationNumberStatesAggragateMap[%s].targetNumbersStates must not be empty", state)
		}

		for i, numState := range stateAgg.TargerNumbersStates {
			if err := model.ValidCodeName(numState); err != nil && numState != "transfered" { //nolint:misspell // "transfered" от БДПН
				return nil, fmt.Errorf("portationNumberStatesAggragateMap[%s].targetNumbersStates[%d]: %w", state, i, err)
			}
		}

		if err := model.ValidCodeName(stateAgg.AggregatedOrderState.Code); err != nil {
			return nil, fmt.Errorf("portationNumberStatesAggragateMap[%s].aggregatedOrderState.code: %w", state, err)
		}

		portNumStatesAgMap[state] = model.PortationNumberStateAgregate{
			TargerNumbersStates: datastructures.NewHashSetFromArr(stateAgg.TargerNumbersStates),
			AgregatedOrderState: stateAgg.AggregatedOrderState,
		}
	}

	return portNumStatesAgMap, nil
}

func getConsulSettings(ctx context.Context, consulCfg *ConsulConfig) (*consulSettings, error) {
	client, err := NewConsulClient(consulCfg)
	if err != nil {
		return nil, err
	}

	kv := client.KV()
	key := consulCfg.Key
	log := diagnostics.LoggerFromContext(ctx)

	pair, _, err := kv.Get(key, nil)
	if err != nil {
		log.Error("failed to fetch data from Consul",
			zap.String(consulKeyLogTag, key),
			zap.Error(err))

		return nil, err
	}

	if pair == nil {
		log.Error("configuration key not found in Consul",
			zap.String(consulKeyLogTag, key),
			zap.Error(err))

		return nil, fmt.Errorf("configuration key %s not found in Consul", key)
	}

	var settings consulSettings

	if err := json.Unmarshal(pair.Value, &settings); err != nil {
		log.Error("failed to parse KV value",
			zap.String(consulKeyLogTag, key),
			zap.Error(err))

		return nil, err
	}

	return &settings, nil
}
