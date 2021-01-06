package secrethub

import (
	"fmt"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui"

	"github.com/secrethub/secrethub-go/internals/api"
)

// RepoInviteCommand handles inviting a user to collaborate on a repository.
type RepoInviteCommand struct {
	path      api.RepoPath
	username  cli.StringValue
	force     bool
	io        ui.IO
	newClient newClientFunc
}

// NewRepoInviteCommand creates a new RepoInviteCommand.
func NewRepoInviteCommand(io ui.IO, newClient newClientFunc) *RepoInviteCommand {
	return &RepoInviteCommand{
		io:        io,
		newClient: newClient,
	}
}

// Register registers the command, arguments and flags on the provided Registerer.
func (cmd *RepoInviteCommand) Register(r cli.Registerer) {
	clause := r.Command("invite", "Invite a user to collaborate on a repository.")
	registerForceFlag(clause, &cmd.force)

	clause.BindAction(cmd.Run)
	clause.BindArguments([]cli.Argument{
		{Value: &cmd.path, Name: "repo-path", Required: true, Placeholder: repoPathPlaceHolder, Description: "The repository to invite the user to."},
		{Value: &cmd.username, Name: "username", Required: true, Description: "Username of the user."},
	})
}

// Run invites the configured user to collaborate on the repo.
func (cmd *RepoInviteCommand) Run() error {
	client, err := cmd.newClient()
	if err != nil {
		return err
	}

	if !cmd.force {
		user, err := client.Users().Get(cmd.username.Value)
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("Are you sure you want to add %s to the %s repository?",
			user.PrettyName(),
			cmd.path)

		confirmed, err := ui.AskYesNo(cmd.io, msg, ui.DefaultNo)
		if err != nil {
			return err
		}

		if !confirmed {
			fmt.Fprintln(cmd.io.Output(), "Aborting.")
			return nil
		}
	}
	fmt.Fprintln(cmd.io.Output(), "Inviting user...")

	_, err = client.Repos().Users().Invite(cmd.path.Value(), cmd.username.Value)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.io.Output(), "Invite complete! The user %s is now a member of the %s repository.\n", cmd.username.Value, cmd.path)

	return nil
}
