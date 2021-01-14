package secrethub

import (
	"errors"
	"testing"
	"time"

	"github.com/secrethub/secrethub-cli/internals/cli/ui/fakeui"
	"github.com/secrethub/secrethub-cli/internals/secrethub/fakes"

	"github.com/secrethub/secrethub-go/internals/api"
	"github.com/secrethub/secrethub-go/internals/assert"
	"github.com/secrethub/secrethub-go/internals/errio"
	"github.com/secrethub/secrethub-go/pkg/secrethub"
	"github.com/secrethub/secrethub-go/pkg/secrethub/fakeclient"
)

func TestAccountInspect(t *testing.T) {
	testErr := errio.Namespace("test").Code("test").Error("test error")

	date := time.Date(2018, time.July, 30, 10, 49, 18, 0, time.UTC)

	cases := map[string]struct {
		cmd AccountInspectCommand
		err error
		out string
	}{
		"success user": {
			cmd: AccountInspectCommand{
				newClient: func() (secrethub.ClientInterface, error) {
					return &fakeclient.Client{
						AccountService: &fakeclient.AccountService{
							MeFunc: func() (*api.Account, error) {
								return &api.Account{
									AccountType: accountTypeUser,
								}, nil
							},
						},
						UserService: &fakeclient.UserService{
							MeFunc: func() (*api.User, error) {
								return &api.User{
									Username:      "dev1",
									FullName:      "Developer Uno",
									Email:         "dev1@keylocker.eu",
									EmailVerified: true,
									CreatedAt:     &date,
									PublicKey:     []byte("abcde"),
								}, nil
							},
						},
					}, nil
				},
				timeFormatter: &fakes.TimeFormatter{
					Response: "2018-07-30T10:49:18Z",
				},
			},
			err: nil,
			out: `{
    "Username": "dev1",
    "FullName": "Developer Uno",
    "Email": "dev1@keylocker.eu",
    "EmailVerified": true,
    "AccountType": "user",
    "AccountName": "dev1",
    "CreatedAt": "2018-07-30T10:49:18Z",
    "PublicAccountKey": "YWJjZGU="
}
`,
		},
		"success service": {
			cmd: AccountInspectCommand{
				newClient: func() (secrethub.ClientInterface, error) {
					return &fakeclient.Client{
						AccountService: &fakeclient.AccountService{
							MeFunc: func() (*api.Account, error) {
								return &api.Account{
									AccountType: accountTypeService,
									Name:        "dev1",
									PublicKey:   []byte("abcde"),
								}, nil
							},
						},
						ServiceService: &fakeclient.ServiceService{
							GetFunc: func(id string) (*api.Service, error) {
								if id == "dev1" {
									return &api.Service{
										ServiceID:   "s-abcdefghijkl",
										Description: "Test description.",
										CreatedAt:   time.Date(2020, 12, 12, 12, 12, 12, 0, time.UTC),
									}, nil
								}
								return nil, errors.New(`expected id to be "dev1"`)
							},
						},
					}, nil
				},
				timeFormatter: &fakes.TimeFormatter{
					Response: "2018-07-30T10:49:18Z",
				},
			},
			err: nil,
			out: `{
    "Description": "Test description.",
    "AccountType": "service",
    "AccountName": "s-abcdefghijkl",
    "CreatedAt": "2018-07-30T10:49:18Z",
    "PublicAccountKey": "YWJjZGU="
}
`,
		},
		"client error": {
			cmd: AccountInspectCommand{
				newClient: func() (secrethub.ClientInterface, error) {
					return fakeclient.Client{
						AccountService: &fakeclient.AccountService{
							MeFunc: func() (*api.Account, error) {
								return &api.Account{
									AccountType: accountTypeUser,
								}, nil
							},
						},
						UserService: &fakeclient.UserService{
							MeFunc: func() (*api.User, error) {
								return nil, api.ErrSignatureNotVerified
							},
						},
					}, nil
				},
			},
			err: api.ErrSignatureNotVerified,
			out: "",
		},
		"client creation error": {
			cmd: AccountInspectCommand{
				newClient: func() (secrethub.ClientInterface, error) {
					return nil, testErr
				},
			},
			err: testErr,
			out: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			io := fakeui.NewIO(t)
			tc.cmd.io = io

			// Act
			err := tc.cmd.Run()

			// Assert
			assert.Equal(t, err, tc.err)
			assert.Equal(t, io.Out.String(), tc.out)
		})
	}
}
