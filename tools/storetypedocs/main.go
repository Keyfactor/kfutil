// Copyright 2026 Keyfactor
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
	"flag"
	"fmt"
	"os"

	"kfutil/internal/docgen/storetypedocs"
)

func main() {
	sourcePath := flag.String("source", storetypedocs.DefaultStoreTypesSource, "path to store_types.json")
	pamPath := flag.String("pam-source", storetypedocs.DefaultPAMTypesSource, "path to pam_types.json")
	outputDir := flag.String("out", storetypedocs.DefaultOutputDir, "output directory for generated docs")
	flag.Parse()

	if err := storetypedocs.Generate(*sourcePath, *pamPath, *outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
