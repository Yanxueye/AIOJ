package models

// AlgorithmTag defines all valid algorithm tags in the system.
// Problems reference tags by name (stored in Problem.Tags as StringSlice).
// This table serves as the single source of truth for tag names.
type AlgorithmTag struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Category string `gorm:"type:varchar(32);index;not null" json:"category"`
	Parent   string `gorm:"type:varchar(64);index" json:"parent,omitempty"`
	OrderNo  int    `gorm:"default:0" json:"orderNo"`
}

func (AlgorithmTag) TableName() string { return "algorithm_tags" }

// AllAlgorithmTags returns the complete list of algorithm tags for seeding.
func AllAlgorithmTags() []AlgorithmTag {
	return []AlgorithmTag{
		// ============ 基础算法 ============
		{Name: "模拟", Category: "基础算法", OrderNo: 1},
		{Name: "枚举", Category: "基础算法", OrderNo: 2},
		{Name: "递归", Category: "基础算法", OrderNo: 3},
		{Name: "分治", Category: "基础算法", OrderNo: 4},
		{Name: "贪心", Category: "基础算法", OrderNo: 5},
		{Name: "排序", Category: "基础算法", OrderNo: 6},
		{Name: "二分", Category: "基础算法", OrderNo: 7},
		{Name: "双指针", Category: "基础算法", OrderNo: 8},
		{Name: "前缀和", Category: "基础算法", OrderNo: 9},
		{Name: "差分", Category: "基础算法", OrderNo: 10},
		{Name: "位运算", Category: "基础算法", OrderNo: 11},
		{Name: "哈希", Category: "基础算法", OrderNo: 12},

		// ============ 动态规划 ============
		{Name: "动态规划", Category: "动态规划", OrderNo: 1},
		{Name: "背包", Category: "动态规划", Parent: "动态规划", OrderNo: 2},
		{Name: "区间DP", Category: "动态规划", Parent: "动态规划", OrderNo: 3},
		{Name: "树形DP", Category: "动态规划", Parent: "动态规划", OrderNo: 4},
		{Name: "状压DP", Category: "动态规划", Parent: "动态规划", OrderNo: 5},
		{Name: "数位DP", Category: "动态规划", Parent: "动态规划", OrderNo: 6},
		{Name: "概率DP", Category: "动态规划", Parent: "动态规划", OrderNo: 7},
		{Name: "计数DP", Category: "动态规划", Parent: "动态规划", OrderNo: 8},
		{Name: "DP优化", Category: "动态规划", Parent: "动态规划", OrderNo: 9},
		{Name: "LIS", Category: "动态规划", Parent: "动态规划", OrderNo: 10},
		{Name: "LCS", Category: "动态规划", Parent: "动态规划", OrderNo: 11},

		// ============ 图论 ============
		{Name: "图论", Category: "图论", OrderNo: 1},
		{Name: "BFS", Category: "图论", Parent: "图论", OrderNo: 2},
		{Name: "DFS", Category: "图论", Parent: "图论", OrderNo: 3},
		{Name: "最短路径", Category: "图论", Parent: "图论", OrderNo: 4},
		{Name: "最小生成树", Category: "图论", Parent: "图论", OrderNo: 5},
		{Name: "拓扑排序", Category: "图论", Parent: "图论", OrderNo: 6},
		{Name: "二分图", Category: "图论", Parent: "图论", OrderNo: 7},
		{Name: "网络流", Category: "图论", Parent: "图论", OrderNo: 8},
		{Name: "强连通分量", Category: "图论", Parent: "图论", OrderNo: 9},
		{Name: "割点与桥", Category: "图论", Parent: "图论", OrderNo: 10},
		{Name: "欧拉路径", Category: "图论", Parent: "图论", OrderNo: 11},
		{Name: "并查集", Category: "图论", Parent: "图论", OrderNo: 12},

		// ============ 数据结构 ============
		{Name: "数据结构", Category: "数据结构", OrderNo: 1},
		{Name: "栈", Category: "数据结构", Parent: "数据结构", OrderNo: 2},
		{Name: "队列", Category: "数据结构", Parent: "数据结构", OrderNo: 3},
		{Name: "堆", Category: "数据结构", Parent: "数据结构", OrderNo: 4},
		{Name: "链表", Category: "数据结构", Parent: "数据结构", OrderNo: 5},
		{Name: "树", Category: "数据结构", Parent: "数据结构", OrderNo: 6},
		{Name: "二叉搜索树", Category: "数据结构", Parent: "树", OrderNo: 7},
		{Name: "线段树", Category: "数据结构", Parent: "数据结构", OrderNo: 8},
		{Name: "树状数组", Category: "数据结构", Parent: "数据结构", OrderNo: 9},
		{Name: "字典树", Category: "数据结构", Parent: "数据结构", OrderNo: 10},
		{Name: "平衡树", Category: "数据结构", Parent: "数据结构", OrderNo: 11},
		{Name: "单调栈", Category: "数据结构", Parent: "栈", OrderNo: 12},
		{Name: "单调队列", Category: "数据结构", Parent: "队列", OrderNo: 13},

		// ============ 字符串 ============
		{Name: "字符串", Category: "字符串", OrderNo: 1},
		{Name: "KMP", Category: "字符串", Parent: "字符串", OrderNo: 2},
		{Name: "字符串哈希", Category: "字符串", Parent: "字符串", OrderNo: 3},
		{Name: "Manacher", Category: "字符串", Parent: "字符串", OrderNo: 4},
		{Name: "后缀数组", Category: "字符串", Parent: "字符串", OrderNo: 5},
		{Name: "后缀自动机", Category: "字符串", Parent: "字符串", OrderNo: 6},
		{Name: "AC自动机", Category: "字符串", Parent: "字符串", OrderNo: 7},

		// ============ 数学 ============
		{Name: "数学", Category: "数学", OrderNo: 1},
		{Name: "数论", Category: "数学", Parent: "数学", OrderNo: 2},
		{Name: "质数", Category: "数学", Parent: "数论", OrderNo: 3},
		{Name: "GCD/LCM", Category: "数学", Parent: "数论", OrderNo: 4},
		{Name: "快速幂", Category: "数学", Parent: "数论", OrderNo: 5},
		{Name: "矩阵", Category: "数学", Parent: "数学", OrderNo: 6},
		{Name: "组合数学", Category: "数学", Parent: "数学", OrderNo: 7},
		{Name: "概率", Category: "数学", Parent: "数学", OrderNo: 8},
		{Name: "博弈论", Category: "数学", Parent: "数学", OrderNo: 9},
		{Name: "高斯消元", Category: "数学", Parent: "数学", OrderNo: 10},
		{Name: "容斥原理", Category: "数学", Parent: "数学", OrderNo: 11},
		{Name: "莫比乌斯反演", Category: "数学", Parent: "数学", OrderNo: 12},
		{Name: "FFT/NTT", Category: "数学", Parent: "数学", OrderNo: 13},

		// ============ 搜索 ============
		{Name: "搜索", Category: "搜索", OrderNo: 1},
		{Name: "回溯", Category: "搜索", Parent: "搜索", OrderNo: 2},
		{Name: "剪枝", Category: "搜索", Parent: "搜索", OrderNo: 3},
		{Name: "迭代加深", Category: "搜索", Parent: "搜索", OrderNo: 4},
		{Name: "IDA*", Category: "搜索", Parent: "搜索", OrderNo: 5},
		{Name: "A*", Category: "搜索", Parent: "搜索", OrderNo: 6},
		{Name: "启发式搜索", Category: "搜索", Parent: "搜索", OrderNo: 7},

		// ============ 计算几何 ============
		{Name: "计算几何", Category: "计算几何", OrderNo: 1},
		{Name: "凸包", Category: "计算几何", Parent: "计算几何", OrderNo: 2},
		{Name: "半平面交", Category: "计算几何", Parent: "计算几何", OrderNo: 3},
		{Name: "旋转卡壳", Category: "计算几何", Parent: "计算几何", OrderNo: 4},
		{Name: "最近点对", Category: "计算几何", Parent: "计算几何", OrderNo: 5},

		// ============ 杂项 ============
		{Name: "设计", Category: "杂项", OrderNo: 1},
		{Name: "矩阵快速幂", Category: "杂项", OrderNo: 2},
		{Name: "离散化", Category: "杂项", OrderNo: 3},
		{Name: "分块", Category: "杂项", OrderNo: 4},
		{Name: "莫队", Category: "杂项", OrderNo: 5},
		{Name: "随机化", Category: "杂项", OrderNo: 6},
	}
}
