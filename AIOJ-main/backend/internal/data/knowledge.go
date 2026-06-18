package data

// KPNode represents a knowledge point in the static OI-Wiki knowledge tree.
type KPNode struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	ParentName  string `json:"parentName,omitempty"` // references another node's Name
	OjWikiURL   string `json:"ojWikiUrl,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// TagDef represents an algorithm tag with its category.
type TagDef struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Parent   string `json:"parent,omitempty"`
	OrderNo  int    `json:"orderNo"`
}

// KnowledgeTree returns the complete static OI-Wiki knowledge point tree.
func KnowledgeTree() []KPNode {
	return []KPNode{
		// ── 基础算法 ──
		{Name: "基础算法（分类）", Category: "基础算法", Description: "常用的基础算法技巧", Color: "#6366f1", Icon: "code"},
		{Name: "基础算法", Category: "基础算法", ParentName: "基础算法（分类）", Description: "常用的基础算法技巧概述", Color: "#6366f1", Icon: "code"},
		{Name: "枚举", Category: "基础算法", ParentName: "基础算法（分类）", Description: "穷举所有可能的情况", OjWikiURL: "https://oi-wiki.org/basic/enumerate/", Color: "#6366f1", Icon: "list"},
		{Name: "模拟", Category: "基础算法", ParentName: "基础算法（分类）", Description: "按照题意直接模拟过程", OjWikiURL: "https://oi-wiki.org/basic/simulate/", Color: "#6366f1", Icon: "play"},
		{Name: "排序", Category: "基础算法", ParentName: "基础算法（分类）", Description: "快速排序、归并排序、计数排序等", OjWikiURL: "https://oi-wiki.org/basic/sort/", Color: "#6366f1", Icon: "arrow-up"},
		{Name: "二分", Category: "基础算法", ParentName: "基础算法（分类）", Description: "二分搜索、二分答案", OjWikiURL: "https://oi-wiki.org/basic/binary/", Color: "#6366f1", Icon: "git-branch"},
		{Name: "双指针", Category: "基础算法", ParentName: "基础算法（分类）", Description: "两指针同向或相向扫描", OjWikiURL: "https://oi-wiki.org/basic/two-pointers/", Color: "#6366f1", Icon: "move-horizontal"},
		{Name: "前缀和", Category: "基础算法", ParentName: "基础算法（分类）", Description: "预处理区间求和", OjWikiURL: "https://oi-wiki.org/basic/prefix-sum/", Color: "#6366f1", Icon: "bar-chart"},
		{Name: "差分", Category: "基础算法", ParentName: "基础算法（分类）", Description: "区间加操作的高效处理", OjWikiURL: "https://oi-wiki.org/basic/difference/", Color: "#6366f1", Icon: "minus-circle"},
		{Name: "分治", Category: "基础算法", ParentName: "基础算法（分类）", Description: "分而治之，将问题拆分为子问题", OjWikiURL: "https://oi-wiki.org/basic/divide-and-conquer/", Color: "#6366f1", Icon: "split"},
		{Name: "贪心", Category: "基础算法", ParentName: "基础算法（分类）", Description: "每步选择局部最优解", OjWikiURL: "https://oi-wiki.org/greedy/", Color: "#6366f1", Icon: "trending-up"},
		{Name: "递归", Category: "基础算法", ParentName: "基础算法（分类）", Description: "函数调用自身的编程技巧", OjWikiURL: "https://oi-wiki.org/basic/recursion/", Color: "#6366f1", Icon: "refresh-cw"},
		{Name: "离散化", Category: "基础算法", ParentName: "基础算法（分类）", Description: "将大范围值域映射到连续小范围", OjWikiURL: "https://oi-wiki.org/misc/discrete/", Color: "#6366f1", Icon: "compress"},

		// ── 数据结构 ──
		{Name: "数据结构（分类）", Category: "数据结构", Description: "组织和存储数据的方式", Color: "#f59e0b", Icon: "database"},
		{Name: "数据结构", Category: "数据结构", ParentName: "数据结构（分类）", Description: "组织和存储数据的方式概述", Color: "#f59e0b", Icon: "database"},
		{Name: "数组", Category: "数据结构", ParentName: "数据结构（分类）", Description: "连续存储的线性结构，支持随机访问", OjWikiURL: "https://oi-wiki.org/ds/array/", Color: "#f59e0b", Icon: "grid"},
		{Name: "链表", Category: "数据结构", ParentName: "数据结构（分类）", Description: "通过指针链接的线性结构", OjWikiURL: "https://oi-wiki.org/ds/linked-list/", Color: "#f59e0b", Icon: "link"},
		{Name: "栈", Category: "数据结构", ParentName: "数据结构（分类）", Description: "后进先出的线性结构", OjWikiURL: "https://oi-wiki.org/ds/stack/", Color: "#f59e0b", Icon: "layers"},
		{Name: "单调栈", Category: "数据结构", ParentName: "数据结构（分类）", Description: "维护单调性的栈结构", OjWikiURL: "https://oi-wiki.org/ds/monotonous-stack/", Color: "#f59e0b", Icon: "trending-up"},
		{Name: "队列", Category: "数据结构", ParentName: "数据结构（分类）", Description: "先进先出的线性结构", OjWikiURL: "https://oi-wiki.org/ds/queue/", Color: "#f59e0b", Icon: "list"},
		{Name: "单调队列", Category: "数据结构", ParentName: "数据结构（分类）", Description: "维护单调性的队列结构", OjWikiURL: "https://oi-wiki.org/ds/monotonous-queue/", Color: "#f59e0b", Icon: "trending-down"},
		{Name: "堆", Category: "数据结构", ParentName: "数据结构（分类）", Description: "优先队列，二叉堆/斐波那契堆", OjWikiURL: "https://oi-wiki.org/ds/heap/", Color: "#f59e0b", Icon: "arrow-up"},
		{Name: "哈希表", Category: "数据结构", ParentName: "数据结构（分类）", Description: "基于哈希函数的键值存储", OjWikiURL: "https://oi-wiki.org/ds/hash/", Color: "#f59e0b", Icon: "key"},
		{Name: "并查集", Category: "数据结构", ParentName: "数据结构（分类）", Description: "不相交集合的合并与查询", OjWikiURL: "https://oi-wiki.org/ds/dsu/", Color: "#f59e0b", Icon: "share"},
		{Name: "字典树", Category: "数据结构", ParentName: "数据结构（分类）", Description: "Trie 树，字符串前缀检索", OjWikiURL: "https://oi-wiki.org/ds/trie/", Color: "#f59e0b", Icon: "book"},
		{Name: "线段树", Category: "数据结构", ParentName: "数据结构（分类）", Description: "区间查询与修改，支持懒标记", OjWikiURL: "https://oi-wiki.org/ds/seg/", Color: "#f59e0b", Icon: "chart-bar"},
		{Name: "树状数组", Category: "数据结构", ParentName: "数据结构（分类）", Description: "前缀和查询与单点修改", OjWikiURL: "https://oi-wiki.org/ds/fenwick/", Color: "#f59e0b", Icon: "bar-chart"},
		{Name: "平衡树", Category: "数据结构", ParentName: "数据结构（分类）", Description: "Treap、Splay、AVL、红黑树等", OjWikiURL: "https://oi-wiki.org/ds/bst/", Color: "#f59e0b", Icon: "scale"},
		{Name: "分块", Category: "数据结构", ParentName: "数据结构（分类）", Description: "将序列分块进行区间操作", OjWikiURL: "https://oi-wiki.org/ds/block-list/", Color: "#f59e0b", Icon: "grid"},

		// ── 动态规划 ──
		{Name: "动态规划（分类）", Category: "动态规划", Description: "通过将问题分解为重叠子问题来求解", Color: "#3b82f6", Icon: "brain"},
		{Name: "动态规划", Category: "动态规划", ParentName: "动态规划（分类）", Description: "通过将问题分解为重叠子问题来求解的算法思想", OjWikiURL: "https://oi-wiki.org/dp/", Color: "#3b82f6", Icon: "brain"},
		{Name: "背包DP", Category: "动态规划", ParentName: "动态规划（分类）", Description: "在容量限制下选择物品以最大化价值", OjWikiURL: "https://oi-wiki.org/dp/knapsack/", Color: "#3b82f6", Icon: "box"},
		{Name: "区间DP", Category: "动态规划", ParentName: "动态规划（分类）", Description: "在连续区间上进行动态规划", OjWikiURL: "https://oi-wiki.org/dp/interval/", Color: "#3b82f6", Icon: "brackets"},
		{Name: "树形DP", Category: "动态规划", ParentName: "动态规划（分类）", Description: "在树结构上进行动态规划", OjWikiURL: "https://oi-wiki.org/dp/tree/", Color: "#3b82f6", Icon: "tree"},
		{Name: "数位DP", Category: "动态规划", ParentName: "动态规划（分类）", Description: "按数字的数位进行状态转移", OjWikiURL: "https://oi-wiki.org/dp/number/", Color: "#3b82f6", Icon: "hash"},
		{Name: "状态压缩DP", Category: "动态规划", ParentName: "动态规划（分类）", Description: "用二进制整数表示集合状态", OjWikiURL: "https://oi-wiki.org/dp/state/", Color: "#3b82f6", Icon: "binary"},
		{Name: "DP优化", Category: "动态规划", ParentName: "动态规划（分类）", Description: "利用单调性、凸性等性质优化转移", OjWikiURL: "https://oi-wiki.org/dp/opt/", Color: "#3b82f6", Icon: "zap"},
		{Name: "计数DP", Category: "动态规划", ParentName: "动态规划（分类）", Description: "统计满足条件的方案数", OjWikiURL: "https://oi-wiki.org/dp/count/", Color: "#3b82f6", Icon: "counter"},
		{Name: "概率DP", Category: "动态规划", ParentName: "动态规划（分类）", Description: "状态转移涉及概率与期望", OjWikiURL: "https://oi-wiki.org/dp/probability/", Color: "#3b82f6", Icon: "dice"},
		{Name: "博弈论DP", Category: "动态规划", ParentName: "动态规划（分类）", Description: "博弈局面的SG函数与必胜态分析", OjWikiURL: "https://oi-wiki.org/dp/game/", Color: "#3b82f6", Icon: "gamepad"},

		// ── 图论 ──
		{Name: "图论（分类）", Category: "图论", Description: "研究图结构及其算法", Color: "#10b981", Icon: "network"},
		{Name: "图论", Category: "图论", ParentName: "图论（分类）", Description: "研究图结构及其算法概述", OjWikiURL: "https://oi-wiki.org/graph/", Color: "#10b981", Icon: "network"},
		{Name: "最短路", Category: "图论", ParentName: "图论（分类）", Description: "Dijkstra、Bellman-Ford、SPFA、Floyd 等算法", OjWikiURL: "https://oi-wiki.org/graph/shortest-path/", Color: "#10b981", Icon: "route"},
		{Name: "最小生成树", Category: "图论", ParentName: "图论（分类）", Description: "Kruskal、Prim 等算法求最小生成树", OjWikiURL: "https://oi-wiki.org/graph/mst/", Color: "#10b981", Icon: "tree"},
		{Name: "网络流", Category: "图论", ParentName: "图论（分类）", Description: "最大流、最小割、费用流", OjWikiURL: "https://oi-wiki.org/graph/flow/", Color: "#10b981", Icon: "flow"},
		{Name: "二分图", Category: "图论", ParentName: "图论（分类）", Description: "匈牙利算法、KM 算法", OjWikiURL: "https://oi-wiki.org/graph/bi-graph/", Color: "#10b981", Icon: "layers"},
		{Name: "拓扑排序", Category: "图论", ParentName: "图论（分类）", Description: "DAG 上的线性排序", OjWikiURL: "https://oi-wiki.org/graph/topo/", Color: "#10b981", Icon: "sort"},
		{Name: "强连通分量", Category: "图论", ParentName: "图论（分类）", Description: "Tarjan、Kosaraju 算法", OjWikiURL: "https://oi-wiki.org/graph/scc/", Color: "#10b981", Icon: "loop"},
		{Name: "桥和割点", Category: "图论", ParentName: "图论（分类）", Description: "寻找图中的桥和割点", OjWikiURL: "https://oi-wiki.org/graph/cut/", Color: "#10b981", Icon: "scissors"},
		{Name: "树上问题", Category: "图论", ParentName: "图论（分类）", Description: "树链剖分、重心、直径等", OjWikiURL: "https://oi-wiki.org/graph/tree-basic/", Color: "#10b981", Icon: "tree-deciduous"},
		{Name: "LCA", Category: "图论", ParentName: "图论（分类）", Description: "最近公共祖先，倍增/Tarjan/树剖", OjWikiURL: "https://oi-wiki.org/graph/lca/", Color: "#10b981", Icon: "merge"},

		// ── 数学 ──
		{Name: "数学（分类）", Category: "数学", Description: "算法竞赛中的数学知识", Color: "#8b5cf6", Icon: "calculator"},
		{Name: "数学", Category: "数学", ParentName: "数学（分类）", Description: "算法竞赛中的数学知识概述", OjWikiURL: "https://oi-wiki.org/math/", Color: "#8b5cf6", Icon: "calculator"},
		{Name: "质数", Category: "数学", ParentName: "数学（分类）", Description: "素数筛法、质因数分解", OjWikiURL: "https://oi-wiki.org/math/number-theory/prime/", Color: "#8b5cf6", Icon: "hash"},
		{Name: "GCD/LCM", Category: "数学", ParentName: "数学（分类）", Description: "欧几里得算法与最小公倍数", OjWikiURL: "https://oi-wiki.org/math/number-theory/gcd/", Color: "#8b5cf6", Icon: "divide"},
		{Name: "快速幂", Category: "数学", ParentName: "数学（分类）", Description: "二进制取模幂运算", OjWikiURL: "https://oi-wiki.org/math/binary-exponentiation/", Color: "#8b5cf6", Icon: "zap"},
		{Name: "模运算", Category: "数学", ParentName: "数学（分类）", Description: "取模运算、模逆元、中国剩余定理", OjWikiURL: "https://oi-wiki.org/math/number-theory/inverse/", Color: "#8b5cf6", Icon: "percent"},
		{Name: "组合数学", Category: "数学", ParentName: "数学（分类）", Description: "排列组合、卡特兰数、斯特林数", OjWikiURL: "https://oi-wiki.org/math/combinatorics/", Color: "#8b5cf6", Icon: "shuffle"},
		{Name: "容斥原理", Category: "数学", ParentName: "数学（分类）", Description: "集合计数中的容斥公式", OjWikiURL: "https://oi-wiki.org/math/combinatorics/inclusion-exclusion/", Color: "#8b5cf6", Icon: "circle"},
		{Name: "概率期望", Category: "数学", ParentName: "数学（分类）", Description: "概率论基础与期望计算", OjWikiURL: "https://oi-wiki.org/math/probability/", Color: "#8b5cf6", Icon: "pie-chart"},
		{Name: "矩阵", Category: "数学", ParentName: "数学（分类）", Description: "矩阵运算与矩阵快速幂", OjWikiURL: "https://oi-wiki.org/math/matrix/", Color: "#8b5cf6", Icon: "grid"},
		{Name: "高斯消元", Category: "数学", ParentName: "数学（分类）", Description: "线性方程组求解", OjWikiURL: "https://oi-wiki.org/math/gauss/", Color: "#8b5cf6", Icon: "columns"},
		{Name: "莫比乌斯反演", Category: "数学", ParentName: "数学（分类）", Description: "数论中的莫比乌斯函数与反演公式", OjWikiURL: "https://oi-wiki.org/math/mobius/", Color: "#8b5cf6", Icon: "refresh-cw"},
		{Name: "博弈论", Category: "数学", ParentName: "数学（分类）", Description: "Nim游戏、SG函数、公平组合博弈", OjWikiURL: "https://oi-wiki.org/math/game-theory/", Color: "#8b5cf6", Icon: "gamepad"},

		// ── 字符串 ──
		{Name: "字符串（分类）", Category: "字符串", Description: "字符串处理与匹配算法", Color: "#ef4444", Icon: "type"},
		{Name: "字符串", Category: "字符串", ParentName: "字符串（分类）", Description: "字符串处理与匹配算法概述", OjWikiURL: "https://oi-wiki.org/string/", Color: "#ef4444", Icon: "type"},
		{Name: "字符串处理", Category: "字符串", ParentName: "字符串（分类）", Description: "字符串模拟、遍历、转换等基础操作", OjWikiURL: "https://oi-wiki.org/string/basic/", Color: "#ef4444", Icon: "edit"},
		{Name: "KMP", Category: "字符串", ParentName: "字符串（分类）", Description: "Knuth-Morris-Pratt 单模式匹配", OjWikiURL: "https://oi-wiki.org/string/kmp/", Color: "#ef4444", Icon: "search"},
		{Name: "Trie", Category: "字符串", ParentName: "字符串（分类）", Description: "字符串前缀树", OjWikiURL: "https://oi-wiki.org/string/trie/", Color: "#ef4444", Icon: "book"},
		{Name: "后缀数组", Category: "字符串", ParentName: "字符串（分类）", Description: "SA 与 LCP 数组", OjWikiURL: "https://oi-wiki.org/string/sa/", Color: "#ef4444", Icon: "align-left"},
		{Name: "后缀自动机", Category: "字符串", ParentName: "字符串（分类）", Description: "SAM，线性构建的后缀结构", OjWikiURL: "https://oi-wiki.org/string/sam/", Color: "#ef4444", Icon: "cpu"},
		{Name: "AC自动机", Category: "字符串", ParentName: "字符串（分类）", Description: "多模式匹配自动机", OjWikiURL: "https://oi-wiki.org/string/ac-automaton/", Color: "#ef4444", Icon: "filter"},
		{Name: "Manacher", Category: "字符串", ParentName: "字符串（分类）", Description: "线性时间求最长回文子串", OjWikiURL: "https://oi-wiki.org/string/manacher/", Color: "#ef4444", Icon: "repeat"},
		{Name: "哈希", Category: "字符串", ParentName: "字符串（分类）", Description: "字符串哈希与子串比较", OjWikiURL: "https://oi-wiki.org/string/hash/", Color: "#ef4444", Icon: "hash"},

		// ── 搜索 ──
		{Name: "搜索（分类）", Category: "搜索", Description: "系统地探索解空间", Color: "#06b6d4", Icon: "search"},
		{Name: "搜索", Category: "搜索", ParentName: "搜索（分类）", Description: "系统地探索解空间概述", OjWikiURL: "https://oi-wiki.org/search/", Color: "#06b6d4", Icon: "search"},
		{Name: "BFS", Category: "搜索", ParentName: "搜索（分类）", Description: "广度优先搜索，最短路模型", OjWikiURL: "https://oi-wiki.org/search/bfs/", Color: "#06b6d4", Icon: "layers"},
		{Name: "DFS", Category: "搜索", ParentName: "搜索（分类）", Description: "深度优先搜索，回溯框架", OjWikiURL: "https://oi-wiki.org/search/dfs/", Color: "#06b6d4", Icon: "git-branch"},
		{Name: "迭代加深", Category: "搜索", ParentName: "搜索（分类）", Description: "逐步增加搜索深度限制", OjWikiURL: "https://oi-wiki.org/search/iterative/", Color: "#06b6d4", Icon: "chevrons-down"},
		{Name: "IDA*", Category: "搜索", ParentName: "搜索（分类）", Description: "迭代加深 A* 搜索", OjWikiURL: "https://oi-wiki.org/search/idastar/", Color: "#06b6d4", Icon: "target"},
		{Name: "双向BFS", Category: "搜索", ParentName: "搜索（分类）", Description: "从起点和终点同时搜索", OjWikiURL: "https://oi-wiki.org/search/bidirectional/", Color: "#06b6d4", Icon: "repeat"},
		{Name: "启发式搜索", Category: "搜索", ParentName: "搜索（分类）", Description: "带估价函数的优先搜索（A*）", OjWikiURL: "https://oi-wiki.org/search/astar/", Color: "#06b6d4", Icon: "star"},
		{Name: "折半搜索", Category: "搜索", ParentName: "搜索（分类）", Description: "将搜索空间分成两半分别枚举", OjWikiURL: "https://oi-wiki.org/search/half/", Color: "#06b6d4", Icon: "scissors"},
		{Name: "回溯", Category: "搜索", ParentName: "搜索（分类）", Description: "试探性搜索，不满足条件时回退", OjWikiURL: "https://oi-wiki.org/search/backtracking/", Color: "#06b6d4", Icon: "rotate-ccw"},

		// ── 贪心 ──
		{Name: "贪心（分类）", Category: "贪心", Description: "每步选择局部最优解的策略", Color: "#f97316", Icon: "trending-up"},
		{Name: "贪心算法", Category: "贪心", ParentName: "贪心（分类）", Description: "每步选择局部最优解的策略概述", OjWikiURL: "https://oi-wiki.org/greedy/", Color: "#f97316", Icon: "trending-up"},
		{Name: "区间贪心", Category: "贪心", ParentName: "贪心（分类）", Description: "区间调度、不相交区间选择", OjWikiURL: "https://oi-wiki.org/greedy/interval/", Color: "#f97316", Icon: "columns"},
		{Name: "排序贪心", Category: "贪心", ParentName: "贪心（分类）", Description: "通过排序确定贪心顺序", OjWikiURL: "https://oi-wiki.org/greedy/sorting/", Color: "#f97316", Icon: "arrow-up"},
		{Name: "反悔贪心", Category: "贪心", ParentName: "贪心（分类）", Description: "允许撤销之前的选择", OjWikiURL: "https://oi-wiki.org/greedy/regret/", Color: "#f97316", Icon: "rotate-ccw"},

		// ── 计算几何 ──
		{Name: "计算几何（分类）", Category: "计算几何", Description: "几何问题的算法处理", Color: "#ec4899", Icon: "triangle"},
		{Name: "计算几何", Category: "计算几何", ParentName: "计算几何（分类）", Description: "几何问题的算法处理概述", OjWikiURL: "https://oi-wiki.org/geometry/", Color: "#ec4899", Icon: "triangle"},
		{Name: "向量", Category: "计算几何", ParentName: "计算几何（分类）", Description: "向量运算、叉积、点积", OjWikiURL: "https://oi-wiki.org/geometry/vector/", Color: "#ec4899", Icon: "arrow-right"},
		{Name: "凸包", Category: "计算几何", ParentName: "计算几何（分类）", Description: "Graham Scan、Andrew 算法", OjWikiURL: "https://oi-wiki.org/geometry/convex-hull/", Color: "#ec4899", Icon: "pentagon"},
		{Name: "半平面交", Category: "计算几何", ParentName: "计算几何（分类）", Description: "多个半平面的交集", OjWikiURL: "https://oi-wiki.org/geometry/half-plane/", Color: "#ec4899", Icon: "intersection"},
		{Name: "最近点对", Category: "计算几何", ParentName: "计算几何（分类）", Description: "分治法求最近点对距离", OjWikiURL: "https://oi-wiki.org/geometry/nearest-points/", Color: "#ec4899", Icon: "minimize"},
		{Name: "旋转卡壳", Category: "计算几何", ParentName: "计算几何（分类）", Description: "利用凸包性质求最远点对等问题", OjWikiURL: "https://oi-wiki.org/geometry/rotating-calipers/", Color: "#ec4899", Icon: "rotate-cw"},

		// ── 位运算 ──
		{Name: "位运算（分类）", Category: "位运算", Description: "按位操作的技巧", Color: "#14b8a6", Icon: "binary"},
		{Name: "位运算", Category: "位运算", ParentName: "位运算（分类）", Description: "按位操作的技巧概述", OjWikiURL: "https://oi-wiki.org/math/bit/", Color: "#14b8a6", Icon: "binary"},
		{Name: "位操作", Category: "位运算", ParentName: "位运算（分类）", Description: "与、或、异或、移位等基本操作", OjWikiURL: "https://oi-wiki.org/math/bit/", Color: "#14b8a6", Icon: "toggle-left"},
		{Name: "状态压缩", Category: "位运算", ParentName: "位运算（分类）", Description: "用二进制表示集合状态", OjWikiURL: "https://oi-wiki.org/dp/state/", Color: "#14b8a6", Icon: "cpu"},
		{Name: "集合运算", Category: "位运算", ParentName: "位运算（分类）", Description: "用位运算实现集合的交并补", OjWikiURL: "https://oi-wiki.org/math/bit/", Color: "#14b8a6", Icon: "venn"},
	}
}

// KnowledgeMap returns a map of node Name → KPNode for quick lookup.
func KnowledgeMap() map[string]KPNode {
	m := make(map[string]KPNode, len(KnowledgeTree()))
	for _, n := range KnowledgeTree() {
		m[n.Name] = n
	}
	return m
}

// Tags returns all algorithm tags derived from the knowledge tree.
// Category nodes (with "（分类）" suffix) become category-only entries,
// leaf nodes become individual tags.
func Tags() []TagDef {
	var tags []TagDef
	orderNo := 0
	for _, n := range KnowledgeTree() {
		parent := ""
		if n.ParentName != "" {
			parent = n.ParentName
		}
		tags = append(tags, TagDef{
			Name:     n.Name,
			Category: n.Category,
			Parent:   parent,
			OrderNo:  orderNo,
		})
		orderNo++
	}
	return tags
}

// TagNames returns just the flat list of tag names.
func TagNames() []string {
	tags := Tags()
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names
}
