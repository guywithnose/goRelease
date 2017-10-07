package command

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// Completion handles bash completion for the commands
func Completion(c *cli.Context) {
	if len(os.Args) > 2 {
		lastParam := os.Args[len(os.Args)-2]
		log.Println(lastParam)
		for _, flag := range c.App.Flags {
			name := strings.Split(flag.GetName(), ",")[0]
			log.Println(name)
			if lastParam == fmt.Sprintf("--%s", name) {
				fmt.Fprintln(c.App.Writer, "fileCompletion")
				return
			}
		}
	}

	completeFlags(c)
}

func completeFlags(c *cli.Context) {
	for _, flag := range c.App.Flags {
		name := strings.Split(flag.GetName(), ",")[0]
		if !c.IsSet(name) {
			fmt.Fprintf(c.App.Writer, "--%s\n", name)
		}
	}
}
