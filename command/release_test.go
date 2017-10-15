package command_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-github/github"
	"github.com/guywithnose/goRelease/command"
	"github.com/guywithnose/runner"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestRelease(t *testing.T) {
	ts := getReleaseTestServer(t, "", "")
	defer ts.Close()
	mainPath := fmt.Sprintf("%s/build", os.TempDir())
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	set.String("mainPath", mainPath, "doc")
	err := set.Parse([]string{"owner", "repo", "tag", "projectName"})
	assert.Nil(t, err)
	defer cleanUp(t, mainPath)
	createFiles(t, mainPath, "tag")
	expectedCommands := getExpectedCommands(t, mainPath)
	expectedRunner := &runner.Test{ExpectedCommands: expectedCommands, AnyOrder: true}
	app, _, _ := appWithTestWriters()
	assert.Nil(t, command.CmdRelease(expectedRunner)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, expectedRunner.ExpectedCommands)
	assert.Equal(t, []error(nil), expectedRunner.Errors)
}

func TestReleaseBuildError(t *testing.T) {
	ts := getReleaseTestServer(t, "", "")
	defer ts.Close()
	mainPath := fmt.Sprintf("%s/build", os.TempDir())
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	set.String("mainPath", mainPath, "doc")
	err := set.Parse([]string{"owner", "repo", "tag", "projectName"})
	assert.Nil(t, err)
	defer cleanUp(t, mainPath)
	createFiles(t, mainPath, "tag")
	goExecutable, err := exec.LookPath("go")
	assert.Nil(t, err)
	expectedCommands := getExpectedCommands(t, mainPath)
	// Skip taring the errored build
	expectedCommands = append(expectedCommands[0:1], expectedCommands[2:]...)
	expectedCommands[0] = runner.NewExpectedCommand(
		mainPath,
		fmt.Sprintf("%s build -o /tmp/build/projectName-linux-386-go1.8-tag", goExecutable),
		"Build error",
		2,
	).WithEnvironment([]string{"GOOS=linux", "GOARCH=386", fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH"))})
	expectedRunner := &runner.Test{ExpectedCommands: expectedCommands, AnyOrder: true}
	app, _, errWriter := appWithTestWriters()
	assert.Nil(t, command.CmdRelease(expectedRunner)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, expectedRunner.ExpectedCommands)
	assert.Equal(t, []error(nil), expectedRunner.Errors)
	output := strings.Split(errWriter.String(), "\n")
	assert.Equal(
		t,
		[]string{
			"Could not run build for linux/386: exit status 2",
			"Output: Build error",
			"",
		},
		output,
	)
}

func TestReleaseCompressError(t *testing.T) {
	ts := getReleaseTestServer(t, "", "")
	defer ts.Close()
	mainPath := fmt.Sprintf("%s/build", os.TempDir())
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	set.String("mainPath", mainPath, "doc")
	err := set.Parse([]string{"owner", "repo", "tag", "projectName"})
	assert.Nil(t, err)
	defer cleanUp(t, mainPath)
	createFiles(t, mainPath, "tag")
	assert.Nil(t, err)
	expectedCommands := getExpectedCommands(t, mainPath)
	expectedCommands[1] = runner.NewExpectedCommand(
		mainPath,
		"gzip /tmp/build/projectName-linux-386-go1.8-tag",
		"Build error",
		2,
	)
	expectedRunner := &runner.Test{ExpectedCommands: expectedCommands, AnyOrder: true}
	app, _, errWriter := appWithTestWriters()
	assert.Nil(t, command.CmdRelease(expectedRunner)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, expectedRunner.ExpectedCommands)
	assert.Equal(t, []error(nil), expectedRunner.Errors)
	output := strings.Split(errWriter.String(), "\n")
	assert.Equal(
		t,
		[]string{
			"Could not compress binary for linux/386: exit status 2",
			"Output: Build error",
			"",
		},
		output,
	)
}

func TestReleaseReleaseFailure(t *testing.T) {
	ts := getReleaseTestServer(t, "/repos/owner/repo/releases?per_page=100", "GET")
	defer ts.Close()
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	err := set.Parse([]string{"owner", "repo", "tag", "projectName"})
	assert.Nil(t, err)
	app, _, _ := appWithTestWriters()
	err = command.CmdRelease(&runner.Test{})(cli.NewContext(app, set, nil))
	assert.EqualError(t, err, fmt.Sprintf("GET %s/repos/owner/repo/releases?per_page=100: 500  []", ts.URL))
}

func TestReleaseReleaseBadApiUrl(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", "%s/mockApi", "doc")
	err := set.Parse([]string{"owner", "repo", "doesntexist", "projectName"})
	assert.Nil(t, err)
	app, _, _ := appWithTestWriters()
	err = command.CmdRelease(&runner.Test{})(cli.NewContext(app, set, nil))
	assert.EqualError(t, err, "parse %s/mockApi: invalid URL escape \"%s/\"")
}

func TestReleaseReleaseCreate(t *testing.T) {
	ts := getReleaseTestServer(t, "", "")
	defer ts.Close()
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	err := set.Parse([]string{"owner", "repo", "doesntexist", "projectName"})
	assert.Nil(t, err)
	app, _, _ := appWithTestWriters()
	err = command.CmdRelease(&runner.Test{})(cli.NewContext(app, set, nil))
	assert.Nil(t, err)
}

func TestReleaseReleaseCreateFail(t *testing.T) {
	ts := getReleaseTestServer(t, "/repos/owner/repo/releases", "POST")
	defer ts.Close()
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	err := set.Parse([]string{"owner", "repo", "doesntexist", "projectName"})
	assert.Nil(t, err)
	app, _, _ := appWithTestWriters()
	err = command.CmdRelease(&runner.Test{})(cli.NewContext(app, set, nil))
	assert.EqualError(t, err, fmt.Sprintf("POST %s/repos/owner/repo/releases: 500  []", ts.URL))
}

func TestReleaseReleaseUpdatePublish(t *testing.T) {
	ts := getReleaseTestServer(t, "", "")
	defer ts.Close()
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	set.Bool("publish", true, "doc")
	mainPath := fmt.Sprintf("%s/build", os.TempDir())
	set.String("mainPath", mainPath, "doc")
	defer cleanUp(t, mainPath)
	createFiles(t, mainPath, "draft")
	err := set.Parse([]string{"owner", "repo", "draft", "projectName"})
	assert.Nil(t, err)
	app, _, _ := appWithTestWriters()
	err = command.CmdRelease(&runner.Test{})(cli.NewContext(app, set, nil))
	assert.Nil(t, err)
}

func TestReleaseReleaseUpdatePublishFailure(t *testing.T) {
	ts := getReleaseTestServer(t, "/repos/owner/repo/releases/2", "PATCH")
	defer ts.Close()
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	set.Bool("publish", true, "doc")
	err := set.Parse([]string{"owner", "repo", "draft", "projectName"})
	assert.Nil(t, err)
	app, _, _ := appWithTestWriters()
	err = command.CmdRelease(&runner.Test{})(cli.NewContext(app, set, nil))
	assert.EqualError(t, err, fmt.Sprintf("PATCH %s/repos/owner/repo/releases/2: 500  []", ts.URL))
}

func TestReleaseReleaseSetOses(t *testing.T) {
	ts := getReleaseTestServer(t, "", "")
	defer ts.Close()
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	set.Bool("publish", true, "doc")
	mainPath := fmt.Sprintf("%s/build", os.TempDir())
	set.String("mainPath", mainPath, "doc")
	defer cleanUp(t, mainPath)
	createFiles(t, mainPath, "draft")
	osFlag := cli.StringSlice{"linux"}
	set.Var(&osFlag, "os", "doc")
	err := set.Parse([]string{"owner", "repo", "draft", "projectName"})
	assert.Nil(t, err)
	app, _, _ := appWithTestWriters()
	err = command.CmdRelease(&runner.Test{})(cli.NewContext(app, set, nil))
	assert.Nil(t, err)
}

func TestReleaseUsage(t *testing.T) {
	ts := getReleaseTestServer(t, "", "")
	defer ts.Close()
	set := flag.NewFlagSet("test", 0)
	set.String("token", "fakeToken", "doc")
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	mainPath := fmt.Sprintf("%s/build", os.TempDir())
	set.String("mainPath", mainPath, "doc")
	defer cleanUp(t, mainPath)
	createFiles(t, mainPath, "tag")
	app, _, _ := appWithTestWriters()
	err := command.CmdRelease(&runner.Test{})(cli.NewContext(app, set, nil))
	assert.EqualError(t, err, "Usage: \"goRelease {owner} {repo} {tagName} {projectName} --token {token} --apiUrl {apiUrl}\"")
}

func TestReleaseNoToken(t *testing.T) {
	ts := getReleaseTestServer(t, "", "")
	defer ts.Close()
	set := flag.NewFlagSet("test", 0)
	set.String("apiUrl", fmt.Sprintf("%s/", ts.URL), "doc")
	mainPath := fmt.Sprintf("%s/build", os.TempDir())
	set.String("mainPath", mainPath, "doc")
	err := set.Parse([]string{"owner", "repo", "tag", "projectName"})
	assert.Nil(t, err)
	app, _, _ := appWithTestWriters()
	runner := &runner.Test{}
	err = command.CmdRelease(runner)(cli.NewContext(app, set, nil))
	assert.EqualError(t, err, "You must specify a token")
}

func getReleaseTestServer(t *testing.T, failureURL, failureMethod string) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	draftID := 2
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == failureURL && r.Method == failureMethod {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if r.URL.String() == "/repos/owner/repo/releases/tags/doesntexist" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.URL.String() == "/repos/owner/repo/releases?per_page=100" {
			tag := "tag"
			idOne := 1
			releases := []*github.RepositoryRelease{
				{TagName: &tag, ID: &idOne},
			}
			bytes, _ := json.Marshal(releases)
			w.Header().Set(
				"Link",
				fmt.Sprintf(
					`<%s/repos/owner/repo/releases?per_page=100&page=2>; rel="next", `+
						`<%s/repos/owner/repo/releases?per_page=100&page=2>; rel="last"`,
					server.URL,
					server.URL,
				),
			)
			response := string(bytes)
			fmt.Fprint(w, response)
			return
		}

		if r.Method == "Patch" {
			expectedBodies := make(map[string]string)
			expectedBodies[fmt.Sprintf("/repos/owner/repo/releases/%d", draftID)] = ""
			body, ok := expectedBodies[r.URL.String()]
			if ok {
				assert.Equal(t, body, r.Body)
			} else {
				t.Errorf("No expected body for %s", r.URL)
			}
		}

		responses := make(map[string]string)

		release := github.RepositoryRelease{}
		bytes, _ := json.Marshal(release)
		responses["/repos/owner/repo/releases"] = string(bytes)

		trueValue := true
		draft := "draft"
		idTwo := 2
		releases := []*github.RepositoryRelease{
			{TagName: &draft, ID: &idTwo, Draft: &trueValue},
		}
		bytes, _ = json.Marshal(releases)
		responses["/repos/owner/repo/releases?page=2&per_page=100"] = string(bytes)

		release = github.RepositoryRelease{
			Draft: &trueValue,
			ID:    &draftID,
		}
		bytes, _ = json.Marshal(release)
		responses["/repos/owner/repo/releases/tags/draft"] = string(bytes)
		responses[fmt.Sprintf("/repos/owner/repo/releases/%d", draftID)] = string(bytes)

		asset := github.ReleaseAsset{}
		bytes, _ = json.Marshal(asset)
		for _, build := range command.ValidBuilds {
			for _, architecture := range build.Architectures {
				responses[fmt.Sprintf(
					"/repos/owner/repo/releases/1/assets?name=projectName-%s-%s-go1.8-tag%s%s",
					build.OperatingSystem,
					architecture,
					build.Extension,
					build.CompressExtension,
				)] = string(bytes)
				responses[fmt.Sprintf(
					"/repos/owner/repo/releases/2/assets?name=projectName-%s-%s-go1.8-draft%s%s",
					build.OperatingSystem,
					architecture,
					build.Extension,
					build.CompressExtension,
				)] = string(bytes)
			}
		}

		responses["/repos/owner/repo/releases/1/assets?name=projectName-linux-386-go1.8-tag"] = string(bytes)

		resp, ok := responses[r.URL.String()]
		if ok {
			fmt.Fprint(w, resp)
			return
		}

		fmt.Printf("Unexpected request: %s\n", r.URL.String())
		w.WriteHeader(http.StatusInternalServerError)
	}))

	return server
}

func TestHelperProcess(*testing.T) {
	runner.ErrorCodeHelper()
}

func cleanUp(t *testing.T, mainPath string) {
	t.Helper()
	err := os.RemoveAll(mainPath)
	assert.Nil(t, err)
}

func createFiles(t *testing.T, path, tagName string) {
	t.Helper()
	err := os.Mkdir(path, 0777)
	assert.Nil(t, err)
	for _, build := range command.ValidBuilds {
		for _, architecture := range build.Architectures {
			err = ioutil.WriteFile(
				fmt.Sprintf("%s/projectName-%s-%s-go1.8-%s%s", path, build.OperatingSystem, architecture, tagName, build.Extension),
				[]byte("foo"),
				0777,
			)
			assert.Nil(t, err)
			err = ioutil.WriteFile(
				fmt.Sprintf("%s/projectName-%s-%s-go1.8-%s%s%s", path, build.OperatingSystem, architecture, tagName, build.Extension, build.CompressExtension),
				[]byte("foo"),
				0777,
			)
			assert.Nil(t, err)
		}
	}
}

func getExpectedCommands(t *testing.T, mainPath string) []*runner.ExpectedCommand {
	t.Helper()
	goExecutable, err := exec.LookPath("go")
	assert.Nil(t, err)
	expectedCommands := []*runner.ExpectedCommand{}
	for _, build := range command.ValidBuilds {
		for _, architecture := range build.Architectures {
			fileName := fmt.Sprintf("%s/projectName-%s-%s-go1.8-tag%s", mainPath, build.OperatingSystem, architecture, build.Extension)
			extra := ""
			if build.IncludeTargetParameter {
				extra = fmt.Sprintf("%s%s ", fileName, build.CompressExtension)
			}

			expectedCommands = append(
				expectedCommands,
				runner.NewExpectedCommand(
					mainPath,
					fmt.Sprintf("%s build -o %s", goExecutable, fileName),
					"",
					0,
				).WithEnvironment([]string{
					fmt.Sprintf("GOOS=%s", build.OperatingSystem),
					fmt.Sprintf("GOARCH=%s", architecture),
					fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")),
				}),
				runner.NewExpectedCommand(
					mainPath,
					fmt.Sprintf(
						"%s %s%s",
						build.CompressBinary,
						extra,
						fileName,
					),
					"",
					0,
				),
				runner.NewExpectedCommand(
					mainPath,
					fmt.Sprintf("%s version", goExecutable),
					"go version go1.8",
					0,
				),
			)
		}
	}

	return expectedCommands
}
