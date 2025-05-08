package version

// Use `-ldflags "-X github.com/BolajiOlajide/kat/v0/internal/version.version=x.x.x"` to set this at Build time.
var version = "dev"

var MockVersion func() string

// Version returns the current version of the application.
func Version() string {
	if MockVersion != nil {
		return MockVersion()
	}
	return version
}

func IsDev() bool {
	return version == "dev"
}
