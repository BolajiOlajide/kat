package version

// Use `-ldflags "-X github.com/BolajiOlajide/kat/internal/version.version=x.x.x"` to set this at Build time.
const version = "dev"

// Version returns the current version of the application.
func Version() string {
	return version
}
