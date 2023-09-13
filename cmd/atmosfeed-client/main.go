package main

import "github.com/pojntfx/atmosfeed/cmd/atmosfeed-client/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
