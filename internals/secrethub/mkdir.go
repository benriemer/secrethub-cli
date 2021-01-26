package secrethub

import (
	"fmt"
	"os"

	"github.com/secrethub/secrethub-go/internals/api"
	"github.com/secrethub/secrethub-go/pkg/secrethub"

	"github.com/secrethub/secrethub-cli/internals/cli"
	"github.com/secrethub/secrethub-cli/internals/cli/ui"
)

// Errors
var (
	ErrMkDirOnRootDir = errMain.Code("mkdir_on_root_dir").Error("You cannot create a directory on the repo path. You can create subdirectories :owner/:repo_name/:directory_name.")
)

// MkDirCommand creates a new directory inside a repository.
type MkDirCommand struct {
	io        ui.IO
	paths     cli.StringListValue
	parents   bool
	newClient newClientFunc
}

// NewMkDirCommand returns a new command.
func NewMkDirCommand(io ui.IO, newClient newClientFunc) *MkDirCommand {
	return &MkDirCommand{
		io:        io,
		newClient: newClient,
	}
}

// Register registers the command, arguments and flags on the provided Registerer.
func (cmd *MkDirCommand) Register(r cli.Registerer) {
	clause := r.Command("mkdir", "Create a new directory.")
	clause.Flags().BoolVar(&cmd.parents, "parents", false, "Create parent directories if needed. Does not error when directories already exist.")

	clause.BindAction(cmd.Run)
	clause.BindArgumentsArr(cli.Argument{Value: &cmd.paths, Name: "path", Required: true, Placeholder: dirPathsPlaceHolder, Description: "The paths to the directories."})
}

// Run executes the command.
func (cmd *MkDirCommand) Run() error {
	client, err := cmd.newClient()
	if err != nil {
		return err
	}

	for _, path := range cmd.paths {
		err := cmd.createDirectory(client, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create a new directory at %s: %s\n", path, err)
		} else {
			fmt.Fprintf(cmd.io.Output(), "Created a new directory at %s\n", path)
		}
	}
	return nil
}

// createDirectory validates the given path and creates a directory on it.
func (cmd *MkDirCommand) createDirectory(client secrethub.ClientInterface, path string) error {
	dirPath, err := api.NewDirPath(path)
	if err != nil {
		return err
	}
	if dirPath.IsRepoPath() {
		return ErrMkDirOnRootDir
	}
	if cmd.parents {
		return client.Dirs().CreateAll(dirPath.Value())
	}
	_, err = client.Dirs().Create(dirPath.Value())
	return err
}
