package secrethub

import (
	"fmt"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui"

	"github.com/secrethub/secrethub-go/internals/api"
)

// RepoInitCommand handles creating new repositories.
type RepoInitCommand struct {
	path      api.RepoPath
	io        ui.IO
	newClient newClientFunc
}

// NewRepoInitCommand creates a new RepoInitCommand
func NewRepoInitCommand(io ui.IO, newClient newClientFunc) *RepoInitCommand {
	return &RepoInitCommand{
		io:        io,
		newClient: newClient,
	}
}

// Register registers the command, arguments and flags on the provided Registerer.
func (cmd *RepoInitCommand) Register(r cli.Registerer) {
	clause := r.Command("init", "Initialize a new repository.")

	clause.BindAction(cmd.Run)
	clause.BindArguments([]cli.Argument{{Value: &cmd.path, Name: "repo-path", Required: true, Description: "Path to the new repository."}})
}

// Run creates a new repository.
func (cmd *RepoInitCommand) Run() error {
	client, err := cmd.newClient()
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.io.Output(), "Creating repository...")

	_, err = client.Repos().Create(cmd.path.Value())
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.io.Output(), "Create complete! The repository %s is now ready to use.\n", cmd.path.String())

	return nil
}
