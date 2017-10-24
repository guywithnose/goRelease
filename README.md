**goRelease** asynchronously builds a golang project for all architectures that are supported, and uploads the results to a Github release.

### Usage
```bash
goRelease {owner} {repo} {tagName} {projectName} --token {github_token}
```

### Access Tokens
Refer to this article for creating a Github personal access token
https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/

### Example
```bash
goRelease guywithnose goRelease v0.1.0 goRelease --token 43432a434324c342321f321312d32132b4343287
```

https://github.com/guywithnose/goRelease/releases/tag/v0.1.0

![Release screenshot](https://raw.githubusercontent.com/guywithnose/goRelease/master/screenshot-0.1.0.png)
