package pipewire

// Version is the semantic version of the pipewire-go library.
const Version = "0.1.0"

// VersionMajor is the major version number.
const VersionMajor = 0

// VersionMinor is the minor version number.
const VersionMinor = 1

// VersionPatch is the patch version number.
const VersionPatch = 0

// Prerelease indicates if this is a pre-release version.
const Prerelease = ""

// Metadata contains additional version metadata.
const Metadata = ""

// String returns the full version string.
func String() string {
	return Version
}
