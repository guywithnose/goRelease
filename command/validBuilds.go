package command

type osBuildInfo struct {
	OperatingSystem    string
	Architectures      []string
	CompressBinary     string
	CompressParameters []string
	CompressExtension  string
	Extension          string
}

// ValidBuilds defines the builds that should be created and uploaded
var ValidBuilds = []osBuildInfo{
	{
		OperatingSystem:    "linux",
		Architectures:      []string{"386", "amd64", "arm", "arm64", "mips", "mips64", "mips64le", "mipsle", "ppc64", "ppc64le", "s390x"},
		CompressBinary:     "tar",
		CompressParameters: []string{"-cf"},
		CompressExtension:  ".tar.gz",
	},
	{
		OperatingSystem:    "darwin",
		Architectures:      []string{"386", "amd64", "arm", "arm64"},
		CompressBinary:     "tar",
		CompressParameters: []string{"-cf"},
		CompressExtension:  ".tar.gz",
	},
	{
		OperatingSystem:    "nacl",
		Architectures:      []string{"386", "amd64p32", "arm"},
		CompressBinary:     "tar",
		CompressParameters: []string{"-cf"},
		CompressExtension:  ".tar.gz",
	},
	{
		OperatingSystem:    "netbsd",
		Architectures:      []string{"386", "amd64", "arm"},
		CompressBinary:     "tar",
		CompressParameters: []string{"-cf"},
		CompressExtension:  ".tar.gz",
	},
	{
		OperatingSystem:    "openbsd",
		Architectures:      []string{"386", "amd64", "arm"},
		CompressBinary:     "tar",
		CompressParameters: []string{"-cf"},
		CompressExtension:  ".tar.gz",
	},
	{
		OperatingSystem:    "plan9",
		Architectures:      []string{"386", "amd64", "arm"},
		CompressBinary:     "tar",
		CompressParameters: []string{"-cf"},
		CompressExtension:  ".tar.gz",
	},
	{
		OperatingSystem:    "solaris",
		Architectures:      []string{"amd64"},
		CompressBinary:     "tar",
		CompressParameters: []string{"-cf"},
		CompressExtension:  ".tar.gz",
	},
	{
		OperatingSystem:    "dragonfly",
		Architectures:      []string{"amd64"},
		CompressBinary:     "tar",
		CompressParameters: []string{"-cf"},
		CompressExtension:  ".tar.gz",
	},
	{
		OperatingSystem:    "windows",
		Architectures:      []string{"386", "amd64"},
		CompressBinary:     "zip",
		CompressParameters: []string{},
		CompressExtension:  ".zip",
		Extension:          ".exe",
	},
}
