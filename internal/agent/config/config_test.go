package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFromFile(t *testing.T) {
	targetConfig := Config{
		ServerURL: "localhost:9999",
		CryptoKey: "./test-crypto-key-pub.pem",
	}

	filepath := prepareTestConfigFile(t, &targetConfig)

	type args struct {
		cfg      *Config
		filepath string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "non-existent file",
			args: args{
				cfg:      New(),
				filepath: "./config-file-must-not-exist",
			},
			wantErr: true,
		},
		{
			name: "file exist ok",
			args: args{
				cfg:      New(),
				filepath: filepath,
			},
			wantErr: false,
		},
		{
			name: "empty filepath",
			args: args{
				cfg:      New(),
				filepath: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.args.cfg, "nil *Config")

			err := ParseFromFile(tt.args.cfg, tt.args.filepath)
			msg := "ParseFromFile() error = %v, wantErr %v"
			if tt.wantErr {
				require.Errorf(t, err, msg, err, tt.wantErr)
				return
			} else {
				require.NoErrorf(t, err, msg, err, tt.wantErr)
			}

			msg = "Config field parsed wrong: %s"
			assert.Equal(t, targetConfig.ServerURL, tt.args.cfg.ServerURL, msg, "ServerAddress")
			assert.Equal(t, targetConfig.CryptoKey, tt.args.cfg.CryptoKey, msg, "CryptoKey")
		})
	}
}

func prepareTestConfigFile(t *testing.T, cfg *Config) (configPath string) {
	t.Helper()

	b, err := json.Marshal(&cfg)
	require.NoError(t, err, "failed to encode Config to JSON format")

	file, err := os.CreateTemp(t.TempDir(), "test-parse-config-*")
	require.NoError(t, err, "failed to create temporary config file for tests")
	defer file.Close()

	_, err = file.Write(b)
	require.NoError(t, err, "failed writing config json to temporary testing file")

	return file.Name()
}
