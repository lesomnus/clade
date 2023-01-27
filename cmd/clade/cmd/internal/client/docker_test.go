package client_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/stretchr/testify/require"
)

func TestLoadAuthFromDockerConfig(t *testing.T) {
	t.Run("load basic auths from Docker config", func(t *testing.T) {
		require := require.New(t)

		svc := "cr.io"
		username := "hypnos"
		password := "secure"

		config := []byte(fmt.Sprintf(`{
			"auths": {
				"%s": {
					"auth": "%s"
				}
			}
		}`, svc, base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", username, password)))))

		tmp := t.TempDir()
		config_path := filepath.Join(tmp, "config.json")

		err := os.WriteFile(config_path, config, 0644)
		require.NoError(err)

		auths, err := client.LoadAuthFromDockerConfig(config_path)
		require.NoError(err)
		require.Len(auths, 1)
		require.Contains(auths, svc)
		require.Equal(username, auths[svc].Username)
		require.Equal(password, auths[svc].Password)
	})

	t.Run("return empty map if path is empty", func(t *testing.T) {
		require := require.New(t)

		auths, err := client.LoadAuthFromDockerConfig("")
		require.NoError(err)
		require.Empty(auths)
	})

	t.Run("fails if", func(t *testing.T) {
		t.Run("path does not exists", func(t *testing.T) {
			require := require.New(t)

			_, err := client.LoadAuthFromDockerConfig("not exists")
			require.ErrorIs(err, os.ErrNotExist)
		})

		tcs := []struct {
			desc string
			data string
			msgs []string
		}{
			{
				desc: "config is not JSON",
				data: "{invalid JSON}",
				msgs: []string{"failed", "parse"},
			},
			{
				desc: "value is not base64",
				data: `{"auths": {"cr.io": {"auth": "not base64"}}}`,
				msgs: []string{"failed", "decode"},
			},
			{
				desc: "auth is ill-formed",
				data: func() string {
					val := base64.StdEncoding.EncodeToString([]byte("no colon included"))
					return fmt.Sprintf(`{"auths": {"cr.io": {"auth": "%s"}}}`, string(val))
				}(),
				msgs: []string{"invalid", "auth", "cr.io"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				tmp := t.TempDir()
				config_path := filepath.Join(tmp, "config.json")

				err := os.WriteFile(config_path, []byte(tc.data), 0644)
				require.NoError(err)

				_, err = client.LoadAuthFromDockerConfig(config_path)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}
