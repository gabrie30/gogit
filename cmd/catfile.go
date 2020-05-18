package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ssrathi/gogit/git"
	"github.com/ssrathi/gogit/util"
)

// CatFileCommand lists the components of "cat-file" comamnd.
type CatFileCommand struct {
	fs       *flag.FlagSet
	revision string
	getType  bool
	getSize  bool
	printObj bool
}

// NewCatFileCommand creates a new command object.
func NewCatFileCommand() *CatFileCommand {
	cmd := &CatFileCommand{
		fs: flag.NewFlagSet("cat-file", flag.ExitOnError),
	}

	cmd.fs.BoolVar(&cmd.getType, "t", false,
		"Instead of the content, show the object type identified by <object>")
	cmd.fs.BoolVar(&cmd.getSize, "s", false,
		"Instead of the content, show the object size identified by <object>")
	cmd.fs.BoolVar(&cmd.printObj, "p", false,
		"Pretty-print the contents of <object> based on its type.")
	return cmd
}

// Name gives the name of the command.
func (cmd *CatFileCommand) Name() string {
	return cmd.fs.Name()
}

// Description gives the description of the command.
func (cmd *CatFileCommand) Description() string {
	return "Provide content or type and size information for repository objects"
}

// Init initializes and validates the given command.
func (cmd *CatFileCommand) Init(args []string) error {
	cmd.fs.Usage = cmd.Usage
	if err := cmd.fs.Parse(args); err != nil {
		return err
	}

	if cmd.fs.NArg() < 1 {
		return errors.New("Error: Missing <object> argument\n")
	}

	// All optional boolean args are mutually exclusive
	if !(cmd.getType || cmd.getSize || cmd.printObj) {
		return errors.New("Error: one of '-t', '-s' or '-p' must be provided\n")
	}

	if cmd.getType && cmd.getSize {
		return errors.New("Error: switch 't' and 's' are incompatible\n")
	}
	if cmd.getSize && cmd.printObj {
		return errors.New("Error: switch 's' and 'p' are incompatible\n")
	}
	if cmd.printObj && cmd.getType {
		return errors.New("Error: switch 'p' and 't' are incompatible\n")
	}

	cmd.revision = cmd.fs.Arg(0)
	return nil
}

// Usage prints the usage string for the end user.
func (cmd *CatFileCommand) Usage() {
	fmt.Printf("%s - %s\n", cmd.Name(), cmd.Description())
	fmt.Printf("usage: %s [<args>] <object>\n", cmd.Name())
	cmd.fs.PrintDefaults()
}

// Execute runs the given command till completion.
func (cmd *CatFileCommand) Execute() {
	repo, err := git.GetRepo(".")
	util.Check(err)

	// Resolve the given hash to a full hash.
	objHash, err := repo.ObjectFind(cmd.revision)
	util.Check(err)

	obj, err := repo.ObjectParse(objHash)
	util.Check(err)

	var gitType git.GitType
	switch obj.ObjType {
	case "blob":
		gitType, err = git.NewBlob(obj)
		util.Check(err)
	case "tree":
		gitType, err = git.NewTree(obj)
		util.Check(err)
	case "commit":
		gitType, err = git.NewCommit(obj)
		util.Check(err)
	}

	// Only one of 'printObj', 'getType' and 'getSize' is provided.
	if cmd.printObj {
		fmt.Print(gitType.Print())
	} else if cmd.getType {
		fmt.Println(gitType.Type())
	} else if cmd.getSize {
		fmt.Println(gitType.DataSize())
	}
}