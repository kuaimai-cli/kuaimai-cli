// Package build holds release metadata injected at compile time.
package build

// Version is the CLI release version (ldflags -X).
var Version = "dev"

// Date is the build date (ldflags -X).
var Date = "unknown"
