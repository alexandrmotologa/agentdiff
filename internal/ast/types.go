package ast

import "github.com/alexandrmotologa/agentdiff/internal/trajectory"

// NodeKind identifies the semantic category of an AST node.
type NodeKind string

const (
	KindFuncDecl   NodeKind = "FuncDecl"
	KindTypeDecl   NodeKind = "TypeDecl"
	KindIfStmt     NodeKind = "IfStmt"
	KindForStmt    NodeKind = "ForStmt"
	KindReturnStmt NodeKind = "ReturnStmt"
	KindAssignStmt NodeKind = "AssignStmt"
	KindCallExpr   NodeKind = "CallExpr"
	KindImportDecl NodeKind = "ImportDecl"
	KindComment    NodeKind = "Comment"
	KindGeneric    NodeKind = "Statement"
)

// SemanticNode represents an individual AST node in a parsed source file.
type SemanticNode struct {
	Kind       NodeKind
	Identifier string
	LineStart  int
	LineEnd    int
	Content    string
}

// ParsedFile contains extracted semantic nodes and line metadata.
type ParsedFile struct {
	Language   string
	TotalLines int
	Nodes      []SemanticNode
}

// ASTDiffSummary reports the structural difference between before and after versions.
type ASTDiffSummary struct {
	TotalLinesDiff     int
	LinesAdded         int
	LinesRemoved       int
	ASTNodesModified   int
	WhitespaceLines    int
	CommentLines       int
	SurgicalScore      float64
	CosmeticChurnRatio float64
	AffectedNodes      []trajectory.ASTNodeRef
}
