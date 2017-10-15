package command

type osBuildInfo struct {
	OperatingSystem        string
	Architectures          []string
	CompressBinary         string
	IncludeTargetParameter bool
	CompressExtension      string
	Extension              string
}

// ValidBuilds defines the builds that should be created and uploaded
var ValidBuilds = []osBuildInfo{
	{
		OperatingSystem:   "linux",
		Architectures:     []string{"386", "amd64", "arm", "arm64", "mips", "mips64", "mips64le", "mipsle", "ppc64", "ppc64le", "s390x"},
		CompressBinary:    "gzip",
		CompressExtension: ".gz",
	},
	{
		OperatingSystem:   "darwin",
		Architectures:     []string{"386", "amd64", "arm", "arm64"},
		CompressBinary:    "gzip",
		CompressExtension: ".gz",
	},
	{
		OperatingSystem:   "nacl",
		Architectures:     []string{"386", "amd64p32", "arm"},
		CompressBinary:    "gzip",
		CompressExtension: ".gz",
	},
	{
		OperatingSystem:   "netbsd",
		Architectures:     []string{"386", "amd64", "arm"},
		CompressBinary:    "gzip",
		CompressExtension: ".gz",
	},
	{
		OperatingSystem:   "openbsd",
		Architectures:     []string{"386", "amd64", "arm"},
		CompressBinary:    "gzip",
		CompressExtension: ".gz",
	},
	{
		OperatingSystem:   "plan9",
		Architectures:     []string{"386", "amd64", "arm"},
		CompressBinary:    "gzip",
		CompressExtension: ".gz",
	},
	{
		OperatingSystem:   "solaris",
		Architectures:     []string{"amd64"},
		CompressBinary:    "gzip",
		CompressExtension: ".gz",
	},
	{
		OperatingSystem:   "dragonfly",
		Architectures:     []string{"amd64"},
		CompressBinary:    "gzip",
		CompressExtension: ".gz",
	},
	{
		OperatingSystem:        "windows",
		Architectures:          []string{"386", "amd64"},
		CompressBinary:         "zip",
		IncludeTargetParameter: true,
		CompressExtension:      ".zip",
		Extension:              ".exe",
	},
}
