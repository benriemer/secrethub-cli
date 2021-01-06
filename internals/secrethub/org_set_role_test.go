package secrethub

import (
	"testing"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui/fakeui"

	"github.com/secrethub/secrethub-go/internals/api"
	"github.com/secrethub/secrethub-go/internals/assert"
	"github.com/secrethub/secrethub-go/internals/errio"
	"github.com/secrethub/secrethub-go/pkg/secrethub"
	"github.com/secrethub/secrethub-go/pkg/secrethub/fakeclient"
)

func TestOrgSetRoleCommand_Run(t *testing.T) {
	testErr := errio.Namespace("test").Code("test").Error("test error")

	cases := map[string]struct {
		cmd          OrgSetRoleCommand
		newClientErr error
		updateFunc   func(org string, username string, role string) (*api.OrgMember, error)
		ArgOrgName   api.OrgName
		ArgUsername  string
		ArgRole      string
		out          string
		err          error
	}{
		"success": {
			cmd: OrgSetRoleCommand{
				username: cli.StringValue{Value: "dev1"},
				orgName:  "company",
				role:     cli.StringValue{Value: api.OrgRoleMember},
			},
			updateFunc: func(org string, username string, role string) (*api.OrgMember, error) {
				return &api.OrgMember{
					User: &api.User{
						Username: "dev1",
					},
					Role: api.OrgRoleMember,
				}, nil
			},
			ArgOrgName:  "company",
			ArgUsername: "dev1",
			ArgRole:     api.OrgRoleMember,
			out: "Setting role...\n" +
				"Set complete! The user dev1 is member of the company organization.\n",
		},
		"new client error": {
			newClientErr: testErr,
			err:          testErr,
		},
		"update org member error": {
			updateFunc: func(org string, username string, role string) (*api.OrgMember, error) {
				return nil, testErr
			},
			out: "Setting role...\n",
			err: testErr,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var argOrgName string
			var argUsername string
			var argRole string

			// Setup
			tc.cmd.newClient = func() (secrethub.ClientInterface, error) {
				return fakeclient.Client{
					OrgService: &fakeclient.OrgService{
						MembersService: &fakeclient.OrgMemberService{
							UpdateFunc: func(org string, username string, role string) (*api.OrgMember, error) {
								argOrgName = org
								argUsername = username
								argRole = role
								return tc.updateFunc(org, username, role)
							},
						},
					},
				}, tc.newClientErr
			}

			io := fakeui.NewIO(t)
			tc.cmd.io = io

			// Run
			err := tc.cmd.Run()

			// Assert
			assert.Equal(t, err, tc.err)
			assert.Equal(t, io.Out.String(), tc.out)
			assert.Equal(t, argOrgName, tc.ArgOrgName)
			assert.Equal(t, argUsername, tc.ArgUsername)
			assert.Equal(t, argRole, tc.ArgRole)
		})
	}
}
