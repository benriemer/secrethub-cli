package secrethub

import (
	"fmt"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui"

	"github.com/secrethub/secrethub-go/internals/api"
)

// OrgInviteCommand handles inviting a user to an organization.
type OrgInviteCommand struct {
	orgName   api.OrgName
	username  cli.StringValue
	role      string
	force     bool
	io        ui.IO
	newClient newClientFunc
}

// NewOrgInviteCommand creates a new OrgInviteCommand.
func NewOrgInviteCommand(io ui.IO, newClient newClientFunc) *OrgInviteCommand {
	return &OrgInviteCommand{
		io:        io,
		newClient: newClient,
	}
}

// Register registers the command, arguments and flags on the provided Registerer.
func (cmd *OrgInviteCommand) Register(r cli.Registerer) {
	clause := r.Command("invite", "Invite a user to join an organization.")
	clause.Flags().StringVar(&cmd.role, "role", "member", "Assign a role to the invited member. This can be either `admin` or `member`.")
	registerForceFlag(clause, &cmd.force)

	clause.BindAction(cmd.Run)
	clause.BindArguments([]cli.Argument{
		{Value: &cmd.orgName, Name: "org-name", Required: true, Description: "The organization name."},
		{Value: &cmd.username, Name: "username", Required: true, Description: "The username of the user to invite."},
	})
}

// Run invites a user to an organization and gives them a certain role.
func (cmd *OrgInviteCommand) Run() error {
	if !cmd.force {
		msg := fmt.Sprintf("Are you sure you want to invite %s to the %s organization?",
			cmd.username.Value,
			cmd.orgName)

		confirmed, err := ui.AskYesNo(cmd.io, msg, ui.DefaultNo)
		if err != nil {
			return err
		}

		if !confirmed {
			fmt.Fprintln(cmd.io.Output(), "Aborting.")
			return nil
		}
	}

	client, err := cmd.newClient()
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.io.Output(), "Inviting user...")

	resp, err := client.Orgs().Members().Invite(cmd.orgName.Value(), cmd.username.Value, cmd.role)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.io.Output(), "Invite complete! The user %s is now %s of the %s organization.\n", resp.User.Username, resp.Role, cmd.orgName)

	return nil
}
