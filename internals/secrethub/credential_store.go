package secrethub

import (
	"time"

	"github.com/secrethub/secrethub-go/pkg/secrethub/configdir"
	"github.com/secrethub/secrethub-go/pkg/secrethub/credentials"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui"
)

// Errors
var (
	ErrCredentialNotExist = errMain.Code("credential_not_exist").Error("could not find credential file. Go to https://signup.secrethub.io/ to create an account or run `secrethub init` to use an already existing account on this machine.")
)

// CredentialConfig handles the configuration necessary for local credentials.
type CredentialConfig interface {
	IsPassphraseSet() bool
	Provider() credentials.Provider
	Import() (credentials.Key, error)
	ConfigDir() configdir.Dir
	PassphraseReader() credentials.Reader

	Register(app *cli.App)
}

// NewCredentialConfig creates a new CredentialConfig.
func NewCredentialConfig(io ui.IO) CredentialConfig {
	dir, _ := configdir.Default()
	c := ConfigDir{Dir: *dir}
	return &credentialConfig{
		configDir: c,
		io:        io,
	}
}

type credentialConfig struct {
	configDir                    ConfigDir
	credentialReader             *flagCredentialReader
	credentialPassphrase         string
	CredentialPassphraseCacheTTL time.Duration
	io                           ui.IO
}

func (store *credentialConfig) ConfigDir() configdir.Dir {
	return store.configDir.Dir
}

func (store *credentialConfig) IsPassphraseSet() bool {
	return store.credentialPassphrase != ""
}

// Register registers the flags for configuring the store on the provided Registerer.
func (store *credentialConfig) Register(app *cli.App) {
	commandClause := cli.CommandClause{
		Cmd: app.Cmd,
		App: app,
	}
	commandClause.PersistentFlags().Var(&store.configDir, "config-dir", "The absolute path to a custom configuration directory.")
	store.credentialReader = &flagCredentialReader{}
	store.credentialReader = credentialReader(commandClause.PersistentFlags().StringVar(&store.credentialReader.value, "credential", "", "Use a specific account credential to authenticate to the API. This overrides the credential stored in the configuration directory."))
	commandClause.PersistentFlags().StringVarP(&store.credentialPassphrase, "p", "p", "", "").NoEnvar() // Shorthand -p is deprecated. Use --credential-passphrase instead.
	commandClause.Cmd.Flag("p").Hidden = true
	commandClause.PersistentFlags().StringVar(&store.credentialPassphrase, "credential-passphrase", "", "The passphrase to unlock your credential file. When set, it will not prompt for the passphrase, nor cache it in the OS keyring. Please only use this if you know what you're doing and ensure your passphrase doesn't end up in bash history.")
	commandClause.PersistentFlags().DurationVar(&store.CredentialPassphraseCacheTTL, "credential-passphrase-cache-ttl", 5*time.Minute, "Cache the credential passphrase in the OS keyring for this duration. The cache is automatically cleared after the timer runs out. Each time the passphrase is read from the cache the timer is reset. Passphrase caching is turned on by default for 5 minutes. Turn it off by setting the duration to 0.")
}

// Provider retrieves a credential from the store.
// When a credential is set, that credential is returned,
// otherwise the credential is read from the configured file.
func (store *credentialConfig) Provider() credentials.Provider {
	return credentials.UseKey(store.getCredentialReader()).Passphrase(store.PassphraseReader())
}

func (store *credentialConfig) Import() (credentials.Key, error) {
	return credentials.ImportKey(store.getCredentialReader(), store.PassphraseReader())
}

func (store *credentialConfig) getCredentialReader() credentials.Reader {
	if store.credentialReader.value == "" {
		return store.configDir.Credential()
	}
	return store.credentialReader
}

// PassphraseReader returns a PassphraseReader configured by the flags.
func (store *credentialConfig) PassphraseReader() credentials.Reader {
	return NewPassphraseReader(store.io, store.credentialPassphrase, store.CredentialPassphraseCacheTTL)
}

// credentialReader returns a credential reader and source that reads from the given flag (and its corresponding env var).
func credentialReader(flag *cli.Flag) *flagCredentialReader {
	reader := flagCredentialReader{Flag: flag}
	//flag.IsSetByUser(&reader.setByUser)
	reader.setByUser = &reader.Changed
	return &reader
}

type flagCredentialReader struct {
	*cli.Flag
	value     string
	setByUser *bool
}

func (f *flagCredentialReader) String() string {
	return f.value
}

func (f *flagCredentialReader) Set(s string) error {
	f.value = s
	return nil
}

func (f *flagCredentialReader) Type() string {
	return "flagCredentialReader"
}

func (f *flagCredentialReader) Read() ([]byte, error) {
	return []byte(f.value), nil
}

func (f *flagCredentialReader) Source() string {
	if f.HasEnvarValue() && !*f.setByUser {
		return "$SECRETHUB_CREDENTIAL"
	}
	return "--credential"
}
