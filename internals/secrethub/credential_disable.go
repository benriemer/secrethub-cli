package secrethub

import (
	"errors"
	"fmt"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui"

	"github.com/secrethub/secrethub-go/internals/api"
)

// CredentialDisableCommand is a command that allows to disable an existing credential.
type CredentialDisableCommand struct {
	io          ui.IO
	force       bool
	fingerprint cli.StringValue
	newClient   newClientFunc
}

// NewCredentialDisableCommand creates a new command for disabling credentials.
func NewCredentialDisableCommand(io ui.IO, newClient newClientFunc) *CredentialDisableCommand {
	return &CredentialDisableCommand{
		io:        io,
		newClient: newClient,
	}
}

// Register registers the command, arguments and flags on the provided Registerer.
func (cmd *CredentialDisableCommand) Register(r cli.Registerer) {
	clause := r.Command("disable", "Disable a credential for usage on SecretHub.")

	fingerprintHelp := fmt.Sprintf("Fingerprint of the credential to disable. At least the first %d characters must be entered.", api.ShortCredentialFingerprintMinimumLength)
	registerForceFlag(clause, &cmd.force)

	clause.BindAction(cmd.Run)
	clause.BindArguments([]cli.Argument{
		{Value: &cmd.fingerprint, Name: "fingerprint", Required: false, Description: fingerprintHelp},
	})
}

// Run disables an existing credential.
func (cmd *CredentialDisableCommand) Run() error {
	client, err := cmd.newClient()
	if err != nil {
		return err
	}

	fingerprint := cmd.fingerprint
	if fingerprint.Value == "" {
		if cmd.force {
			return errors.New("fingerprint argument must be set when using --force")
		}
		fingerprint.Value, err = ui.AskAndValidate(cmd.io, "What is the fingerprint of the credential you want to disable? ", 3, api.ValidateShortCredentialFingerprint)
		if err != nil {
			return err
		}
	}

	err = api.ValidateShortCredentialFingerprint(fingerprint.Value)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.io.Output(),
		"A disabled credential can no longer be used to access SecretHub. "+
			"This process can currently not be reversed.")

	if !cmd.force {
		ok, err := ui.AskYesNo(cmd.io, fmt.Sprintf("Are you sure you want to disable the credential with fingerprint %s?", fingerprint.Value), ui.DefaultNo)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(cmd.io.Output(), "Aborting.")
			return nil
		}
	}

	err = client.Credentials().Disable(fingerprint.Value)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.io.Output(), "Credential disabled.")

	return nil
}
