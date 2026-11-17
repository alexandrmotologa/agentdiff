package ast

import (
	"bufio"
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectLanguage infers programming language from file extension.
func DetectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	default:
		return "generic"
	}
}

// ParseSource parses the code content based on detected or specified language.
func ParseSource(path, content string) (*ParsedFile, error) {
	lang := DetectLanguage(path)
	switch lang {
	case "go":
		return ParseGo(content)
	case "python":
		return ParsePython(content)
	case "typescript", "javascript":
		return ParseTypeScript(content)
	default:
		return ParseGeneric(content)
	}
}

// ParseGo parses Go source using the native Go standard library AST parser.
func ParseGo(content string) (*ParsedFile, error) {
	fset := gotoken.NewFileSet()
	f, err := goparser.ParseFile(fset, "", content, goparser.ParseComments)
	lines := strings.Split(content, "\n")

	pf := &ParsedFile{
		Language:   "go",
		TotalLines: len(lines),
		Nodes:      make([]SemanticNode, 0),
	}

	if err != nil {
		// Fallback to generic line parser if code has partial syntax errors
		return ParseGeneric(content)
	}

	// Capture comments
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			start := fset.Position(c.Pos()).Line
			end := fset.Position(c.End()).Line
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindComment,
				Identifier: "comment",
				LineStart:  start,
				LineEnd:    end,
				Content:    c.Text,
			})
		}
	}

	goast.Inspect(f, func(n goast.Node) bool {
		if n == nil {
			return true
		}

		startPos := fset.Position(n.Pos())
		endPos := fset.Position(n.End())
		start := startPos.Line
		end := endPos.Line

		switch node := n.(type) {
		case *goast.FuncDecl:
			name := node.Name.Name
			if node.Recv != nil && len(node.Recv.List) > 0 {
				name = "method:" + name
			}
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindFuncDecl,
				Identifier: name,
				LineStart:  start,
				LineEnd:    end,
			})
		case *goast.GenDecl:
			if node.Tok == gotoken.TYPE {
				for _, spec := range node.Specs {
					if ts, ok := spec.(*goast.TypeSpec); ok {
						pf.Nodes = append(pf.Nodes, SemanticNode{
							Kind:       KindTypeDecl,
							Identifier: ts.Name.Name,
							LineStart:  start,
							LineEnd:    end,
						})
					}
				}
			} else if node.Tok == gotoken.IMPORT {
				pf.Nodes = append(pf.Nodes, SemanticNode{
					Kind:       KindImportDecl,
					Identifier: "import",
					LineStart:  start,
					LineEnd:    end,
				})
			}
		case *goast.IfStmt:
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindIfStmt,
				Identifier: "if",
				LineStart:  start,
				LineEnd:    end,
			})
		case *goast.ForStmt, *goast.RangeStmt:
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindForStmt,
				Identifier: "for",
				LineStart:  start,
				LineEnd:    end,
			})
		case *goast.AssignStmt:
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindAssignStmt,
				Identifier: "assign",
				LineStart:  start,
				LineEnd:    end,
			})
		case *goast.ReturnStmt:
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindReturnStmt,
				Identifier: "return",
				LineStart:  start,
				LineEnd:    end,
			})
		case *goast.CallExpr:
			name := "call"
			if ident, ok := node.Fun.(*goast.Ident); ok {
				name = ident.Name
			} else if sel, ok := node.Fun.(*goast.SelectorExpr); ok {
				if id, ok := sel.X.(*goast.Ident); ok {
					name = id.Name + "." + sel.Sel.Name
				}
			}
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindCallExpr,
				Identifier: name,
				LineStart:  start,
				LineEnd:    end,
			})
		}
		return true
	})

	return pf, nil
}

var (
	pyDefRegex    = regexp.MustCompile(`^\s*def\s+([a-zA-Z0-9_]+)\s*\(`)
	pyClassRegex  = regexp.MustCompile(`^\s*class\s+([a-zA-Z0-9_]+)`)
	pyIfRegex     = regexp.MustCompile(`^\s*(if|elif|else)\s*.*:`)
	pyForRegex    = regexp.MustCompile(`^\s*(for|while)\s+.*:`)
	pyImportRegex = regexp.MustCompile(`^\s*(import|from)\s+`)
	pyReturnRegex = regexp.MustCompile(`^\s*return(\s+.*)?$`)
)

// ParsePython parses Python code using structural statement and block detection.
func ParsePython(content string) (*ParsedFile, error) {
	lines := strings.Split(content, "\n")
	pf := &ParsedFile{
		Language:   "python",
		TotalLines: len(lines),
		Nodes:      make([]SemanticNode, 0),
	}

	for idx, line := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindComment,
				Identifier: "comment",
				LineStart:  lineNum,
				LineEnd:    lineNum,
				Content:    trimmed,
			})
			continue
		}

		if match := pyDefRegex.FindStringSubmatch(line); len(match) > 1 {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindFuncDecl,
				Identifier: match[1],
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if match := pyClassRegex.FindStringSubmatch(line); len(match) > 1 {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindTypeDecl,
				Identifier: match[1],
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if pyIfRegex.MatchString(line) {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindIfStmt,
				Identifier: "if",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if pyForRegex.MatchString(line) {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindForStmt,
				Identifier: "loop",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if pyImportRegex.MatchString(line) {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindImportDecl,
				Identifier: "import",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if pyReturnRegex.MatchString(line) {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindReturnStmt,
				Identifier: "return",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if strings.Contains(line, "=") && !strings.Contains(line, "==") {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindAssignStmt,
				Identifier: "assign",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		}
	}

	return pf, nil
}

var (
	tsFuncRegex   = regexp.MustCompile(`^\s*(export\s+)?(async\s+)?function\s+([a-zA-Z0-9_]+)`)
	tsArrowRegex  = regexp.MustCompile(`^\s*(export\s+)?(const|let|var)\s+([a-zA-Z0-9_]+)\s*=\s*(async\s*)?\(.*?\)\s*=>`)
	tsTypeRegex   = regexp.MustCompile(`^\s*(export\s+)?(type|interface|class)\s+([a-zA-Z0-9_]+)`)
	tsIfRegex     = regexp.MustCompile(`^\s*(if|else\s+if|else)\s*\(?`)
	tsForRegex    = regexp.MustCompile(`^\s*(for|while)\s*\(?`)
	tsImportRegex = regexp.MustCompile(`^\s*import\s+`)
	tsReturnRegex = regexp.MustCompile(`^\s*return(\s+.*)?$`)
)

// ParseTypeScript parses TypeScript or JavaScript syntax constructs.
func ParseTypeScript(content string) (*ParsedFile, error) {
	lines := strings.Split(content, "\n")
	pf := &ParsedFile{
		Language:   "typescript",
		TotalLines: len(lines),
		Nodes:      make([]SemanticNode, 0),
	}

	for idx, line := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindComment,
				Identifier: "comment",
				LineStart:  lineNum,
				LineEnd:    lineNum,
				Content:    trimmed,
			})
			continue
		}

		if match := tsFuncRegex.FindStringSubmatch(line); len(match) > 3 && match[3] != "" {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindFuncDecl,
				Identifier: match[3],
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if match := tsArrowRegex.FindStringSubmatch(line); len(match) > 3 && match[3] != "" {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindFuncDecl,
				Identifier: match[3],
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if match := tsTypeRegex.FindStringSubmatch(line); len(match) > 3 && match[3] != "" {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindTypeDecl,
				Identifier: match[3],
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if tsIfRegex.MatchString(line) {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindIfStmt,
				Identifier: "if",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if tsForRegex.MatchString(line) {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindForStmt,
				Identifier: "loop",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if tsImportRegex.MatchString(line) {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindImportDecl,
				Identifier: "import",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if tsReturnRegex.MatchString(line) {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindReturnStmt,
				Identifier: "return",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else if strings.Contains(line, "=") && !strings.Contains(line, "==") {
			pf.Nodes = append(pf.Nodes, SemanticNode{
				Kind:       KindAssignStmt,
				Identifier: "assign",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		}
	}

	return pf, nil
}

// ParseGeneric is a fallback for unsupported file extensions.
func ParseGeneric(content string) (*ParsedFile, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	nodes := make([]SemanticNode, 0)
	for scanner.Scan() {
		lineNum++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "#") || strings.HasPrefix(text, "//") {
			nodes = append(nodes, SemanticNode{
				Kind:       KindComment,
				Identifier: "comment",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		} else {
			nodes = append(nodes, SemanticNode{
				Kind:       KindGeneric,
				Identifier: "statement",
				LineStart:  lineNum,
				LineEnd:    lineNum,
			})
		}
	}
	return &ParsedFile{
		Language:   "generic",
		TotalLines: lineNum,
		Nodes:      nodes,
	}, nil
}
