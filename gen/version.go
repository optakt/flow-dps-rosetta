// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:generate go run version.go

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"golang.org/x/mod/modfile"
)

const versionFileTemplate = `// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package configuration

const (
	RosettaVersion    = "{{ .RosettaVersion }}"
	NodeVersion       = "{{ .NodeVersion }}"
	MiddlewareVersion = "{{ .MiddlewareVersion }}"
)
`

const (
	rosettaVersion = "1.4.10"

	pathToGoMod            = "../go.mod"
	rosettaVersionFilePath = "../service/configuration/version.go"

	flowModPath = "github.com/onflow/flow-go"
	dpsModPath  = "github.com/optakt/flow-dps"
)

func main() {

	fmt.Println("Using rosetta version", rosettaVersion)

	nodeVersion, err := NodeVersion()
	if err != nil {
		log.Fatalf("could not compute node version: %v", err)
	}

	fmt.Println("Found node version", nodeVersion)

	middlewareVersion, err := MiddlewareVersion()
	if err != nil {
		log.Fatalf("could not compute middleware version: %v", err)
	}

	fmt.Println("Found middleware version", middlewareVersion)

	tmpl := template.Must(template.New("version.go").Parse(versionFileTemplate))

	versionFile, err := os.Create(rosettaVersionFilePath)
	if err != nil {
		log.Fatalf("could not open version file: %v", err)
	}

	args := struct {
		RosettaVersion    string
		NodeVersion       string
		MiddlewareVersion string
	}{
		RosettaVersion:    rosettaVersion,
		NodeVersion:       nodeVersion,
		MiddlewareVersion: middlewareVersion,
	}

	err = tmpl.Execute(versionFile, args)
	if err != nil {
		log.Fatalf("could not execute template: %v", err)
	}
}

// MiddlewareVersion parses the Go.mod file to retrieve the version of the Flow-DPS
// dependency.
func MiddlewareVersion() (string, error) {
	gomod, err := os.ReadFile(pathToGoMod)
	if err != nil {
		return "", fmt.Errorf("could not read go mod file: %w", err)
	}

	modfile, err := modfile.Parse("go.mod", gomod, nil)
	if err != nil {
		return "", fmt.Errorf("could not parse go mod file: %w", err)
	}

	for _, module := range modfile.Require {
		if module.Mod.Path == dpsModPath {
			// Strip leading `v` from the tag if it exists.
			nodeVersion := strings.TrimPrefix(module.Mod.Version, "v")
			return nodeVersion, nil
		}
	}

	return "", fmt.Errorf("could not find github.com/optakt/flow-dps dependency in go mod file")
}

// NodeVersion parses the Go.mod file to retrieve the version of the Flow-go
// dependency.
func NodeVersion() (string, error) {
	gomod, err := os.ReadFile(pathToGoMod)
	if err != nil {
		return "", fmt.Errorf("could not read go mod file: %w", err)
	}

	modfile, err := modfile.Parse("go.mod", gomod, nil)
	if err != nil {
		return "", fmt.Errorf("could not parse go mod file: %w", err)
	}

	for _, module := range modfile.Require {
		if module.Mod.Path == flowModPath {
			// Strip leading `v` from the tag if it exists.
			nodeVersion := strings.TrimPrefix(module.Mod.Version, "v")
			return nodeVersion, nil
		}
	}

	return "", fmt.Errorf("could not find github.com/onflow/flow-go dependency in go mod file")
}
