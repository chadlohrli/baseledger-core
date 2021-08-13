package protocol

import (
	"errors"
	"fmt"

	"github.com/providenetwork/baseledger/common"
	"github.com/provideplatform/provide-go/api/baseline"
	"github.com/provideplatform/provide-go/api/ident"
	"github.com/provideplatform/provide-go/api/nchain"
	"github.com/provideplatform/provide-go/api/privacy"
	"github.com/provideplatform/provide-go/api/vault"
)

// Service instances expose a compliant implementation of the Baseline protocol
type Service struct {
	baseline *baseline.Service
	ident    *ident.Service
	nchain   *nchain.Service
	privacy  *privacy.Service
	vault    *vault.Service
}

func authorizeAccessToken(refreshToken string) (*ident.Token, error) {
	token, err := ident.CreateToken(refreshToken, map[string]interface{}{
		"grant_type": "refresh_token",
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func serviceFactory(cfg *common.Config) (*Service, error) {
	if cfg.ProvideRefreshToken == nil {
		return nil, errors.New("failed to initialize service; no bearer refresh token provided")
	}

	token, err := authorizeAccessToken(*cfg.ProvideRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service; no bearer access token authorized; %s", err.Error())
	}

	return &Service{
		baseline: baseline.InitBaselineService(*token.AccessToken),
		ident:    ident.InitIdentService(token.AccessToken),
		nchain:   nchain.InitNChainService(*token.AccessToken),
		privacy:  privacy.InitPrivacyService(*token.AccessToken),
		vault:    vault.InitVaultService(token.AccessToken),
	}, nil
}
