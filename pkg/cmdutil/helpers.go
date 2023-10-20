/*
Copyright 2023 The Keyfactor Command Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmdutil

import (
	"fmt"
	"runtime"
)

func PrintError(err error) {
	// Print errors in red
	if err != nil {
		fmt.Printf("\033[31m%s\u001B[0m\n", err)
	}
}

func GetOs() string {
	return runtime.GOOS
}
