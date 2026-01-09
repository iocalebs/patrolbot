package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/pflag"
)

// See testdata/integration-tests/config_view.txt for further coverage of config.Load.
func TestLoadErrNoConfigFlag(t *testing.T) {
	t.Parallel()

	fs := pflag.NewFlagSet("empty", pflag.ContinueOnError)
	_, err := config.Load(fs)

	want := "error reading config flag: flag accessed but not defined: config"
	if err == nil || err.Error() != want {
		t.Errorf("Expected error %q, got: %v", want, err)
	}
}

func TestCurrentWiki(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cfg          config.Config
		expectResult config.Wiki
		expectErr    error
		expectErrMsg string
	}{
		{
			name: "Valid wiki config",
			cfg: config.Config{
				Wiki: "zwen",
				Wikis: map[string]config.Wiki{
					"zwen": {
						Site: config.WikiSite{
							URL:         "https://zeldawiki.wiki",
							ScriptPath:  "/w",
							ArticlePath: "/wiki",
						},
						Auth: config.WikiAuth{
							Username: "PhantomCaleb@PatrolBot",
							Password: "password",
						},
						Reports: config.Reports{
							TemplateDir: "./templates",
						},
					},
				},
			},
			expectResult: config.Wiki{
				Site: config.WikiSite{
					URL:         "https://zeldawiki.wiki",
					ScriptPath:  "/w",
					ArticlePath: "/wiki",
				},
				Auth: config.WikiAuth{
					Username: "PhantomCaleb@PatrolBot",
					Password: "password",
				},
				Reports: config.Reports{
					TemplateDir: "./templates",
				},
			},
			expectErr: nil,
		},
		{
			name: "wiki not set",
			cfg: config.Config{
				Wiki: "",
				Wikis: map[string]config.Wiki{
					"zwen": {
						Site: config.WikiSite{
							URL:         "https://zeldawiki.wiki",
							ScriptPath:  "/w",
							ArticlePath: "/wiki",
						},
						Auth: config.WikiAuth{
							Username: "PhantomCaleb@PatrolBot",
							Password: "password",
						},
					},
				},
			},
			expectResult: config.Wiki{},
			expectErr:    config.ErrWikiNotSet,
		},
		{
			name: "wiki not found",
			cfg: config.Config{
				Wiki: "foo",
			},
			expectResult: config.Wiki{},
			expectErr:    config.ErrWikiNotFound,
		},
		{
			name: "required config not set",
			cfg: config.Config{
				Wiki: "zwen",
				Wikis: map[string]config.Wiki{
					"zwen": {
						Site: config.WikiSite{
							URL:         "https://zeldawiki.wiki",
							ScriptPath:  "/w",
							ArticlePath: "/wiki",
						},
						Auth: config.WikiAuth{
							Username: "",
							Password: "",
						},
						Reports: config.Reports{
							TemplateDir: "",
						},
					},
				},
			},
			expectErr: config.ErrWikiInvalid,
			expectErrMsg: `invalid wiki configuration for "zwen": 
- .auth.username not set
- .auth.password not set
- .reports.templateDir not set`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := test.cfg.CurrentWiki()

			diff := cmp.Diff(test.expectResult, got)
			if diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}

			if !errors.Is(err, test.expectErr) {
				t.Errorf("Expected error %v, got %v", test.expectErr, err)
			}

			if err != nil && !strings.Contains(err.Error(), test.expectErrMsg) {
				t.Errorf("Expected error message to contain %q, got %q", test.expectErrMsg, err.Error())
			}
		})
	}
}
