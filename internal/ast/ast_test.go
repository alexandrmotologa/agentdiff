package ast

import (
	"testing"
)

func TestParseGoAndCompare(t *testing.T) {
	origGo := `package main

import "fmt"

func Process(val int) int {
	return val * 2
}
`

	modGoSurgical := `package main

import "fmt"

func Process(val int) int {
	if val < 0 {
		return 0
	}
	return val * 2
}
`

	modGoChurn := `package main

import "fmt"

// Process handles value calculation
// Added extra comments
// More comments
func Process(val int) int {



	return val * 2
}
`

	// Surgical edit
	summarySurgical, err := CompareSourceFiles("main.go", origGo, modGoSurgical)
	if err != nil {
		t.Fatalf("CompareSourceFiles failed: %v", err)
	}

	if summarySurgical.TotalLinesDiff != 3 {
		t.Errorf("expected 3 lines diff, got %d", summarySurgical.TotalLinesDiff)
	}
	if summarySurgical.SurgicalScore < 0.25 {
		t.Errorf("expected surgical score >= 0.25, got %f", summarySurgical.SurgicalScore)
	}
	if summarySurgical.CosmeticChurnRatio > 0.1 {
		t.Errorf("expected cosmetic churn near 0, got %f", summarySurgical.CosmeticChurnRatio)
	}

	// Churn edit
	summaryChurn, err := CompareSourceFiles("main.go", origGo, modGoChurn)
	if err != nil {
		t.Fatalf("CompareSourceFiles failed: %v", err)
	}
	if summaryChurn.CosmeticChurnRatio < 0.5 {
		t.Errorf("expected high cosmetic churn ratio for comments/whitespace, got %f", summaryChurn.CosmeticChurnRatio)
	}
}

func TestParsePythonAndCompare(t *testing.T) {
	origPy := `def calculate_total(items):
    total = 0
    for item in items:
        total += item.price
    return total
`

	modPy := `def calculate_total(items):
    total = 0
    for item in items:
        if item is not None and item.price > 0:
            total += item.price
    return total
`

	summary, err := CompareSourceFiles("calc.py", origPy, modPy)
	if err != nil {
		t.Fatalf("CompareSourceFiles failed: %v", err)
	}

	if summary.TotalLinesDiff == 0 {
		t.Errorf("expected non-zero lines diff")
	}
	if summary.ASTNodesModified == 0 {
		t.Errorf("expected AST nodes modified in python")
	}
}

func TestParseTypeScriptAndCompare(t *testing.T) {
	origTS := `export interface User {
    id: string;
    name: string;
}

export function formatUser(user: User): string {
    return user.name;
}
`

	modTS := `export interface User {
    id: string;
    name: string;
    email?: string;
}

export function formatUser(user: User): string {
    if (!user.name) {
        return "Unknown";
    }
    return user.name;
}
`

	summary, err := CompareSourceFiles("user.ts", origTS, modTS)
	if err != nil {
		t.Fatalf("CompareSourceFiles failed: %v", err)
	}

	if summary.TotalLinesDiff == 0 {
		t.Errorf("expected lines diff")
	}
	if summary.ASTNodesModified == 0 {
		t.Errorf("expected AST nodes modified in typescript")
	}
}
