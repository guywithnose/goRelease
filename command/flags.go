package command

import "github.com/urfave/cli"

// Flags is the valid command parameters
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:   "token, t",
		Usage:  "The github access token for this profile",
		EnvVar: "GO_RELEASE_GITHUB_TOKEN",
	},
	cli.StringFlag{
		Name:  "apiUrl, a",
		Usage: "The url for accessing the github API (You only need to specify this for Enterprise Github)",
	},
	cli.StringFlag{
		Name:  "mainPath, p",
		Usage: "The path that contains the main package (Default: current)",
	},
	cli.StringSliceFlag{
		Name:  "os",
		Usage: "Set the OSes to build against",
	},
	cli.BoolFlag{
		Name:  "publish",
		Usage: "Should the new release be published.  If not specified and the release does not exist, the release will be created as draft.",
	},
}
