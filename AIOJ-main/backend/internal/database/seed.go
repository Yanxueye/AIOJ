package database

import (
	"log"
	"time"

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
	if err := seedAdditionalProblems(conn); err != nil {
		return err
	}
	if err := seedStudyPlans(conn); err != nil {
		return err
	}
	if err := seedDailyChallenges(conn); err != nil {
		return err
	}
	if err := seedAnnouncements(conn); err != nil {
		return err
	}
	if err := seedRatingHistory(conn); err != nil {
		return err
	}
	if err := seedSubmissions(conn); err != nil {
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
		Rating:       1320,
	}
	admin := models.User{
		Username:     "admin",
		Email:        "admin@terminaloj.com",
		PasswordHash: hash,
		Role:         "admin",
		Bio:          "题库与系统管理员",
		Rating:       1600,
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
		// Always sync seed user ratings
		if user.Username == "coder_test" || user.Username == "admin" {
			user.Rating = seed.Rating
		} else if user.Rating == 0 {
			user.Rating = seed.Rating
		}
		return conn.Save(&user).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return conn.Create(&seed).Error
}

type seededProblem struct {
	ID              uint64
	Title           string
	Difficulty      string
	DifficultyScore int
	Rating          int
	Tags            models.StringSlice
	Source          string
	TimeLimit       int
	MemoryLimit     int
	OutputLimitKB   int32
	Content         string
	Constraints     string
	Editorial       string
	Samples         []models.ProblemSample
	TestCases       []models.ProblemTestCase
	Templates       []models.ProblemTemplate
}

func seedProblems(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.Problem{}).Count(&count)
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	adminID := seededAdminID(conn)
	problems := []seededProblem{
		{
			ID:              1001,
			Title:           "两数之和",
			Difficulty:      "简单",
			DifficultyScore: 800,
			Rating:          800,
			Tags:            models.StringSlice{"数组", "哈希表"},
			Source:          "TerminalOJ 原创",
			TimeLimit:       1000,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
			Content:         "# 两数之和\n\n给定整数数组 `nums` 和目标值 `target`，请在数组中找到和为目标值的两个下标。\n",
			Constraints:     "1 <= n <= 1e5，输出从 0 开始。",
			Editorial:       "使用哈希表记录已经遍历过的值与下标，单次扫描即可完成。",
			Samples: []models.ProblemSample{
				{CaseNo: 1, Input: "4 9\n2 7 11 15\n", Expected: "0 1", Explanation: "nums[0] + nums[1] = 9"},
				{CaseNo: 2, Input: "3 6\n3 2 4\n", Expected: "1 2", Explanation: "nums[1] + nums[2] = 6"},
			},
			TestCases: []models.ProblemTestCase{
				{CaseNo: 1, Input: "4 9\n2 7 11 15\n", Expected: "0 1", IsHidden: false},
				{CaseNo: 2, Input: "3 6\n3 2 4\n", Expected: "1 2", IsHidden: false},
				{CaseNo: 3, Input: "5 14\n1 5 3 7 9\n", Expected: "1 4", IsHidden: true},
			},
			Templates: defaultTemplates(),
		},
		{
			ID:              1002,
			Title:           "最长回文子串",
			Difficulty:      "中等",
			DifficultyScore: 1300,
			Rating:          1300,
			Tags:            models.StringSlice{"字符串", "动态规划"},
			Source:          "TerminalOJ 原创",
			TimeLimit:       1500,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
			Content:         "# 最长回文子串\n\n给定字符串 `s`，找到其中最长的回文子串。\n",
			Constraints:     "1 <= |s| <= 1000。",
			Editorial:       "可以使用中心扩展或 DP。中心扩展实现更简洁。",
			Samples: []models.ProblemSample{
				{CaseNo: 1, Input: "babad", Expected: "bab", Explanation: "也可以输出 aba"},
				{CaseNo: 2, Input: "cbbd", Expected: "bb", Explanation: "最长回文长度为 2"},
			},
			TestCases: []models.ProblemTestCase{
				{CaseNo: 1, Input: "babad", Expected: "bab", IsHidden: false},
				{CaseNo: 2, Input: "cbbd", Expected: "bb", IsHidden: false},
				{CaseNo: 3, Input: "a", Expected: "a", IsHidden: true},
			},
			Templates: defaultTemplates(),
		},
		{
			ID:              1003,
			Title:           "合并 K 个升序链表",
			Difficulty:      "困难",
			DifficultyScore: 1900,
			Rating:          1900,
			Tags:            models.StringSlice{"堆", "链表", "分治"},
			Source:          "TerminalOJ 原创",
			TimeLimit:       2000,
			MemoryLimit:     512,
			OutputLimitKB:   1024,
			Content:         "# 合并 K 个升序链表\n\n将多个升序序列合并成一个升序结果。\n",
			Constraints:     "K <= 1e4，总节点数 <= 1e5。",
			Editorial:       "优先队列维护每个链表头节点，时间复杂度 O(N log K)。",
			Samples: []models.ProblemSample{
				{CaseNo: 1, Input: "3\n1 4 5\n1 3 4\n2 6\n", Expected: "1 1 2 3 4 4 5 6", Explanation: "按升序合并"},
				{CaseNo: 2, Input: "0\n", Expected: "", Explanation: "空输入输出空"},
			},
			TestCases: []models.ProblemTestCase{
				{CaseNo: 1, Input: "3\n1 4 5\n1 3 4\n2 6\n", Expected: "1 1 2 3 4 4 5 6", IsHidden: false},
				{CaseNo: 2, Input: "0\n", Expected: "", IsHidden: false},
				{CaseNo: 3, Input: "2\n\n1 2 3\n", Expected: "1 2 3", IsHidden: true},
			},
			Templates: defaultTemplates(),
		},
		{
			ID:              1004,
			Title:           "零钱兑换",
			Difficulty:      "中等",
			DifficultyScore: 1400,
			Rating:          1400,
			Tags:            models.StringSlice{"动态规划", "贪心"},
			Source:          "TerminalOJ 原创",
			TimeLimit:       1000,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
			Content:         "# 零钱兑换\n\n给定硬币面额和总金额，求最少硬币数，不可达则输出 -1。\n",
			Constraints:     "1 <= amount <= 1e4。",
			Editorial:       "经典完全背包 / 最短路式 DP。",
			Samples: []models.ProblemSample{
				{CaseNo: 1, Input: "3 11\n1 2 5\n", Expected: "3", Explanation: "11 = 5 + 5 + 1"},
				{CaseNo: 2, Input: "1 3\n2\n", Expected: "-1", Explanation: "无法凑出 3"},
			},
			TestCases: []models.ProblemTestCase{
				{CaseNo: 1, Input: "3 11\n1 2 5\n", Expected: "3", IsHidden: false},
				{CaseNo: 2, Input: "1 3\n2\n", Expected: "-1", IsHidden: false},
				{CaseNo: 3, Input: "1 0\n1\n", Expected: "0", IsHidden: true},
			},
			Templates: defaultTemplates(),
		},
		{
			ID:              1005,
			Title:           "岛屿数量",
			Difficulty:      "中等",
			DifficultyScore: 1200,
			Rating:          1200,
			Tags:            models.StringSlice{"搜索", "图论", "并查集"},
			Source:          "TerminalOJ 原创",
			TimeLimit:       1200,
			MemoryLimit:     256,
			OutputLimitKB:   1024,
			Content:         "# 岛屿数量\n\n给定 01 矩阵，统计连通陆地块数量。\n",
			Constraints:     "m,n <= 300。",
			Editorial:       "DFS/BFS/并查集均可，注意访问标记。",
			Samples: []models.ProblemSample{
				{CaseNo: 1, Input: "3 3\n110\n110\n001\n", Expected: "2", Explanation: "左上角一块，右下角一块"},
				{CaseNo: 2, Input: "1 5\n00000\n", Expected: "0", Explanation: "没有陆地"},
			},
			TestCases: []models.ProblemTestCase{
				{CaseNo: 1, Input: "3 3\n110\n110\n001\n", Expected: "2", IsHidden: false},
				{CaseNo: 2, Input: "1 5\n00000\n", Expected: "0", IsHidden: false},
				{CaseNo: 3, Input: "2 2\n11\n11\n", Expected: "1", IsHidden: true},
			},
			Templates: defaultTemplates(),
		},
	}

	for _, item := range problems {
		version := models.ProblemVersion{
			ProblemID:       item.ID,
			VersionNo:       1,
			Title:           item.Title,
			Difficulty:      item.Difficulty,
			DifficultyScore: item.DifficultyScore,
			Tags:            item.Tags,
			Content:         item.Content,
			Constraints:     item.Constraints,
			Source:          item.Source,
			TimeLimit:       item.TimeLimit,
			MemoryLimit:     item.MemoryLimit,
			OutputLimitKB:   item.OutputLimitKB,
			Editorial:       item.Editorial,
			CreatedBy:       uint64Ptr(adminID),
			PublishedAt:     &now,
		}
		if err := conn.Create(&version).Error; err != nil {
			return err
		}

		for i := range item.Samples {
			item.Samples[i].VersionID = version.ID
		}
		for i := range item.TestCases {
			item.TestCases[i].VersionID = version.ID
		}
		for i := range item.Templates {
			item.Templates[i].VersionID = version.ID
		}

		problem := models.Problem{
			ID:                 item.ID,
			Title:              item.Title,
			Difficulty:         item.Difficulty,
			DifficultyScore:    item.DifficultyScore,
			Tags:               item.Tags,
			Source:             item.Source,
			Status:             models.ProblemStatusPublished,
			CurrentVersionID:   &version.ID,
			PublishedVersionID: &version.ID,
			PublishedAt:        &now,
			PublishedBy:        uint64Ptr(adminID),
			LastEditedBy:       uint64Ptr(adminID),
		}
		if err := conn.Create(&problem).Error; err != nil {
			return err
		}
		if len(item.Samples) > 0 {
			if err := conn.Create(&item.Samples).Error; err != nil {
				return err
			}
		}
		if len(item.TestCases) > 0 {
			if err := conn.Create(&item.TestCases).Error; err != nil {
				return err
			}
		}
		if len(item.Templates) > 0 {
			if err := conn.Create(&item.Templates).Error; err != nil {
				return err
			}
		}
	}

	log.Printf("[seed] %d versioned problems inserted", len(problems))

	// Set auto-increment counter to start after the highest seeded ID
	conn.Exec("ALTER TABLE problems AUTO_INCREMENT = 2000")

	return nil
}

func seedAdditionalProblems(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.Problem{}).Count(&count)
	if count >= 50 {
		return nil // Already have enough problems
	}

	now := time.Now().UTC()
	adminID := seededAdminID(conn)
	problems := additionalProblems()

	for _, item := range problems {
		// Skip if problem already exists
		var existing models.Problem
		if err := conn.Where("id = ?", item.ID).First(&existing).Error; err == nil {
			continue
		}

		version := models.ProblemVersion{
			ProblemID:       item.ID,
			VersionNo:       1,
			Title:           item.Title,
			Difficulty:      item.Difficulty,
			DifficultyScore: item.DifficultyScore,
			Tags:            item.Tags,
			Content:         item.Content,
			Constraints:     item.Constraints,
			Source:          item.Source,
			TimeLimit:       item.TimeLimit,
			MemoryLimit:     item.MemoryLimit,
			OutputLimitKB:   item.OutputLimitKB,
			Editorial:       item.Editorial,
			CreatedBy:       uint64Ptr(adminID),
			PublishedAt:     &now,
		}
		if err := conn.Create(&version).Error; err != nil {
			return err
		}

		for i := range item.Samples {
			item.Samples[i].VersionID = version.ID
		}
		for i := range item.TestCases {
			item.TestCases[i].VersionID = version.ID
		}
		for i := range item.Templates {
			item.Templates[i].VersionID = version.ID
		}

		problem := models.Problem{
			ID:                 item.ID,
			Title:              item.Title,
			Difficulty:         item.Difficulty,
			DifficultyScore:    item.DifficultyScore,
			Rating:             item.Rating,
			Tags:               item.Tags,
			Source:             item.Source,
			Status:             models.ProblemStatusPublished,
			CurrentVersionID:   &version.ID,
			PublishedVersionID: &version.ID,
			PublishedAt:        &now,
			PublishedBy:        uint64Ptr(adminID),
			LastEditedBy:       uint64Ptr(adminID),
		}
		if err := conn.Create(&problem).Error; err != nil {
			return err
		}
		if len(item.Samples) > 0 {
			if err := conn.Create(&item.Samples).Error; err != nil {
				return err
			}
		}
		if len(item.TestCases) > 0 {
			if err := conn.Create(&item.TestCases).Error; err != nil {
				return err
			}
		}
		if len(item.Templates) > 0 {
			if err := conn.Create(&item.Templates).Error; err != nil {
				return err
			}
		}
	}

	log.Printf("[seed] %d additional problems inserted", len(problems))

	// Update auto-increment counter
	conn.Exec("ALTER TABLE problems AUTO_INCREMENT = 2000")

	return nil
}

func seededAdminID(conn *gorm.DB) uint64 {
	var admin models.User
	if err := conn.Where("username = ?", "admin").First(&admin).Error; err == nil {
		return admin.ID
	}
	return 0
}

func uint64Ptr(v uint64) *uint64 {
	if v == 0 {
		return nil
	}
	return &v
}

func defaultTemplates() []models.ProblemTemplate {
	return models.DefaultTemplates()
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

func seedStudyPlans(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.StudyPlan{}).Count(&count)
	if count > 0 {
		return nil
	}

	plans := []models.StudyPlan{
		{
			Title:       "哈希与字符串入门",
			Description: "适合刚开始刷题的用户，覆盖哈希表、字符串和基础动态规划。",
			Difficulty:  "简单",
			Tags:        models.StringSlice{"哈希表", "字符串", "动态规划"},
		},
		{
			Title:       "图搜索与进阶结构",
			Description: "围绕搜索、图论和堆结构的练习题单。",
			Difficulty:  "中等",
			Tags:        models.StringSlice{"搜索", "图论", "堆"},
		},
	}
	if err := conn.Create(&plans).Error; err != nil {
		return err
	}

	items := []models.StudyPlanItem{
		{PlanID: plans[0].ID, ProblemID: 1001, OrderNo: 1, Title: "两数之和", Difficulty: "简单"},
		{PlanID: plans[0].ID, ProblemID: 1002, OrderNo: 2, Title: "最长回文子串", Difficulty: "中等"},
		{PlanID: plans[0].ID, ProblemID: 1004, OrderNo: 3, Title: "零钱兑换", Difficulty: "中等"},
		{PlanID: plans[1].ID, ProblemID: 1005, OrderNo: 1, Title: "岛屿数量", Difficulty: "中等"},
		{PlanID: plans[1].ID, ProblemID: 1003, OrderNo: 2, Title: "合并 K 个升序链表", Difficulty: "困难"},
	}
	return conn.Create(&items).Error
}

func seedDailyChallenges(conn *gorm.DB) error {
	challenges := []models.DailyChallenge{
		{ProblemID: 1002, Title: "最长回文子串", Difficulty: "中等", Date: "2026-06-09"},
		{ProblemID: 1004, Title: "零钱兑换", Difficulty: "中等", Date: "2026-06-10"},
		{ProblemID: 1005, Title: "岛屿数量", Difficulty: "中等", Date: "2026-06-11"},
		{ProblemID: 1001, Title: "两数之和", Difficulty: "简单", Date: "2026-06-12"},
		{ProblemID: 1003, Title: "合并K个升序链表", Difficulty: "困难", Date: "2026-06-13"},
	}
	for _, ch := range challenges {
		var count int64
		conn.Model(&models.DailyChallenge{}).Where("date = ?", ch.Date).Count(&count)
		if count == 0 {
			conn.Create(&ch)
		}
	}
	return nil
}

func seedRatingHistory(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.RatingHistory{}).Where("reason = ?", "AC").Count(&count)
	if count > 0 {
		return nil
	}

	// Get user IDs
	var coder, admin models.User
	conn.Where("username = ?", "coder_test").First(&coder)
	conn.Where("username = ?", "admin").First(&admin)

	if coder.ID == 0 || admin.ID == 0 {
		return nil
	}

	now := time.Now()
	history := []models.RatingHistory{
		// coder_test: 1000 → 1120 → 1250 → 1320
		{UserID: coder.ID, OldRating: 1000, NewRating: 1120, Delta: 120, ProblemID: 1001, Reason: "AC", CreatedAt: now.Add(-30 * 24 * time.Hour)},
		{UserID: coder.ID, OldRating: 1120, NewRating: 1250, Delta: 130, ProblemID: 1002, Reason: "AC", CreatedAt: now.Add(-20 * 24 * time.Hour)},
		{UserID: coder.ID, OldRating: 1250, NewRating: 1320, Delta: 70, ProblemID: 1004, Reason: "AC", CreatedAt: now.Add(-10 * 24 * time.Hour)},
		// admin: 1000 → 1200 → 1400 → 1600
		{UserID: admin.ID, OldRating: 1000, NewRating: 1200, Delta: 200, ProblemID: 1001, Reason: "AC", CreatedAt: now.Add(-45 * 24 * time.Hour)},
		{UserID: admin.ID, OldRating: 1200, NewRating: 1400, Delta: 200, ProblemID: 1003, Reason: "AC", CreatedAt: now.Add(-30 * 24 * time.Hour)},
		{UserID: admin.ID, OldRating: 1400, NewRating: 1600, Delta: 200, ProblemID: 1005, Reason: "AC", CreatedAt: now.Add(-15 * 24 * time.Hour)},
	}

	if err := conn.Create(&history).Error; err != nil {
		return err
	}
	log.Println("[seed] rating history seeded")
	return nil
}

func seedSubmissions(conn *gorm.DB) error {
	var count int64
	conn.Model(&models.Submission{}).Where("trace_id LIKE ?", "seed-%").Count(&count)
	if count > 0 {
		return nil
	}

	var coder, admin models.User
	conn.Where("username = ?", "coder_test").First(&coder)
	conn.Where("username = ?", "admin").First(&admin)
	if coder.ID == 0 || admin.ID == 0 {
		return nil
	}

	now := time.Now()

	type seedSub struct {
		ProblemID    uint64
		ProblemTitle string
		UserID       uint64
		TraceID      string
		Status       string
		Runtime      int
		Memory       string
		ErrorMessage string
		CreatedAt    time.Time
		Code         string
	}

	seeds := []seedSub{
		// coder_test submissions
		{UserID: coder.ID, ProblemID: 1001, ProblemTitle: "两数之和", TraceID: "seed-wa-1", Status: "Wrong Answer", Runtime: 0, Memory: "0", ErrorMessage: "期望输出 0 1，实际输出 1 2", CreatedAt: now.Add(-32 * 24 * time.Hour), Code: "#include<bits/stdc++.h>\nint main(){int n,t;std::cin>>n>>t;std::vector<int>a(n);for(auto&x:a)std::cin>>x;for(int i=0;i<n;i++)for(int j=i+1;j<n;j++)if(a[i]+a[j]==t){std::cout<<i<<\" \"<<j;return 0;}}"},
		{UserID: coder.ID, ProblemID: 1001, ProblemTitle: "两数之和", TraceID: "seed-ac-1", Status: "Accepted", Runtime: 12, Memory: "15.2", CreatedAt: now.Add(-30 * 24 * time.Hour), Code: "#include<bits/stdc++.h>\nint main(){int n,t;std::cin>>n>>t;std::unordered_map<int,int>m;for(int i=0,x;i<n;i++){std::cin>>x;if(m.count(t-x)){std::cout<<m[t-x]<<\" \"<<i;return 0;}m[x]=i;}}"},
		{UserID: coder.ID, ProblemID: 1002, ProblemTitle: "最长回文子串", TraceID: "seed-ac-2", Status: "Accepted", Runtime: 28, Memory: "12.8", CreatedAt: now.Add(-20 * 24 * time.Hour), Code: "#include<bits/stdc++.h>\nstd::string longest(std::string s){int n=s.size(),start=0,maxlen=1;std::vector<std::vector<bool>>dp(n,std::vector<bool>(n));for(int i=0;i<n;i++)dp[i][i]=true;for(int len=2;len<=n;len++)for(int i=0;i+len<=n;i++){int j=i+len-1;if(s[i]==s[j]&&(len==2||dp[i+1][j-1])){dp[i][j]=true;if(len>maxlen){maxlen=len;start=i;}}}return s.substr(start,maxlen);}\nint main(){std::string s;std::cin>>s;std::cout<<longest(s);}"},
		{UserID: coder.ID, ProblemID: 1004, ProblemTitle: "零钱兑换", TraceID: "seed-ac-3", Status: "Accepted", Runtime: 8, Memory: "8.4", CreatedAt: now.Add(-10 * 24 * time.Hour), Code: "#include<bits/stdc++.h>\nint main(){int n,amount;std::cin>>n>>amount;std::vector<int>coins(n);for(auto&c:coins)std::cin>>c;std::vector<int>dp(amount+1,1e9);dp[0]=0;for(int i=1;i<=amount;i++)for(int c:coins)if(i>=c)dp[i]=std::min(dp[i],dp[i-c]+1);std::cout<<(dp[amount]>=1e9?-1:dp[amount]);}"},
		// admin submissions
		{UserID: admin.ID, ProblemID: 1001, ProblemTitle: "两数之和", TraceID: "seed-ac-4", Status: "Accepted", Runtime: 10, Memory: "14.8", CreatedAt: now.Add(-45 * 24 * time.Hour), Code: "#include<bits/stdc++.h>\nint main(){int n,t;std::cin>>n>>t;std::unordered_map<int,int>m;for(int i=0,x;i<n;i++){std::cin>>x;if(m.count(t-x)){std::cout<<m[t-x]<<\" \"<<i;return 0;}m[x]=i;}}"},
		{UserID: admin.ID, ProblemID: 1003, ProblemTitle: "合并 K 个升序链表", TraceID: "seed-ac-5", Status: "Accepted", Runtime: 45, Memory: "32.1", CreatedAt: now.Add(-30 * 24 * time.Hour), Code: "// merge k sorted lists using priority queue"},
		{UserID: admin.ID, ProblemID: 1005, ProblemTitle: "岛屿数量", TraceID: "seed-ac-6", Status: "Accepted", Runtime: 20, Memory: "18.5", CreatedAt: now.Add(-15 * 24 * time.Hour), Code: "// island count using DFS"},
	}

	subs := make([]models.Submission, 0, len(seeds))
	for _, s := range seeds {
		fin := s.CreatedAt.Add(time.Duration(100+s.Runtime) * time.Millisecond)
		subs = append(subs, models.Submission{
			UserID:         s.UserID,
			ProblemID:      s.ProblemID,
			ProblemTitle:   s.ProblemTitle,
			TraceID:        s.TraceID,
			Source:         "submit",
			Language:       "cpp",
			Code:           s.Code,
			CodeLength:     len(s.Code),
			Status:         s.Status,
			Runtime:        s.Runtime,
			RuntimeMS:      s.Runtime,
			Memory:         s.Memory,
			QueueStartedAt: &s.CreatedAt,
			FinishedAt:     &fin,
			CreatedAt:      s.CreatedAt,
			UpdatedAt:      s.CreatedAt,
		})
	}

	if err := conn.Create(&subs).Error; err != nil {
		return err
	}

	// Set auto-increment to start at 200000, well above the sequence range (100000+)
	// to prevent any collision between auto-increment and nextSubmissionID()
	conn.Exec("ALTER TABLE submissions AUTO_INCREMENT = 200000")
	log.Printf("[seed] %d seed submissions created", len(subs))
	return nil
}
