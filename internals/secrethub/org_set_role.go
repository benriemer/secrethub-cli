package secrethub

import (
	"fmt"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui"

	"github.com/secrethub/secrethub-go/internals/api"
)

// OrgSetRoleCommand handles updating the role of an organization member.
type OrgSetRoleCommand struct {
	orgName   api.OrgName
	username  cli.StringValue
	role      cli.StringValue
	io        ui.IO
	newClient newClientFunc
}

// NewOrgSetRoleCommand creates a new OrgSetRoleCommand.
func NewOrgSetRoleCommand(io ui.IO, newClient newClientFunc) *OrgSetRoleCommand {
	return &OrgSetRoleCommand{
		io:        io,
		newClient: newClient,
	}
}

// Register registers the command, arguments and flags on the provided Registerer.
func (cmd *OrgSetRoleCommand) Register(r cli.Registerer) {
	clause := r.Command("set-role", "Set a user's organization role.")

	clause.BindAction(cmd.Run)
	clause.BindArguments([]cli.Argument{
		{Value: &cmd.orgName, Name: "org-name", Required: true, Description: "The organization name."},
		{Value: &cmd.username, Name: "username", Required: true, Description: "The username of the user."},
		{Value: &cmd.role, Name: "role", Required: true, Description: "The role to assign to the user. Can be either `admin` or `member`."},
	})
}

// Run updates the role of an organization member.
func (cmd *OrgSetRoleCommand) Run() error {
	client, err := cmd.newClient()
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.io.Output(), "Setting role...\n")

	resp, err := client.Orgs().Members().Update(cmd.orgName.Value(), cmd.username.Value, cmd.role.Value)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.io.Output(), "Set complete! The user %s is %s of the %s organization.\n", resp.User.Username, resp.Role, cmd.orgName)

	return nil
}
