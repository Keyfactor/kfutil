// Package main Copyright 2023 Keyfactor
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	_ "embed"

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
