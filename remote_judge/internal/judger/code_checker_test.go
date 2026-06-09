package judger

import "testing"

func TestAC_Search(t *testing.T) {
	a := newAC([]string{"he", "she", "his", "hers"})
	matches := a.search("ushers")
	want := map[string]bool{"she": true, "he": true, "hers": true}
	if len(matches) != len(want) {
		t.Fatalf("got %d matches, want %d: %v", len(matches), len(want), matches)
	}
	for _, m := range matches {
		if !want[m] {
			t.Fatalf("unexpected match: %q", m)
		}
	}
}

func TestAC_Empty(t *testing.T) {
	a := newAC(nil)
	matches := a.search("hello world")
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %v", matches)
	}
}

func TestAC_Overlap(t *testing.T) {
	a := newAC([]string{"ab", "bc", "abc"})
	matches := a.search("abc")
	// All three should match (abc contains "ab", "bc", and "abc").
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d: %v", len(matches), matches)
	}
}

func TestAC_SingleMatch(t *testing.T) {
	a := newAC([]string{"system("})
	matches := a.search(`int main() { system("cat flag"); }`)
	if len(matches) != 1 || matches[0] != "system(" {
		t.Fatalf("expected [system(], got %v", matches)
	}
}

func TestAC_NoFalsePositive(t *testing.T) {
	// "ecosystem" should NOT match "system(".
	a := newAC([]string{"system("})
	matches := a.search("ecosystem management")
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %v", matches)
	}
}

func TestValidateCode_Cpp(t *testing.T) {
	tests := []struct {
		name string
		code string
		ok   bool
	}{
		{"clean a+b", "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a+b;}", true},
		{"system call", "#include <cstdlib>\nint main(){system(\"cat flag\");}", false},
		{"fork", "#include <unistd.h>\nint main(){fork();}", false},
		{"socket", "#include <sys/socket.h>\nint main(){socket(2,1,0);}", false},
		{"proc access", "#include <fstream>\nint main(){std::ifstream f(\"/proc/self/mem\");}", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateCode(tt.code, "cpp17")
			if ok != tt.ok {
				t.Fatalf("ValidateCode = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestValidateCode_Python(t *testing.T) {
	tests := []struct {
		name string
		code string
		ok   bool
	}{
		{"clean a+b", "a,b=map(int,input().split())\nprint(a+b)", true},
		{"os.system", "import os\nos.system('cat flag')", false},
		{"eval", "x = eval(input())", false},
		{"subprocess", "import subprocess\nsubprocess.run(['ls'])", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateCode(tt.code, "python3.11")
			if ok != tt.ok {
				t.Fatalf("ValidateCode = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestValidateCode_Go(t *testing.T) {
	tests := []struct {
		name string
		code string
		ok   bool
	}{
		{"clean a+b", "package main\nimport \"fmt\"\nfunc main(){var a,b int;fmt.Scan(&a,&b);fmt.Println(a+b)}", true},
		{"os/exec", "package main\nimport \"os/exec\"\nfunc main(){exec.Command(\"cat\",\"flag\")}", false},
		{"syscall", "package main\nimport \"syscall\"\nfunc main(){syscall.Exec(\"/bin/sh\",nil,nil)}", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateCode(tt.code, "go1.22")
			if ok != tt.ok {
				t.Fatalf("ValidateCode = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestValidateCode_Unknown(t *testing.T) {
	ok, _ := ValidateCode("anything", "rust")
	if !ok {
		t.Fatalf("unknown language should pass (no blacklist)")
	}
}
