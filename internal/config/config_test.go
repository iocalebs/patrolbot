package config_test

import (
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
		name    string
		cfg     config.Config
		want    config.Wiki
		wantErr string
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
			want: config.Wiki{
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
			wantErr: "",
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
			want:    config.Wiki{},
			wantErr: config.ErrWikiNotSet.Error(),
		},
		{
			name: "wiki not found",
			cfg: config.Config{
				Wiki: "foo",
			},
			wantErr: "wiki not found: foo",
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
			wantErr: `invalid wiki configuration for "zwen":
- .auth.username not set
- .auth.password not set
- .reports.templateDir not set`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := test.cfg.CurrentWiki()

			diff := cmp.Diff(test.want, got)
			if diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}

			if err != nil && test.wantErr == "" {
				t.Errorf("Unexpected error: %v", err)
			}

			if err != nil && test.wantErr != "" && err.Error() != test.wantErr {
				t.Errorf("Got error message %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}
