package models

// DefaultTemplates returns the standard C++/Python/Go starter templates.
// This is the single source of truth — seed.go, mysql.go, and handler/problem.go
// all reference this function instead of duplicating the template strings.
func DefaultTemplates() []ProblemTemplate {
	return []ProblemTemplate{
		{Language: "cpp", Code: "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    return 0;\n}\n"},
		{Language: "python", Code: "import sys\ninput = sys.stdin.readline\n\ndef solve():\n    pass\n\nsolve()\n"},
		{Language: "go", Code: "package main\n\nfunc main() {\n}\n"},
	}
}
