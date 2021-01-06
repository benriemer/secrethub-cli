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

func TestRepoInviteCommand_Run(t *testing.T) {
	testErr := errio.Namespace("test").Code("test").Error("test error")

	cases := map[string]struct {
		cmd               RepoInviteCommand
		newClientErr      error
		GetFunc           func(username string) (*api.User, error)
		InviteFunc        func(path string, username string) (*api.RepoMember, error)
		getArgUsername    string
		inviteArgUsername string
		inviteArgPath     api.RepoPath
		out               string
		err               error
	}{
		"new client error": {
			newClientErr: testErr,
			err:          testErr,
		},
		"get user error": {
			GetFunc: func(username string) (*api.User, error) {
				return nil, testErr
			},
			err: testErr,
		},
		"success force": {
			cmd: RepoInviteCommand{
				path:     "dev2/repo",
				username: cli.StringValue{Value: "dev1"},
				force:    true,
			},
			InviteFunc: func(path string, username string) (*api.RepoMember, error) {
				return &api.RepoMember{}, nil
			},
			inviteArgUsername: "dev1",
			inviteArgPath:     "dev2/repo",
			out:               "Inviting user...\nInvite complete! The user dev1 is now a member of the dev2/repo repository.\n",
		},
		"invite error": {
			cmd: RepoInviteCommand{
				path:     "dev2/repo",
				username: cli.StringValue{Value: "dev1"},
				force:    true,
			},
			InviteFunc: func(path string, username string) (*api.RepoMember, error) {
				return nil, testErr
			},
			inviteArgUsername: "dev1",
			inviteArgPath:     "dev2/repo",
			out:               "Inviting user...\n",
			err:               testErr,
		},
		// TODO SHDEV-1029: Add cases for confirm and abort after extracting AskForConfirmation out of ui.IO.
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var argInviteUsername string
			var argGetUsername string
			var argPath string

			// Setup
			if tc.newClientErr != nil {
				tc.cmd.newClient = func() (secrethub.ClientInterface, error) {
					return nil, tc.newClientErr
				}
			} else {
				tc.cmd.newClient = func() (secrethub.ClientInterface, error) {
					return fakeclient.Client{
						RepoService: &fakeclient.RepoService{
							UserService: &fakeclient.RepoUserService{
								InviteFunc: func(path string, username string) (*api.RepoMember, error) {
									argPath = path
									argInviteUsername = username
									return tc.InviteFunc(path, username)
								},
							},
						},
						UserService: &fakeclient.UserService{
							GetFunc: func(username string) (*api.User, error) {
								argGetUsername = username
								return tc.GetFunc(username)
							},
						},
					}, nil
				}
			}

			io := fakeui.NewIO(t)
			tc.cmd.io = io

			// Run
			err := tc.cmd.Run()

			// Assert
			assert.Equal(t, err, tc.err)
			assert.Equal(t, io.Out.String(), tc.out)
			assert.Equal(t, argGetUsername, tc.getArgUsername)
			assert.Equal(t, argInviteUsername, tc.inviteArgUsername)
			assert.Equal(t, argPath, tc.inviteArgPath)
		})
	}
}
