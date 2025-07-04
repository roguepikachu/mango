package testmeta

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

// Metadata represents a single test case metadata.
type Metadata struct {
	Name    string
	File    string
	Package string
	Ginkgo  bool
}

// Extract scans the repository for tests and returns their metadata.
func Extract() ([]Metadata, error) {
	var meta []Metadata
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}
		tests, err := parseFile(path)
		if err != nil {
			return err
		}
		meta = append(meta, tests...)
		return nil
	})
	return meta, err
}

func parseFile(path string) ([]Metadata, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	pkg := f.Name.Name
	var meta []Metadata
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if strings.HasPrefix(x.Name.Name, "Test") && x.Recv == nil {
				meta = append(meta, Metadata{Name: x.Name.Name, File: path, Package: pkg})
			}
		case *ast.CallExpr:
			if ident, ok := x.Fun.(*ast.Ident); ok {
				if isGinkgoFunc(ident.Name) {
					if len(x.Args) > 0 {
						if lit, ok := x.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							name, _ := strconv.Unquote(lit.Value)
							meta = append(meta, Metadata{Name: name, File: path, Package: pkg, Ginkgo: true})
						}
					}
				}
			}
		}
		return true
	})
	return meta, nil
}

func isGinkgoFunc(name string) bool {
	switch name {
	case "Describe", "Context", "When", "It", "Specify", "By", "Measure":
		return true
	default:
		return false
	}
}
