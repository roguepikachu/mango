package diff

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Change represents a modified file or symbol.
type Change struct {
	File      string
	Functions []string
}

var hunkRegexp = regexp.MustCompile(`@@ .*\+(\d+)(?:,(\d+))? @@`)

// AnalyzeDiff runs git diff for the given range and returns list of changed files and functions.
func AnalyzeDiff(diffRange string) ([]Change, error) {
	if diffRange == "" {
		diffRange = "HEAD~1"
	}

	cmd := exec.Command("git", "diff", diffRange, "--unified=0")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(out.Bytes()))
	fileChanges := map[string][]int{}
	var currentFile string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				currentFile = strings.TrimPrefix(parts[3], "b/")
			}
			continue
		}
		if strings.HasPrefix(line, "@@") {
			m := hunkRegexp.FindStringSubmatch(line)
			if len(m) >= 2 {
				start, _ := strconv.Atoi(m[1])
				count := 1
				if len(m) > 2 && m[2] != "" {
					count, _ = strconv.Atoi(m[2])
				}
				for i := 0; i < count; i++ {
					fileChanges[currentFile] = append(fileChanges[currentFile], start+i)
				}
			}
		}
	}

	var result []Change
	for file, lines := range fileChanges {
		funcs, err := changedFunctions(file, lines)
		if err != nil {
			return nil, err
		}
		result = append(result, Change{File: file, Functions: funcs})
	}

	return result, nil
}

func changedFunctions(file string, lines []int) ([]string, error) {
	if !strings.HasSuffix(file, ".go") {
		return nil, nil
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		// file might be deleted or moved
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lineSet := make(map[int]struct{}, len(lines))
	for _, l := range lines {
		lineSet[l] = struct{}{}
	}
	var funcs []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		start := fset.Position(fn.Pos()).Line
		end := fset.Position(fn.End()).Line
		for l := range lineSet {
			if l >= start && l <= end {
				funcs = append(funcs, fn.Name.Name)
				break
			}
		}
	}
	return funcs, nil
}
