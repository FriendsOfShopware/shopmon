package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/friendsofshopware/shopmon/api"

func TestHTTPTransportDoesNotImportInfrastructure(t *testing.T) {
	root := apiRoot(t)
	forbidden := []string{
		modulePath + "/internal/config",
		modulePath + "/internal/database",
		modulePath + "/internal/jobs",
		modulePath + "/internal/mail",
		modulePath + "/internal/storage",
		"github.com/jackc/pgx",
		"github.com/redis/go-redis",
		"github.com/shyim/go-queue",
	}

	for _, directory := range []string{
		"internal/auth",
		"internal/handler",
		"internal/middleware",
	} {
		assertNoForbiddenImports(t, root, directory, false, forbidden)
	}
}

func TestCapabilityCoreDoesNotImportTransportOrPersistence(t *testing.T) {
	root := apiRoot(t)
	forbidden := []string{
		modulePath + "/internal/api",
		modulePath + "/internal/auth",
		modulePath + "/internal/authapi",
		modulePath + "/internal/config",
		modulePath + "/internal/database",
		modulePath + "/internal/handler",
		modulePath + "/internal/jobs",
		modulePath + "/internal/mail",
		modulePath + "/internal/readmodel",
		modulePath + "/internal/storage",
		"github.com/jackc/pgx",
		"github.com/redis/go-redis",
		"github.com/shyim/go-queue",
	}

	for _, directory := range []string{
		"internal/access",
		"internal/audit",
		"internal/catalog",
		"internal/deployment",
		"internal/identity",
		"internal/monitoring",
		"internal/notification",
		"internal/organization",
		"internal/packagesmirror",
		"internal/suppression",
	} {
		// Only inspect files in the capability root. Infrastructure adapters and
		// specialized worker implementations intentionally live below it.
		assertNoForbiddenImports(t, root, directory, false, forbidden)
	}
}

func TestCapabilitiesDoNotImportHTTPTransport(t *testing.T) {
	root := apiRoot(t)
	forbidden := []string{
		modulePath + "/internal/auth",
		modulePath + "/internal/handler",
	}

	for _, directory := range []string{
		"internal/access",
		"internal/audit",
		"internal/catalog",
		"internal/deployment",
		"internal/identity",
		"internal/maintenance",
		"internal/monitoring",
		"internal/notification",
		"internal/organization",
		"internal/packagesmirror",
		"internal/readmodel",
		"internal/suppression",
	} {
		assertNoForbiddenImports(t, root, directory, true, forbidden)
	}
}

func apiRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve API root: %v", err)
	}
	return root
}

func assertNoForbiddenImports(t *testing.T, root, relativeDirectory string, recursive bool, forbidden []string) {
	t.Helper()

	directory := filepath.Join(root, relativeDirectory)
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != directory && !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range parsed.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			for _, prefix := range forbidden {
				if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
					relativePath, relErr := filepath.Rel(root, path)
					if relErr != nil {
						return relErr
					}
					t.Errorf("%s imports forbidden dependency %s", relativePath, importPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect %s: %v", relativeDirectory, err)
	}
}
