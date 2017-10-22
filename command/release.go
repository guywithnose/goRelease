package command

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/guywithnose/runner"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
)

// CmdRelease builds a release
func CmdRelease(cmdWrapper runner.Builder) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		return cmdReleaseHelper(c, cmdWrapper)
	}
}

func cmdReleaseHelper(c *cli.Context, cmdWrapper runner.Builder) error {
	token := c.String("token")
	apiURL := c.String("apiUrl")
	publish := c.Bool("publish")
	removeOldAssets := c.Bool("removeOldAssets")
	_ = c.StringSlice("os")
	mainPath := c.String("mainPath")
	if token == "" {
		return cli.NewExitError("You must specify a token", 1)
	}

	if c.NArg() != 4 {
		return cli.NewExitError("Usage: \"goRelease {owner} {repo} {tagName} {projectName} --token {token} --apiUrl {apiUrl}\"", 1)
	}

	var err error
	if mainPath == "" {
		mainPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Unable to get current working directory: %v", err)
		}
	}

	owner := c.Args().Get(0)
	repo := c.Args().Get(1)
	tagName := c.Args().Get(2)
	projectName := c.Args().Get(3)
	client, err := getGithubClient(&token, &apiURL)
	if err != nil {
		return err
	}

	releaseResponse, err := getRelease(client, owner, repo, tagName, publish)
	if err != nil {
		return err
	}

	id := releaseResponse.GetID()

	binaries, err := buildBinaries(cmdWrapper, mainPath, projectName, tagName, c.App.ErrWriter)
	if err != nil {
		return err
	}

	if removeOldAssets {
		err = clearAssets(client, id, owner, repo)
		if err != nil {
			return err
		}
	}

	uploadBinaries(client, owner, repo, id, binaries, c.App.ErrWriter)
	return nil
}

func clearAssets(client *github.Client, id int, owner, repo string) error {
	assets, err := getAssets(client, id, owner, repo)
	if err != nil {
		return err
	}

	for _, asset := range assets {
		_, err = client.Repositories.DeleteReleaseAsset(context.Background(), owner, repo, asset.GetID())
		if err != nil {
			return err
		}
	}

	return err
}

func getAssets(client *github.Client, id int, owner, repo string) ([]*github.ReleaseAsset, error) {
	opt := github.ListOptions{
		PerPage: 100,
	}

	allAssets := make([]*github.ReleaseAsset, 0, 100)
	for {
		assets, resp, err := client.Repositories.ListReleaseAssets(context.Background(), owner, repo, id, &opt)
		if err != nil {
			return nil, err
		}

		allAssets = append(allAssets, assets...)

		if resp.NextPage == 0 {
			return allAssets, nil
		}

		opt.Page = resp.NextPage
	}
}

func uploadBinaries(client *github.Client, owner, repo string, id int, binaries <-chan string, errWriter io.Writer) {
	for fileName := range binaries {
		err := uploadToRelease(client, id, owner, repo, fileName)
		if err != nil {
			fmt.Fprintf(errWriter, "Unable to upload binary %s: %v\n", fileName, err)
		} else {
			err := os.Remove(fileName)
			if err != nil {
				fmt.Fprintf(errWriter, "Unable to cleanup binary %s: %v\n", fileName, err)
			}
		}
	}
}

func uploadToRelease(client *github.Client, id int, owner, repo, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	_, _, err = client.Repositories.UploadReleaseAsset(context.Background(), owner, repo, id, &github.UploadOptions{Name: path.Base(fileName)}, file)
	return err
}

func buildBinaries(cmdWrapper runner.Builder, mainPath, projectName, tagName string, errWriter io.Writer) (<-chan string, error) {
	files := make(chan string, 10)
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		return files, err
	}

	wg := sync.WaitGroup{}
	for _, build := range ValidBuilds {
		for _, architecture := range build.Architectures {
			wg.Add(1)
			go func(build osBuildInfo, architecture string) {
				defer wg.Done()
				cmd := cmdWrapper.New(mainPath, goExecutable, "version")
				versionInfo, _ := cmd.CombinedOutput()
				goVersion := "UNKNOWN"
				versionParts := strings.Split(string(versionInfo), " ")
				if len(versionParts) >= 3 {
					goVersion = strings.Split(string(versionInfo), " ")[2]
				}

				fileName := fmt.Sprintf("%s/%s-%s-%s-%s-%s%s", mainPath, projectName, build.OperatingSystem, architecture, goVersion, tagName, build.Extension)
				environment := []string{
					fmt.Sprintf("GOOS=%s", build.OperatingSystem),
					fmt.Sprintf("GOARCH=%s", architecture),
					fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")),
				}
				cmd = cmdWrapper.NewWithEnvironment(mainPath, environment, goExecutable, "build", "-o", fileName)
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Fprintf(errWriter, "Could not run build for %s/%s: %v\nOutput: %s\n", build.OperatingSystem, architecture, err, output)
				} else {
					_, err := exec.LookPath(build.CompressBinary)
					if err != nil {
						fmt.Fprintf(errWriter, "Could not compress binary for %s/%s: %v\n", build.OperatingSystem, architecture, err)
						files <- fileName
					}

					compressedBinary := fmt.Sprintf("%s%s", fileName, build.CompressExtension)
					var command []string
					if build.IncludeTargetParameter {
						command = []string{build.CompressBinary, compressedBinary, fileName}
					} else {
						command = []string{build.CompressBinary, fileName}
					}

					cmd := cmdWrapper.New(mainPath, command...)
					output, err := cmd.CombinedOutput()
					if err != nil {
						fmt.Fprintf(errWriter, "Could not compress binary for %s/%s: %v\nOutput: %s\n", build.OperatingSystem, architecture, err, output)
						files <- fileName
					} else {
						files <- compressedBinary
						_ = os.Remove(fileName)
					}
				}
			}(build, architecture)
		}
	}

	go func() {
		wg.Wait()
		close(files)
	}()

	return files, nil
}

func getRelease(client *github.Client, owner, repo, tagName string, publish bool) (*github.RepositoryRelease, error) {
	draft := !publish
	releases, err := getReleases(client, owner, repo)
	if err != nil {
		return nil, err
	}

	for _, release := range releases {
		if release.GetTagName() == tagName {
			if release.GetDraft() && publish {
				releasePatch := github.RepositoryRelease{
					Draft: &draft,
				}

				_, _, err = client.Repositories.EditRelease(context.Background(), owner, repo, release.GetID(), &releasePatch)
				if err != nil {
					return nil, err
				}
			}

			return release, nil
		}
	}

	release := github.RepositoryRelease{
		TagName: &tagName,
		Draft:   &draft,
	}

	newRelease, _, err := client.Repositories.CreateRelease(context.Background(), owner, repo, &release)
	if err != nil {
		return nil, err
	}

	return newRelease, nil
}

func getReleases(client *github.Client, owner, repo string) ([]*github.RepositoryRelease, error) {
	opt := github.ListOptions{
		PerPage: 100,
	}

	allReleases := make([]*github.RepositoryRelease, 0, 100)
	for {
		releases, resp, err := client.Repositories.ListReleases(context.Background(), owner, repo, &opt)
		if err != nil {
			return nil, err
		}

		allReleases = append(allReleases, releases...)

		if resp.NextPage == 0 {
			return allReleases, nil
		}

		opt.Page = resp.NextPage
	}
}

func getGithubClient(token, apiURL *string) (*github.Client, error) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *token})
	tokenClient := oauth2.NewClient(context.Background(), tokenSource)

	client := github.NewClient(tokenClient)
	if apiURL != nil && *apiURL != "" {
		url, err := url.Parse(*apiURL)
		if err != nil {
			return nil, err
		}

		client.BaseURL = url
		client.UploadURL = url
	}

	return client, nil
}
