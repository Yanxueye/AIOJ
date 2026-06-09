package database

import (
	"log"

	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

// Seed ensures the database has a minimum set of demo data.
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
	hash, err := utils.HashPassword("123456")
	if err != nil {
		return err
	}

	coder := models.User{
		Username:     "coder_test",
		Email:        "test@terminaloj.com",
		PasswordHash: hash,
		Role:         "user",
		Bio:          "热爱算法的开发者",
		Rating:       1520,
	}
	admin := models.User{
		Username:     "admin",
		Email:        "admin@terminaloj.com",
		PasswordHash: hash,
		Role:         "admin",
		Bio:          "题库与系统管理员",
		Rating:       1800,
	}

	if err := ensureSeedUser(conn, coder); err != nil {
		return err
	}
	if err := ensureSeedUser(conn, admin); err != nil {
		return err
	}
	log.Println("[seed] ensured default users: coder_test(user) / 123456, admin(admin) / 123456")
	return nil
}

func ensureSeedUser(conn *gorm.DB, seed models.User) error {
	var user models.User
	err := conn.Where("username = ?", seed.Username).First(&user).Error
	if err == nil {
		user.Email = seed.Email
		user.PasswordHash = seed.PasswordHash
		user.Role = seed.Role
		if user.Bio == "" || user.Username == "admin" || user.Username == "coder_test" {
			user.Bio = seed.Bio
		}
		if user.Rating == 0 {
			user.Rating = seed.Rating
		}
		return conn.Save(&user).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return conn.Create(&seed).Error
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
			Content:         "# 两数之和\n\n给定整数数组 `nums` 和目标值 `target`，请在数组中找到和为目标值的两个下标。\n\n## 输入\n第一行包含 `n target`，第二行包含 `n` 个整数。\n\n## 输出\n输出两个下标，从 0 开始。\n",
			TimeLimit:       1000,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
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
			Content:         "# 最长回文子串\n\n给定字符串 `s`，找到其中最长的回文子串。\n",
			TimeLimit:       1500,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
			Source:          "TerminalOJ 原创",
			TestCases: models.TestCases{
				{Input: "babad", Expected: "bab"},
			},
		},
		{
			ID:              1003,
			Title:           "合并 K 个升序链表",
			Difficulty:      "困难",
			DifficultyScore: 1900,
			Tags:            models.StringSlice{"堆", "链表", "分治"},
			Content:         "# 合并 K 个升序链表\n\n将多个升序序列合并成一个升序结果。\n",
			TimeLimit:       2000,
			MemoryLimit:     512,
			OutputLimitKB:   1024,
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
			Content:         "# 零钱兑换\n\n给定硬币面额和总金额，求最少硬币数，不可达则输出 -1。\n",
			TimeLimit:       1000,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
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
			Content:         "# 岛屿数量\n\n给定 01 矩阵，统计连通陆地块数量。\n",
			TimeLimit:       1200,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
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
		{Title: "TerminalOJ 正式上线", Content: "欢迎使用 TerminalOJ，现已提供题库、在线评测和 AI 学习支持。", Type: "success", Date: "2026-04-01"},
		{Title: "每周题目更新", Content: "每周固定更新新题，欢迎持续练习。", Type: "info", Date: "2026-04-06"},
		{Title: "服务维护通知", Content: "评测服务将在维护窗口短暂停机。", Type: "warning", Date: "2026-04-18"},
		{Title: "AI 训练功能上线", Content: "支持题目讲解、诊断与学习辅助。", Type: "primary", Date: "2026-04-15"},
	}
	return conn.Create(&items).Error
}
