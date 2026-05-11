/*
Copyright © 2026 Julian Cantillo <julian@cantillo.dev>
*/
package main

import "cantillo.dev/kidsboard/cmd"

// version is stamped at build time via:
//
//	go build -ldflags="-X main.version=0.1.0"
//
// The Dockerfile and CI pipeline populate this from the release tag.
// Default `dev` signals a local/unstamped build.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
