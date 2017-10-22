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

func TestReleaseCompletion(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Flags = command.Flags
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
			"--removeOldAssets",
			"",
		},
		output,
	)
}

func TestReleaseCompletionMainPath(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	os.Args = []string{os.Args[0], "release", "--mainPath", "--completion"}
	app, writer, _ := appWithTestWriters()
	app.Flags = command.Flags
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
