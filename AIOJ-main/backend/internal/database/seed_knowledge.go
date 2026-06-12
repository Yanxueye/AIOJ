package database

import (
	"log"

	"github.com/terminaloj/backend/internal/models"
	"gorm.io/gorm"
)

func seedKnowledge(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.KnowledgePoint{}).Count(&count)
	// Check if children exist too (handles partial seed from crash)
	var childCount int64
	conn.Model(&models.KnowledgePoint{}).Where("parent_id IS NOT NULL").Count(&childCount)
	if count > 0 && childCount > 0 {
		return nil
	}
	// If partial seed, clear and re-seed
	if count > 0 {
		conn.Exec("DELETE FROM knowledge_points")
	}

	type kpDef struct {
		Name        string
		Description string
		Category    string
		Parent      string // empty for top-level
		OjWikiURL   string
		Color       string
		Icon        string
	}

	defs := []kpDef{
		// ── 动态规划 ──────────────────────────────────────────────
		{Name: "动态规划", Category: "动态规划", Description: "通过将问题分解为重叠子问题来求解的算法思想", Color: "#3B82F6", Icon: "brain"},
		{Name: "背包DP", Category: "动态规划", Parent: "动态规划", Description: "在容量限制下选择物品以最大化价值", OjWikiURL: "https://oi-wiki.org/dp/knapsack/", Color: "#3B82F6", Icon: "box"},
		{Name: "区间DP", Category: "动态规划", Parent: "动态规划", Description: "在连续区间上进行动态规划", OjWikiURL: "https://oi-wiki.org/dp/interval/", Color: "#3B82F6", Icon: "brackets"},
		{Name: "树形DP", Category: "动态规划", Parent: "动态规划", Description: "在树结构上进行动态规划", OjWikiURL: "https://oi-wiki.org/dp/tree/", Color: "#3B82F6", Icon: "tree"},
		{Name: "数位DP", Category: "动态规划", Parent: "动态规划", Description: "按数字的数位进行状态转移", OjWikiURL: "https://oi-wiki.org/dp/number/", Color: "#3B82F6", Icon: "hash"},
		{Name: "状态压缩DP", Category: "动态规划", Parent: "动态规划", Description: "用二进制整数表示集合状态", OjWikiURL: "https://oi-wiki.org/dp/state/", Color: "#3B82F6", Icon: "binary"},
		{Name: "DP优化", Category: "动态规划", Parent: "动态规划", Description: "利用单调性、凸性等性质优化转移", OjWikiURL: "https://oi-wiki.org/dp/opt/", Color: "#3B82F6", Icon: "zap"},
		{Name: "计数DP", Category: "动态规划", Parent: "动态规划", Description: "统计满足条件的方案数", OjWikiURL: "https://oi-wiki.org/dp/count/", Color: "#3B82F6", Icon: "counter"},
		{Name: "概率DP", Category: "动态规划", Parent: "动态规划", Description: "状态转移涉及概率与期望", OjWikiURL: "https://oi-wiki.org/dp/probability/", Color: "#3B82F6", Icon: "dice"},
		{Name: "博弈论DP", Category: "动态规划", Parent: "动态规划", Description: "博弈局面的SG函数与必胜态分析", OjWikiURL: "https://oi-wiki.org/dp/game/", Color: "#3B82F6", Icon: "gamepad"},

		// ── 图论 ──────────────────────────────────────────────────
		{Name: "图论", Category: "图论", Description: "研究图结构及其算法", Color: "#10B981", Icon: "network"},
		{Name: "最短路", Category: "图论", Parent: "图论", Description: "Dijkstra、Bellman-Ford、SPFA、Floyd 等算法", OjWikiURL: "https://oi-wiki.org/graph/shortest-path/", Color: "#10B981", Icon: "route"},
		{Name: "最小生成树", Category: "图论", Parent: "图论", Description: "Kruskal、Prim 等算法求最小生成树", OjWikiURL: "https://oi-wiki.org/graph/mst/", Color: "#10B981", Icon: "tree"},
		{Name: "网络流", Category: "图论", Parent: "图论", Description: "最大流、最小割、费用流", OjWikiURL: "https://oi-wiki.org/graph/flow/", Color: "#10B981", Icon: "flow"},
		{Name: "二分图", Category: "图论", Parent: "图论", Description: "匈牙利算法、KM 算法", OjWikiURL: "https://oi-wiki.org/graph/bi-graph/", Color: "#10B981", Icon: "layers"},
		{Name: "拓扑排序", Category: "图论", Parent: "图论", Description: "DAG 上的线性排序", OjWikiURL: "https://oi-wiki.org/graph/topo/", Color: "#10B981", Icon: "sort"},
		{Name: "强连通分量", Category: "图论", Parent: "图论", Description: "Tarjan、Kosaraju 算法", OjWikiURL: "https://oi-wiki.org/graph/scc/", Color: "#10B981", Icon: "loop"},
		{Name: "桥和割点", Category: "图论", Parent: "图论", Description: "寻找图中的桥和割点", OjWikiURL: "https://oi-wiki.org/graph/cut/", Color: "#10B981", Icon: "scissors"},
		{Name: "树上问题", Category: "图论", Parent: "图论", Description: "树链剖分、重心、直径等", OjWikiURL: "https://oi-wiki.org/graph/tree-basic/", Color: "#10B981", Icon: "tree-deciduous"},
		{Name: "LCA", Category: "图论", Parent: "图论", Description: "最近公共祖先，倍增/Tarjan/树剖", OjWikiURL: "https://oi-wiki.org/graph/lca/", Color: "#10B981", Icon: "merge"},

		// ── 数据结构 ──────────────────────────────────────────────
		{Name: "数据结构", Category: "数据结构", Description: "组织和存储数据的方式", Color: "#F59E0B", Icon: "database"},
		{Name: "线段树", Category: "数据结构", Parent: "数据结构", Description: "区间查询与修改，支持懒标记", OjWikiURL: "https://oi-wiki.org/ds/seg/", Color: "#F59E0B", Icon: "chart-bar"},
		{Name: "树状数组", Category: "数据结构", Parent: "数据结构", Description: "前缀和查询与单点修改", OjWikiURL: "https://oi-wiki.org/ds/fenwick/", Color: "#F59E0B", Icon: "bar-chart"},
		{Name: "并查集", Category: "数据结构", Parent: "数据结构", Description: "不相交集合的合并与查询", OjWikiURL: "https://oi-wiki.org/ds/dsu/", Color: "#F59E0B", Icon: "link"},
		{Name: "平衡树", Category: "数据结构", Parent: "数据结构", Description: "Treap、Splay、AVL、红黑树等", OjWikiURL: "https://oi-wiki.org/ds/bst/", Color: "#F59E0B", Icon: "scale"},
		{Name: "哈希表", Category: "数据结构", Parent: "数据结构", Description: "基于哈希函数的键值存储", OjWikiURL: "https://oi-wiki.org/ds/hash/", Color: "#F59E0B", Icon: "key"},
		{Name: "栈", Category: "数据结构", Parent: "数据结构", Description: "后进先出的线性结构", OjWikiURL: "https://oi-wiki.org/ds/stack/", Color: "#F59E0B", Icon: "layers"},
		{Name: "队列", Category: "数据结构", Parent: "数据结构", Description: "先进先出的线性结构", OjWikiURL: "https://oi-wiki.org/ds/queue/", Color: "#F59E0B", Icon: "list"},
		{Name: "堆", Category: "数据结构", Parent: "数据结构", Description: "优先队列，二叉堆/斐波那契堆", OjWikiURL: "https://oi-wiki.org/ds/heap/", Color: "#F59E0B", Icon: "arrow-up"},
		{Name: "字典树", Category: "数据结构", Parent: "数据结构", Description: "Trie 树，字符串前缀检索", OjWikiURL: "https://oi-wiki.org/ds/trie/", Color: "#F59E0B", Icon: "book"},
		{Name: "分块", Category: "数据结构", Parent: "数据结构", Description: "将序列分块进行区间操作", OjWikiURL: "https://oi-wiki.org/ds/block-list/", Color: "#F59E0B", Icon: "grid"},

		// ── 数学 ──────────────────────────────────────────────────
		{Name: "数学", Category: "数学", Description: "算法竞赛中的数学知识", Color: "#8B5CF6", Icon: "calculator"},
		{Name: "质数", Category: "数学", Parent: "数学", Description: "素数筛法、质因数分解", OjWikiURL: "https://oi-wiki.org/math/number-theory/prime/", Color: "#8B5CF6", Icon: "hash"},
		{Name: "GCD/LCM", Category: "数学", Parent: "数学", Description: "欧几里得算法与最小公倍数", OjWikiURL: "https://oi-wiki.org/math/number-theory/gcd/", Color: "#8B5CF6", Icon: "divide"},
		{Name: "快速幂", Category: "数学", Parent: "数学", Description: "二进制取模幂运算", OjWikiURL: "https://oi-wiki.org/math/binary-exponentiation/", Color: "#8B5CF6", Icon: "zap"},
		{Name: "组合数学", Category: "数学", Parent: "数学", Description: "排列组合、卡特兰数、斯特林数", OjWikiURL: "https://oi-wiki.org/math/combinatorics/", Color: "#8B5CF6", Icon: "shuffle"},
		{Name: "容斥原理", Category: "数学", Parent: "数学", Description: "集合计数中的容斥公式", OjWikiURL: "https://oi-wiki.org/math/combinatorics/inclusion-exclusion/", Color: "#8B5CF6", Icon: "circle"},
		{Name: "概率期望", Category: "数学", Parent: "数学", Description: "概率论基础与期望计算", OjWikiURL: "https://oi-wiki.org/math/probability/", Color: "#8B5CF6", Icon: "pie-chart"},
		{Name: "矩阵", Category: "数学", Parent: "数学", Description: "矩阵运算与矩阵快速幂", OjWikiURL: "https://oi-wiki.org/math/matrix/", Color: "#8B5CF6", Icon: "grid"},
		{Name: "高斯消元", Category: "数学", Parent: "数学", Description: "线性方程组求解", OjWikiURL: "https://oi-wiki.org/math/gauss/", Color: "#8B5CF6", Icon: "columns"},
		{Name: "莫比乌斯反演", Category: "数学", Parent: "数学", Description: "数论中的莫比乌斯函数与反演公式", OjWikiURL: "https://oi-wiki.org/math/mobius/", Color: "#8B5CF6", Icon: "refresh-cw"},

		// ── 字符串 ────────────────────────────────────────────────
		{Name: "字符串", Category: "字符串", Description: "字符串匹配与处理算法", Color: "#EF4444", Icon: "type"},
		{Name: "KMP", Category: "字符串", Parent: "字符串", Description: "Knuth-Morris-Pratt 单模式匹配", OjWikiURL: "https://oi-wiki.org/string/kmp/", Color: "#EF4444", Icon: "search"},
		{Name: "Trie", Category: "字符串", Parent: "字符串", Description: "字符串前缀树", OjWikiURL: "https://oi-wiki.org/string/trie/", Color: "#EF4444", Icon: "book"},
		{Name: "后缀数组", Category: "字符串", Parent: "字符串", Description: "SA 与 LCP 数组", OjWikiURL: "https://oi-wiki.org/string/sa/", Color: "#EF4444", Icon: "align-left"},
		{Name: "后缀自动机", Category: "字符串", Parent: "字符串", Description: "SAM，线性构建的后缀结构", OjWikiURL: "https://oi-wiki.org/string/sam/", Color: "#EF4444", Icon: "cpu"},
		{Name: "AC自动机", Category: "字符串", Parent: "字符串", Description: "多模式匹配自动机", OjWikiURL: "https://oi-wiki.org/string/ac-automaton/", Color: "#EF4444", Icon: "filter"},
		{Name: "Manacher", Category: "字符串", Parent: "字符串", Description: "线性时间求最长回文子串", OjWikiURL: "https://oi-wiki.org/string/manacher/", Color: "#EF4444", Icon: "repeat"},
		{Name: "哈希", Category: "字符串", Parent: "字符串", Description: "字符串哈希与子串比较", OjWikiURL: "https://oi-wiki.org/string/hash/", Color: "#EF4444", Icon: "hash"},

		// ── 搜索 ──────────────────────────────────────────────────
		{Name: "搜索", Category: "搜索", Description: "系统地探索解空间", Color: "#06B6D4", Icon: "search"},
		{Name: "BFS", Category: "搜索", Parent: "搜索", Description: "广度优先搜索，最短路模型", OjWikiURL: "https://oi-wiki.org/search/bfs/", Color: "#06B6D4", Icon: "layers"},
		{Name: "DFS", Category: "搜索", Parent: "搜索", Description: "深度优先搜索，回溯框架", OjWikiURL: "https://oi-wiki.org/search/dfs/", Color: "#06B6D4", Icon: "git-branch"},
		{Name: "迭代加深", Category: "搜索", Parent: "搜索", Description: "逐步增加搜索深度限制", OjWikiURL: "https://oi-wiki.org/search/iterative/", Color: "#06B6D4", Icon: "chevrons-down"},
		{Name: "IDA*", Category: "搜索", Parent: "搜索", Description: "迭代加深 A* 搜索", OjWikiURL: "https://oi-wiki.org/search/idastar/", Color: "#06B6D4", Icon: "target"},
		{Name: "双向BFS", Category: "搜索", Parent: "搜索", Description: "从起点和终点同时搜索", OjWikiURL: "https://oi-wiki.org/search/bidirectional/", Color: "#06B6D4", Icon: "repeat"},
		{Name: "启发式搜索(A*)", Category: "搜索", Parent: "搜索", Description: "带估价函数的优先搜索", OjWikiURL: "https://oi-wiki.org/search/astar/", Color: "#06B6D4", Icon: "star"},
		{Name: "折半搜索", Category: "搜索", Parent: "搜索", Description: "将搜索空间分成两半分别枚举", OjWikiURL: "https://oi-wiki.org/search/half/", Color: "#06B6D4", Icon: "scissors"},

		// ── 贪心 ──────────────────────────────────────────────────
		{Name: "贪心", Category: "贪心", Description: "每步选择局部最优解", Color: "#F97316", Icon: "trending-up"},
		{Name: "区间贪心", Category: "贪心", Parent: "贪心", Description: "区间调度、不相交区间选择", OjWikiURL: "https://oi-wiki.org/greedy/interval/", Color: "#F97316", Icon: "columns"},
		{Name: "排序贪心", Category: "贪心", Parent: "贪心", Description: "通过排序确定贪心顺序", OjWikiURL: "https://oi-wiki.org/greedy/sorting/", Color: "#F97316", Icon: "arrow-up"},
		{Name: "反悔贪心", Category: "贪心", Parent: "贪心", Description: "允许撤销之前的选择", OjWikiURL: "https://oi-wiki.org/greedy/regret/", Color: "#F97316", Icon: "rotate-ccw"},

		// ── 计算几何 ──────────────────────────────────────────────
		{Name: "计算几何", Category: "计算几何", Description: "几何问题的算法处理", Color: "#EC4899", Icon: "triangle"},
		{Name: "向量", Category: "计算几何", Parent: "计算几何", Description: "向量运算、叉积、点积", OjWikiURL: "https://oi-wiki.org/geometry/vector/", Color: "#EC4899", Icon: "arrow-right"},
		{Name: "凸包", Category: "计算几何", Parent: "计算几何", Description: "Graham Scan、Andrew 算法", OjWikiURL: "https://oi-wiki.org/geometry/convex-hull/", Color: "#EC4899", Icon: "pentagon"},
		{Name: "半平面交", Category: "计算几何", Parent: "计算几何", Description: "多个半平面的交集", OjWikiURL: "https://oi-wiki.org/geometry/half-plane/", Color: "#EC4899", Icon: "intersection"},
		{Name: "最近点对", Category: "计算几何", Parent: "计算几何", Description: "分治法求最近点对距离", OjWikiURL: "https://oi-wiki.org/geometry/nearest-points/", Color: "#EC4899", Icon: "minimize"},
		{Name: "旋转卡壳", Category: "计算几何", Parent: "计算几何", Description: "利用凸包性质求最远点对等问题", OjWikiURL: "https://oi-wiki.org/geometry/rotating-calipers/", Color: "#EC4899", Icon: "rotate-cw"},

		// ── 基础算法 ──────────────────────────────────────────────
		{Name: "基础算法", Category: "基础算法", Description: "常用的基础算法技巧", Color: "#6B7280", Icon: "code"},
		{Name: "二分", Category: "基础算法", Parent: "基础算法", Description: "二分搜索、二分答案", OjWikiURL: "https://oi-wiki.org/basic/binary/", Color: "#6B7280", Icon: "git-branch"},
		{Name: "双指针", Category: "基础算法", Parent: "基础算法", Description: "两指针同向或相向扫描", OjWikiURL: "https://oi-wiki.org/basic/two-pointers/", Color: "#6B7280", Icon: "move-horizontal"},
		{Name: "前缀和", Category: "基础算法", Parent: "基础算法", Description: "预处理区间求和", OjWikiURL: "https://oi-wiki.org/basic/prefix-sum/", Color: "#6B7280", Icon: "bar-chart"},
		{Name: "差分", Category: "基础算法", Parent: "基础算法", Description: "区间加操作的高效处理", OjWikiURL: "https://oi-wiki.org/basic/difference/", Color: "#6B7280", Icon: "minus-circle"},
		{Name: "单调栈", Category: "基础算法", Parent: "基础算法", Description: "维护单调性的栈结构", OjWikiURL: "https://oi-wiki.org/ds/monotonous-stack/", Color: "#6B7280", Icon: "trending-up"},
		{Name: "单调队列", Category: "基础算法", Parent: "基础算法", Description: "维护单调性的队列结构", OjWikiURL: "https://oi-wiki.org/ds/monotonous-queue/", Color: "#6B7280", Icon: "trending-down"},
		{Name: "离散化", Category: "基础算法", Parent: "基础算法", Description: "将大范围值域映射到连续小范围", OjWikiURL: "https://oi-wiki.org/misc/discrete/", Color: "#6B7280", Icon: "compress"},
		{Name: "模拟", Category: "基础算法", Parent: "基础算法", Description: "按题意直接模拟过程", OjWikiURL: "https://oi-wiki.org/basic/simulate/", Color: "#6B7280", Icon: "play"},

		// ── 位运算 ────────────────────────────────────────────────
		{Name: "位运算", Category: "位运算", Description: "按位操作的技巧", Color: "#14B8A6", Icon: "binary"},
		{Name: "位操作", Category: "位运算", Parent: "位运算", Description: "与、或、异或、移位等基本操作", OjWikiURL: "https://oi-wiki.org/math/bit/", Color: "#14B8A6", Icon: "toggle-left"},
		{Name: "状态压缩", Category: "位运算", Parent: "位运算", Description: "用二进制表示集合状态", OjWikiURL: "https://oi-wiki.org/dp/state/", Color: "#14B8A6", Icon: "cpu"},
		{Name: "集合运算", Category: "位运算", Parent: "位运算", Description: "用位运算实现集合的交并补", OjWikiURL: "https://oi-wiki.org/math/bit/", Color: "#14B8A6", Icon: "venn"},
	}

	// First pass: create all top-level category entries and build name->ID map.
	nameToID := make(map[string]uint64)
	for _, d := range defs {
		if d.Parent != "" {
			continue
		}
		kp := models.KnowledgePoint{
			Name:        d.Name,
			Description: d.Description,
			Category:    d.Category,
			OjWikiURL:   d.OjWikiURL,
			Color:       d.Color,
			Icon:        d.Icon,
		}
		if err := conn.Create(&kp).Error; err != nil {
			return err
		}
		nameToID[d.Name] = kp.ID
	}

	// Second pass: create child entries with ParentID set.
	for _, d := range defs {
		if d.Parent == "" {
			continue
		}
		parentID, ok := nameToID[d.Parent]
		if !ok {
			log.Printf("[seed] knowledge parent %q not found for %q, skipping", d.Parent, d.Name)
			continue
		}
		kp := models.KnowledgePoint{
			Name:        d.Name,
			Description: d.Description,
			Category:    d.Category,
			ParentID:    &parentID,
			OjWikiURL:   d.OjWikiURL,
			Color:       d.Color,
			Icon:        d.Icon,
		}
		if err := conn.Create(&kp).Error; err != nil {
			return err
		}
		nameToID[d.Name] = kp.ID
	}

	log.Printf("[seed] %d knowledge points inserted (10 categories, %d sub-topics)", len(nameToID), len(nameToID)-10)
	return nil
}
