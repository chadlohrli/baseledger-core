package consensus

// // NodeKeyFactory initializes a types.NodeKey using the given config;
// // a vault id and vault key id are the preferred way to initialize
// // a node key. The node key file is checked if no vault is configured.
// func NodeKeyFactory(cfg *Config) (*types.NodeKey, error) {
// 	var nodeKey *types.NodeKey

// 	if cfg.VaultID != nil && cfg.VaultKeyID != nil {
// 		common.Log.Debugf("baseledger node key configured to use vault: %s", cfg.VaultID)

// 		return &types.NodeKey{
// 			PrivKey: &VaultedPrivateKey{
// 				VaultID:           *cfg.VaultID,
// 				VaultKeyID:        *cfg.VaultKeyID,
// 				VaultRefreshToken: *cfg.VaultRefreshToken,
// 			},
// 		}, nil
// 	}

// 	if _, err := os.Stat(cfg.NodeKeyFile()); err == nil {
// 		nk, err := types.LoadNodeKey(cfg.NodeKeyFile())
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to load baseledger node key; %s", err.Error())
// 		}
// 		nodeKey = &nk
// 		common.Log.Debugf("loaded baseledger node key: %s", cfg.NodeKeyFile())
// 		return nodeKey, nil
// 	}

// 	return nil, errors.New("baseledger node key not configured; a vault id and vault key id or a valid path are required")
// }
