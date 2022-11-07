/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/spf13/cobra/doc"
	"kfutil/cmd"
)

func main() {
	//var docsFlag bool
	//flag.BoolVar(&docsFlag, "makedocs", false, "Create markdown docs.")
	//flag.Parse()
	//if docsFlag {
	//	docs()
	//	os.Exit(0)
	//}
	cmd.Execute()
}

func docs() {
	doc.GenMarkdownTree(cmd.RootCmd, "./docs")
}
