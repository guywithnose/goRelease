package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/guywithnose/goRelease/command"
	"github.com/guywithnose/runner"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = command.Name
	app.Version = fmt.Sprintf("%s-%s", command.Version, runtime.Version())
	app.Author = "Robert Bittle"
	app.Email = "guywithnose@gmail.com"
	app.Usage = "Update a github release with all binaries that can be built"

	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
		os.Exit(2)
	}

	app.Flags = command.Flags
	app.Action = command.CmdRelease(runner.Real{})
	app.EnableBashCompletion = true
	app.BashComplete = command.Completion
	app.ErrWriter = os.Stderr

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
