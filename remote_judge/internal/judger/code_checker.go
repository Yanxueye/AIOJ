package judger

import "strings"

// ac is an Aho-Corasick automaton for multi-pattern matching.
type ac struct {
	trie    []map[byte]int // goto transitions: node → byte → child
	fail    []int           // failure links
	outputs [][]string      // patterns matched at each node (merged via fail links)
}

// newAC builds an Aho-Corasick automaton from the given patterns.
func newAC(patterns []string) *ac {
	a := &ac{
		trie:    []map[byte]int{nil}, // root at index 0
		fail:    []int{0},
		outputs: [][]string{nil},
	}
	a.trie[0] = make(map[byte]int)

	for _, p := range patterns {
		if p == "" {
			continue
		}
		a.insert(p)
	}
	a.buildFail()
	return a
}

// insert adds a pattern into the trie.
func (a *ac) insert(pattern string) {
	node := 0
	for i := 0; i < len(pattern); i++ {
		b := pattern[i]
		if next, ok := a.trie[node][b]; ok {
			node = next
		} else {
			newNode := len(a.trie)
			a.trie = append(a.trie, make(map[byte]int))
			a.fail = append(a.fail, 0)
			a.outputs = append(a.outputs, nil)
			a.trie[node][b] = newNode
			node = newNode
		}
	}
	a.outputs[node] = append(a.outputs[node], pattern)
}

// buildFail builds failure links via BFS and merges outputs.
func (a *ac) buildFail() {
	queue := make([]int, 0)

	// 初始化深度为 1 的节点的 failure links。
	for _, child := range a.trie[0] {
		a.fail[child] = 0
		queue = append(queue, child)
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		for b, child := range a.trie[node] {
			// 查找此子节点的 failure link。
			f := a.fail[node]
			for f != 0 {
				if next, ok := a.trie[f][b]; ok {
					f = next
					break
				}
				f = a.fail[f]
			}
			if f == 0 {
				if next, ok := a.trie[0][b]; ok {
					a.fail[child] = next
				} else {
					a.fail[child] = 0
				}
			} else {
				a.fail[child] = f
			}

			// 合并 failure link 节点的输出模式。
			failNode := a.fail[child]
			if len(a.outputs[failNode]) > 0 {
				a.outputs[child] = append(a.outputs[child], a.outputs[failNode]...)
			}

			queue = append(queue, child)
		}
	}
}

// search scans text and returns all matched patterns (deduplicated).
func (a *ac) search(text string) []string {
	node := 0
	seen := make(map[string]bool)
	var matches []string

	for i := 0; i < len(text); i++ {
		b := text[i]

		// 沿 failure links 回溯，直到找到 goto 转移或到达根节点。
		for node != 0 {
			if next, ok := a.trie[node][b]; ok {
				node = next
				break
			}
			node = a.fail[node]
		}
		if node == 0 {
			if next, ok := a.trie[0][b]; ok {
				node = next
			}
		}

		// 收集当前节点匹配的所有模式。
		for _, p := range a.outputs[node] {
			if !seen[p] {
				seen[p] = true
				matches = append(matches, p)
			}
		}
	}
	return matches
}

// 按语言区分的黑名单。
var blacklists = map[string][]string{
	"cpp17": {
		"system(", "popen(", "execvp(", "execve(", "execlp(",
		"fork(", "socket(", "connect(", "ptrace(", "dlopen(",
		"setuid(", "/proc/", "/sys/", "__asm__", "<unistd.h>",
		"<windows.h>",
	},
	"python3.11": {
		"os.system", "os.popen", "subprocess", "__import__(",
		"eval(", "exec(", "compile(", "ctypes", "pdb",
		"builtins.", "/proc/", "/sys/",
	},
	"go1.22": {
		"os/exec", "syscall.", "net.Dial", "net.Listen",
		"os.StartProcess", "unsafe.", "C.", "/proc/", "/sys/",
	},
}

// ValidateCode 检查代码是否包含黑名单模式。
// 若代码合法返回 (true, "")；若发现禁止模式则返回 (false, reason)。
func ValidateCode(code, language string) (bool, string) {
	patterns, ok := blacklists[language]
	if !ok || len(patterns) == 0 {
		return true, ""
	}

	a := newAC(patterns)
	matches := a.search(code)
	if len(matches) == 0 {
		return true, ""
	}
	return false, "forbidden pattern: " + strings.Join(matches, ", ")
}
