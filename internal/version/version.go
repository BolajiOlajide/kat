package version

// Use `-ldflags "-X github.com/BolajiOlajide/kat/internal/version.version=x.x.x"` to set this at Build time.
var version = "dev"

// Version returns the current version of the application.
func Version() string {
	return version
}

func IsDev() bool {
	return version == "dev"
}
