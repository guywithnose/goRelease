package command_test

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/guywithnose/goRelease/command"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestRootCompletion(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = append(command.Commands, cli.Command{Hidden: true, Name: "don't show"})
	os.Args = []string{os.Args[0], "release"}
	command.RootCompletion(cli.NewContext(app, set, nil))
	assert.Equal(t, "release:Create or update a release\n", writer.String())
}

func TestRootCompletionConfig(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = command.Commands
	os.Args = []string{os.Args[0], "release", "--config", "--completion"}
	command.RootCompletion(cli.NewContext(app, set, nil))
	assert.Equal(t, "fileCompletion\n", writer.String())
}

func TestReleaseCompletion(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = command.Commands
	command.Completion(cli.NewContext(app, set, nil))
	output := strings.Split(writer.String(), "\n")
	assert.Equal(
		t,
		[]string{
			"--token",
			"--apiUrl",
			"--mainPath",
			"--os",
			"--publish",
			"",
		},
		output,
	)
}

func TestReleaseCompletionMainPath(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	os.Args = []string{"goRelease", "release", "--mainPath", ""}
	app, writer, _ := appWithTestWriters()
	app.Commands = command.Commands
	command.Completion(cli.NewContext(app, set, nil))
	output := strings.Split(writer.String(), "\n")
	assert.Equal(
		t,
		[]string{
			"fileCompletion",
			"",
		},
		output,
	)
}

func appWithTestWriters() (*cli.App, *bytes.Buffer, *bytes.Buffer) {
	app := cli.NewApp()
	writer := new(bytes.Buffer)
	errWriter := new(bytes.Buffer)
	app.Writer = writer
	app.ErrWriter = errWriter
	return app, writer, errWriter
}
