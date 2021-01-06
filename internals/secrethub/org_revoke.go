package secrethub

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui"

	"github.com/secrethub/secrethub-go/internals/api"
)

// OrgRevokeCommand handles revoking a member from an organization.
type OrgRevokeCommand struct {
	orgName   api.OrgName
	username  cli.StringValue
	io        ui.IO
	newClient newClientFunc
}

// NewOrgRevokeCommand creates a new OrgRevokeCommand.
func NewOrgRevokeCommand(io ui.IO, newClient newClientFunc) *OrgRevokeCommand {
	return &OrgRevokeCommand{
		io:        io,
		newClient: newClient,
	}
}

// Register registers the command, arguments and flags on the provided Registerer.
func (cmd *OrgRevokeCommand) Register(r cli.Registerer) {
	clause := r.Command("revoke", "Revoke a user from an organization. This automatically revokes the user from all of the organization's repositories. A list of repositories containing secrets that should be rotated will be printed out.")

	clause.BindAction(cmd.Run)
	clause.BindArguments([]cli.Argument{
		{Value: &cmd.orgName, Name: "org-name", Required: true, Description: "The organization name."},
		{Value: &cmd.username, Name: "username", Required: true, Description: "The username of the user."},
	})
}

// Run revokes an organization member.
func (cmd *OrgRevokeCommand) Run() error {
	client, err := cmd.newClient()
	if err != nil {
		return err
	}

	opts := &api.RevokeOpts{
		DryRun: true,
	}
	planned, err := client.Orgs().Members().Revoke(cmd.orgName.Value(), cmd.username.Value, opts)
	if err != nil {
		return err
	}

	if len(planned.Repos) > 0 {
		fmt.Fprintf(
			cmd.io.Output(),
			"[WARNING] Revoking %s from the %s organization will revoke the user from %d repositories, "+
				"automatically flagging secrets for rotation.\n\n"+
				"A revocation plan has been generated and is shown below. "+
				"Flagged repositories will contain secrets flagged for rotation, "+
				"failed repositories require a manual removal or access rule changes before proceeding and "+
				"OK repos will not require rotation.\n\n",
			cmd.username.Value,
			cmd.orgName,
			len(planned.Repos),
		)

		err = writeOrgRevokeRepoList(cmd.io.Output(), planned.Repos...)
		if err != nil {
			return err
		}

		flagged := planned.StatusCounts[api.StatusFlagged]
		failed := planned.StatusCounts[api.StatusFailed]
		unaffected := planned.StatusCounts[api.StatusOK]

		fmt.Fprintf(cmd.io.Output(), "Revocation plan: %d to flag, %d to fail, %d OK.\n\n", flagged, failed, unaffected)
	} else {
		fmt.Fprintf(
			cmd.io.Output(),
			"The user %s has no memberships to any of %s's repos and can be safely removed.\n\n",
			cmd.username.Value,
			cmd.orgName,
		)
	}

	confirmed, err := ui.ConfirmCaseInsensitive(
		cmd.io,
		"Please type in the username of the user to confirm and proceed with revocation",
		cmd.username.Value,
	)
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Fprintln(cmd.io.Output(), "Name does not match. Aborting.")
		return nil
	}

	fmt.Fprintf(cmd.io.Output(), "\nRevoking user...\n")

	revoked, err := client.Orgs().Members().Revoke(cmd.orgName.Value(), cmd.username.Value, nil)
	if err != nil {
		return err
	}

	if len(revoked.Repos) > 0 {
		fmt.Fprintln(cmd.io.Output(), "")
		err = writeOrgRevokeRepoList(cmd.io.Output(), revoked.Repos...)
		if err != nil {
			return err
		}

		flagged := revoked.StatusCounts[api.StatusFlagged]
		failed := revoked.StatusCounts[api.StatusFailed]
		unaffected := revoked.StatusCounts[api.StatusOK]

		fmt.Fprintf(
			cmd.io.Output(),
			"Revoke complete! Repositories: %d flagged, %d failed, %d OK.\n",
			flagged,
			failed,
			unaffected,
		)
	} else {
		fmt.Fprintln(cmd.io.Output(), "Revoke complete!")
	}

	return nil
}

// writeOrgRevokeRepoList is a helper function that writes repos with a status.
func writeOrgRevokeRepoList(w io.Writer, repos ...*api.RevokeRepoResponse) error {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	for _, resp := range repos {
		fmt.Fprintf(tw, "\t%s/%s\t=> %s\n", resp.Namespace, resp.Name, resp.Status)
	}
	err := tw.Flush()
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "")
	return nil
}
