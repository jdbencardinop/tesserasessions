// Package buildinfo exposes the resolved tss version string.
//
// Resolution order:
//
//  1. ldflags-injected value from the Makefile:
//
//     -ldflags "-X github.com/jdbencardinop/tesserasessions/internal/buildinfo.Version=v0.1.0"
//
//  2. runtime/debug.ReadBuildInfo(), populated by module-aware installs.
//
//  3. The literal "dev" for local go run/go build workflows.
package buildinfo

import "runtime/debug"

// Version is overridden by linker flags in Makefile builds.
var Version = "dev"

// String returns the best available version.
func String() string {
	return resolve(Version, debug.ReadBuildInfo)
}

func resolve(injected string, readInfo func() (*debug.BuildInfo, bool)) string {
	if injected != "" && injected != "dev" {
		return injected
	}
	if info, ok := readInfo(); ok {
		v := info.Main.Version
		if v != "" && v != "(devel)" {
			return v
		}
	}
	if injected == "" {
		return "dev"
	}
	return injected
}
