/*
Implementation of git internal commands in Go language.

This project is part of a learning exercise to implement a subset of "git"
commands. It can be used to create and maintain git objects, such as blobs,
trees, commits, branches and tags.

Code Organization

- "git": internal git objects and related APIs.

- "cmd": Command line parsing and execution.

- "util": Miscellaneous utility APIs.

*/
package main

import (
	"github.com/ssrathi/gogit/cmd"
)

func main() {
	cmd.Execute()
}