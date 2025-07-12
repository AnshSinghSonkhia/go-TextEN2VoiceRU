package main

import (
	"os"

	"github.com/AnshSinghSonkhia/go-TextEN2VoiceRU/cmd"
)

func main() {
	cli := &cmd.CLI{
		ErrStream: os.Stderr,
	}

	os.Exit(cli.Run(os.Args[1:]))
}
