package database

import (
	"log"

	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

// Seed ensures the database has a minimum set of problems / announcements
// / demo user so the frontend can be exercised end-to-end out of the box.
func Seed(conn *gorm.DB) error {
	if err := seedUsers(conn); err != nil {
		return err
	}
	if err := seedProblems(conn); err != nil {
		return err
	}
	if err := seedAnnouncements(conn); err != nil {
		return err
	}
	return nil
}

func seedUsers(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.User{}).Count(&count)
	if count > 0 {
		return nil
	}
	hash, err := utils.HashPassword("123456")
	if err != nil {
		return err
	}
	demo := &models.User{
		Username:     "coder_test",
		Email:        "test@terminaloj.com",
		PasswordHash: hash,
		Bio:          "热爱算法的开发者",
		Rating:       1520,
	}
	if err := conn.Create(demo).Error; err != nil {
		return err
	}
	log.Println("[seed] default user coder_test / 123456 created")
	return nil
}

func seedProblems(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.Problem{}).Count(&count)
	if count > 0 {
		return nil
	}
	problems := []models.Problem{
		{
			ID:              1001,
			Title:           "两数之和",
			Difficulty:      "简单",
			DifficultyScore: 800,
			Tags:            models.StringSlice{"数组", "哈希表"},
			Content: "# 两数之和\n\n给定整数数组 `nums` 和目标值 $target$，在数组中找出两个数使其和等于目标值，返回它们的下标。\n\n## 输入\n第一行两个整数 $n, target$，第二行 $n$ 个整数。\n\n## 输出\n输出两个下标（从 0 开始），用空格分隔。\n\n## 样例\n```\n4 9\n2 7 11 15\n```\n输出：\n```\n0 1\n```",
			TimeLimit:       1000,
			MemoryLimit:     256,
			Source:          "TerminalOJ 原创",
			TestCases: models.TestCases{
				{Input: "4 9\n2 7 11 15\n", Expected: "0 1"},
				{Input: "3 6\n3 2 4\n", Expected: "1 2"},
			},
		},
		{
			ID:              1002,
			Title:           "最长回文子串",
			Difficulty:      "中等",
			DifficultyScore: 1300,
			Tags:            models.StringSlice{"字符串", "动态规划"},
			Content: "# 最长回文子串\n\n给定字符串 $s$，找到其中最长的回文子串。\n\n## 样例\n输入：`babad` 输出：`bab`（或 `aba`）",
			TimeLimit:       1500,
			MemoryLimit:     256,
			Source:          "TerminalOJ 原创",
			TestCases: models.TestCases{
				{Input: "babad", Expected: "bab"},
			},
		},
		{
			ID:              1003,
			Title:           "合并K个升序链表",
			Difficulty:      "困难",
			DifficultyScore: 1900,
			Tags:            models.StringSlice{"堆", "链表", "分治"},
			Content: "# 合并 K 个升序链表\n\n将 $k$ 个有序链表合并为一个有序链表并输出。\n\n复杂度要求：$O(N \\log k)$，其中 $N$ 为总节点数。",
			TimeLimit:       2000,
			MemoryLimit:     512,
			Source:          "TerminalOJ 原创",
			TestCases: models.TestCases{
				{Input: "3\n1 4 5\n1 3 4\n2 6\n", Expected: "1 1 2 3 4 4 5 6"},
			},
		},
		{
			ID:              1004,
			Title:           "零钱兑换",
			Difficulty:      "中等",
			DifficultyScore: 1400,
			Tags:            models.StringSlice{"动态规划", "贪心"},
			Content: "# 零钱兑换\n\n给定不同面额的硬币和一个总金额，计算凑成总金额所需的最少硬币个数，若不能凑成则输出 -1。",
			TimeLimit:       1000,
			MemoryLimit:     256,
			Source:          "TerminalOJ 原创",
			TestCases: models.TestCases{
				{Input: "3 11\n1 2 5\n", Expected: "3"},
			},
		},
		{
			ID:              1005,
			Title:           "岛屿数量",
			Difficulty:      "中等",
			DifficultyScore: 1200,
			Tags:            models.StringSlice{"搜索", "图论", "并查集"},
			Content: "# 岛屿数量\n\n给定 $m \\times n$ 的 01 矩阵，'1' 表示陆地 '0' 表示水域，相邻陆地属于同一岛屿，求岛屿数量。",
			TimeLimit:       1200,
			MemoryLimit:     256,
			Source:          "TerminalOJ 原创",
			TestCases: models.TestCases{
				{Input: "3 3\n110\n110\n001\n", Expected: "2"},
			},
		},
	}
	if err := conn.CreateInBatches(problems, 20).Error; err != nil {
		return err
	}
	log.Printf("[seed] %d problems inserted", len(problems))
	return nil
}

func seedAnnouncements(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.Announcement{}).Count(&count)
	if count > 0 {
		return nil
	}
	items := []models.Announcement{
		{Title: "TerminalOJ 正式上线", Content: "欢迎使用 TerminalOJ！我们提供题库、在线评测、AI 辅导一体化服务。", Type: "success", Date: "2026-04-01"},
		{Title: "每周题目更新", Content: "每周一、周四 18:00 更新 10 道新题，欢迎挑战。", Type: "info", Date: "2026-04-06"},
		{Title: "服务维护通知", Content: "4 月 20 日凌晨 2:00-3:00 评测队列将短暂停机维护。", Type: "warning", Date: "2026-04-18"},
		{Title: "AI 训练功能上线", Content: "集成 AI 辅助解题，支持 Markdown / LaTeX 渲染，欢迎体验。", Type: "primary", Date: "2026-04-15"},
	}
	return conn.Create(&items).Error
}
