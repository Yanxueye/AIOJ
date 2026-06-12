package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OIWikiPage represents a crawled OI-Wiki page
type OIWikiPage struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Category string            `json:"category"`
	URL      string            `json:"url"`
	Metadata map[string]string `json:"metadata"`
}

var pagesToCrawl = []struct {
	Path     string
	Category string
	ID       string
	Title    string
}{
	// 基础算法
	{"docs/basic/binary.md", "基础算法", "oiwiki-binary-search", "二分查找"},
	{"docs/basic/greedy.md", "基础算法", "oiwiki-greedy", "贪心"},
	{"docs/basic/merge-sort.md", "基础算法", "oiwiki-merge-sort", "归并排序"},
	{"docs/basic/quick-sort.md", "基础算法", "oiwiki-quick-sort", "快速排序"},
	{"docs/basic/bubble-sort.md", "基础算法", "oiwiki-bubble-sort", "冒泡排序"},
	{"docs/basic/insertion-sort.md", "基础算法", "oiwiki-insertion-sort", "插入排序"},
	{"docs/basic/selection-sort.md", "基础算法", "oiwiki-selection-sort", "选择排序"},
	{"docs/basic/heap-sort.md", "基础算法", "oiwiki-heap-sort", "堆排序"},
	{"docs/basic/counting-sort.md", "基础算法", "oiwiki-counting-sort", "计数排序"},
	{"docs/basic/radix-sort.md", "基础算法", "oiwiki-radix-sort", "基数排序"},
	{"docs/basic/bucket-sort.md", "基础算法", "oiwiki-bucket-sort", "桶排序"},

	// 动态规划
	{"docs/dp/knapsack.md", "动态规划", "oiwiki-dp-knapsack", "背包DP"},
	{"docs/dp/interval.md", "动态规划", "oiwiki-dp-interval", "区间DP"},
	{"docs/dp/tree.md", "动态规划", "oiwiki-dp-tree", "树形DP"},
	{"docs/dp/state.md", "动态规划", "oiwiki-dp-state", "状压DP"},
	{"docs/dp/number.md", "动态规划", "oiwiki-dp-number", "数位DP"},
	{"docs/dp/memo.md", "动态规划", "oiwiki-dp-memo", "记忆化搜索"},

	// 图论
	{"docs/graph/dfs.md", "图论", "oiwiki-graph-dfs", "DFS"},
	{"docs/graph/bfs.md", "图论", "oiwiki-graph-bfs", "BFS"},
	{"docs/graph/shortest-path.md", "图论", "oiwiki-graph-shortest-path", "最短路径"},
	{"docs/graph/mst.md", "图论", "oiwiki-graph-mst", "最小生成树"},
	{"docs/graph/topo.md", "图论", "oiwiki-graph-topo", "拓扑排序"},
	{"docs/graph/2-sat.md", "图论", "oiwiki-graph-2-sat", "2-SAT"},

	// 数据结构
	{"docs/ds/stack.md", "数据结构", "oiwiki-ds-stack", "栈"},
	{"docs/ds/heap.md", "数据结构", "oiwiki-ds-heap", "堆"},
	{"docs/ds/dsu.md", "数据结构", "oiwiki-ds-dsu", "并查集"},
	{"docs/ds/seg.md", "数据结构", "oiwiki-ds-seg", "线段树"},
	{"docs/ds/fenwick.md", "数据结构", "oiwiki-ds-fenwick", "树状数组"},
	{"docs/ds/bst.md", "数据结构", "oiwiki-ds-bst", "平衡树"},

	// 字符串
	{"docs/string/kmp.md", "字符串", "oiwiki-string-kmp", "KMP"},
	{"docs/string/hash.md", "字符串", "oiwiki-string-hash", "字符串哈希"},
	{"docs/string/manacher.md", "字符串", "oiwiki-string-manacher", "Manacher"},
	{"docs/string/sam.md", "字符串", "oiwiki-string-sam", "后缀自动机"},
	{"docs/string/sa.md", "字符串", "oiwiki-string-sa", "后缀数组"},

	// 数学
	{"docs/math/number-theory/basic.md", "数学", "oiwiki-math-number-basic", "数论基础"},
	{"docs/math/number-theory/prime.md", "数学", "oiwiki-math-prime", "质数"},
	{"docs/math/number-theory/gcd.md", "数学", "oiwiki-math-gcd", "GCD与LCM"},
	{"docs/math/number-theory/inverse.md", "数学", "oiwiki-math-inverse", "逆元"},
	{"docs/math/combinatorics/combination.md", "数学", "oiwiki-math-combination", "组合数"},
	{"docs/math/combinatorics/catalan.md", "数学", "oiwiki-math-catalan", "卡特兰数"},
	{"docs/math/poly/ntt.md", "数学", "oiwiki-math-ntt", "NTT"},
	{"docs/math/bit.md", "数学", "oiwiki-math-bit", "位运算"},

	// 搜索
	{"docs/search/backtracking.md", "搜索", "oiwiki-search-backtracking", "回溯"},
	{"docs/search/bfs.md", "搜索", "oiwiki-search-bfs", "BFS"},
	{"docs/search/astar.md", "搜索", "oiwiki-search-astar", "A*"},
	{"docs/search/iterative.md", "搜索", "oiwiki-search-iterative", "迭代加深"},
	{"docs/search/bidirectional.md", "搜索", "oiwiki-search-bidirectional", "双向BFS"},
	{"docs/search/heuristic.md", "搜索", "oiwiki-search-heuristic", "启发式搜索"},

	// 计算几何
	{"docs/geometry/convex-hull.md", "计算几何", "oiwiki-geo-convex-hull", "凸包"},
	{"docs/geometry/half-plane.md", "计算几何", "oiwiki-geo-half-plane", "半平面交"},
	{"docs/geometry/nearest-points.md", "计算几何", "oiwiki-geo-nearest-points", "最近点对"},
	{"docs/geometry/rotating-calipers.md", "计算几何", "oiwiki-geo-rotating-calipers", "旋转卡壳"},
}

func main() {
	fmt.Println("=== OI-Wiki Crawler ===")
	fmt.Printf("Crawling %d pages...\n\n", len(pagesToCrawl))

	baseURL := "https://raw.githubusercontent.com/OI-wiki/OI-wiki/master/"

	// Skip SSL verification for environments with certificate issues
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: transport}

	var pages []OIWikiPage
	success := 0
	failed := 0

	for i, p := range pagesToCrawl {
		url := baseURL + p.Path
		fmt.Printf("[%d/%d] Fetching %s ... ", i+1, len(pagesToCrawl), p.Title)

		content, err := fetchMarkdown(client, url)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			failed++
			continue
		}

		cleaned := cleanMarkdown(content)

		oiURL := fmt.Sprintf("https://oi-wiki.org/%s", strings.TrimPrefix(strings.TrimSuffix(p.Path, ".md"), "docs/"))
		page := OIWikiPage{
			ID:       p.ID,
			Title:    p.Title,
			Content:  cleaned,
			Category: p.Category,
			URL:      oiURL,
			Metadata: map[string]string{
				"topic":    p.Title,
				"category": p.Category,
				"url":      oiURL,
			},
		}
		pages = append(pages, page)
		success++
		fmt.Printf("OK (%d bytes)\n", len(cleaned))

		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("\n=== Results: %d success, %d failed ===\n", success, failed)

	// Create output directory
	outputDir := "oiwiki_docs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Write each page as a markdown file
	for _, page := range pages {
		filename := fmt.Sprintf("%s/%s.md", outputDir, page.ID)
		header := fmt.Sprintf("---\ntitle: \"%s\"\ncategory: \"%s\"\nurl: \"%s\"\nid: \"%s\"\n---\n\n",
			page.Title, page.Category, page.URL, page.ID)
		content := header + page.Content
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
		}
	}

	// Write JSON index
	output := struct {
		Pages     []OIWikiPage `json:"pages"`
		CrawledAt string       `json:"crawledAt"`
		Count     int          `json:"count"`
	}{
		Pages:     pages,
		CrawledAt: time.Now().Format(time.RFC3339),
		Count:     len(pages),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	indexFile := fmt.Sprintf("%s/index.json", outputDir)
	if err := os.WriteFile(indexFile, data, 0644); err != nil {
		fmt.Printf("Error writing index: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Output written to %s/ (%d markdown files + index.json)\n", outputDir, len(pages))
}

func fetchMarkdown(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}

	return string(body), nil
}

func cleanMarkdown(content string) string {
	lines := strings.Split(content, "\n")
	var cleaned []string
	inFrontMatter := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip YAML front matter
		if trimmed == "---" {
			if len(cleaned) == 0 || inFrontMatter {
				inFrontMatter = !inFrontMatter
				continue
			}
		}
		if inFrontMatter {
			continue
		}

		// Skip empty lines at start
		if len(cleaned) == 0 && trimmed == "" {
			continue
		}

		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}
