// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	pkgcloud "github.com/edwarnicke/pkgcloud/pkgcloudlib"
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all <user/repo>",
	Short: "List all the packages in a repo",
	Long:  `List all the packages in a repo`,
	Run: func(cmd *cobra.Command, args []string) {
		repo := args[0]
		client, err := pkgcloud.NewClient("")
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}
		t := template.Must(template.New("package-tmpl").Parse(allTemplateString))

		next := func() (*pkgcloud.PaginatedPackages, error) {
			return client.PaginatedAll(repo)
		}
		var packages []*pkgcloud.Package
		for next != nil {
			paginatedPackages, err := next()
			if err != nil {
				log.Fatalf("pagination error: %s\n", err)
			}
			packages = append(packages, paginatedPackages.Packages...)
			// log.Printf("len(packages)(+%d): %d\n", len(paginatedPackages.Packages), len(packages))
			for _, p := range paginatedPackages.Packages {
				pack := &Package{Package: p}
				t.Execute(os.Stdout, pack)
			}
			next = paginatedPackages.Next
		}
	},
	Args:             cobra.ExactArgs(1),
	TraverseChildren: true,
}

var allTemplateString string

func init() {
	allCmd.Flags().StringVarP(&allTemplateString, "template", "t", "{{.PackageHTMLURL}}\n", "Golang text template for output")
}

// Package - wraps pkgcloud.Package in order to allow adding 'convenience' method
type Package struct {
	*pkgcloud.Package
}

// Promote - promote Package to repo
func (p *Package) Promote(repo string) string {
	if !DryRun {
		client, err := pkgcloud.NewClient("")
		if err != nil {
			log.Fatalf("Error Promoting to %s : %s : %s", repo, p.PromoteURL, err)
		}
		err = client.Promote(p.Package, repo)
		if err != nil {
			log.Fatalf("Error Promoting to %s : %s : %s", repo, p.PromoteURL, err)
		}
		return fmt.Sprintf("Promoted to %s : %s", repo, p.PromoteURL)
	}
	return fmt.Sprintf("Dry Run for Promoting to %s : %s", repo, p.PromoteURL)
}

// DaysOld - Number of days old the Package is
func (p *Package) DaysOld() int {
	return int(time.Since(p.CreatedAt).Hours() / 24)
}

// Destroy - destroy the package referenced by pkgcloud.Package
func (p *Package) Destroy() string {
	if !DryRun {
		client, err := pkgcloud.NewClient("")
		if err != nil {
			log.Fatalf("Error when trying to Destroy %s : %s", p.PackageHTMLURL, err)
		}
		err = client.DestroyFromPackage(p.Package)
		if err != nil {
			log.Fatalf("Error when trying to Destroy %s : %s", p.PackageHTMLURL, err)
		}
		return fmt.Sprintf("Destroying %s", p.PackageHTMLURL)
	}
	return fmt.Sprintf("Dry Run for Destroying %s", p.PackageHTMLURL)
}
