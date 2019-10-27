/*
 * 
 */

package main

import "log"

var (
	buildVersion = ""
	buildDate    = ""
	commitHash   = ""
)

func main() {
	cli := NewCli()
	err := cli.Init()
	if err != nil {
		log.Fatal(err)
	}
	cli.AddCommands(commands)
	cli.Execute()
}
