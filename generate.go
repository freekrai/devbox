// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package devbox

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"go.jetpack.io/devbox/debug"
	"go.jetpack.io/devbox/planner/plansdk"
)

//go:embed tmpl/* tmpl/.*
var tmplFS embed.FS

var shellFiles = []string{".gitignore", "shell.nix"}
var buildFiles = []string{".gitignore", "development.nix", "runtime.nix", "Dockerfile", "Dockerfile.dockerignore"}

func generate(rootPath string, plan *plansdk.Plan, files []string) error {
	outPath := filepath.Join(rootPath, ".devbox/gen")

	for _, file := range files {
		err := writeFromTemplate(outPath, plan, file)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	for name, content := range plan.GeneratedFiles {
		filePath := filepath.Join(outPath, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func writeFromTemplate(path string, plan *plansdk.Plan, tmplName string) error {
	embeddedPath := fmt.Sprintf("tmpl/%s.tmpl", tmplName)

	// Should we clear the directory so we start "fresh"?
	outPath := filepath.Join(path, tmplName)
	outDir := filepath.Dir(outPath)
	err := os.MkdirAll(outDir, 0755) // Ensure directory exists.
	if err != nil {
		return errors.WithStack(err)
	}

	f, err := os.Create(outPath)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return errors.WithStack(err)
	}
	t := template.Must(template.New(tmplName+".tmpl").Funcs(templateFuncs).ParseFS(tmplFS, embeddedPath))
	return errors.WithStack(t.Execute(f, plan))
}

func toJSON(a any) string {
	data, _ := json.Marshal(a)
	return string(data)
}

var templateFuncs = template.FuncMap{
	"json":     toJSON,
	"contains": strings.Contains,
	"debug":    debug.IsEnabled,
}
