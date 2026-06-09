package database

import (
	"fmt"
	"log"
	"time"

	"github.com/terminaloj/backend/internal/config"
	"github.com/terminaloj/backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Init opens the MySQL connection, configures the pool and migrates schema.
func Init(cfg config.MySQLConfig) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	conn, err := gorm.Open(mysql.Open(cfg.DSN), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if cfg.AutoMigrate {
		if err := conn.AutoMigrate(
			&models.User{},
			&models.Problem{},
			&models.ProblemVersion{},
			&models.ProblemSample{},
			&models.ProblemTestCase{},
			&models.ProblemTemplate{},
			&models.RejudgeJob{},
			&models.AuditLog{},
			&models.Favorite{},
			&models.ProblemSolution{},
			&models.StudyPlan{},
			&models.StudyPlanItem{},
			&models.UserPlanProgress{},
			&models.UserPlanProgressItem{},
			&models.DailyChallenge{},
			&models.StudyCheckin{},
			&models.Submission{},
			&models.SubmissionCaseResult{},
			&models.Announcement{},
			&models.Conversation{},
			&models.Message{},
		); err != nil {
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
		log.Println("[db] auto migration complete")
	}

	db = conn

	if cfg.AutoMigrate {
		if err := backfillLegacyProblems(conn); err != nil {
			log.Printf("[db] legacy problem backfill warn: %v", err)
		}
	}

	if cfg.Seed {
		if err := Seed(conn); err != nil {
			log.Printf("[db] seed warn: %v", err)
		}
	}
	return conn, nil
}

// DB returns the shared *gorm.DB instance.
func DB() *gorm.DB { return db }

func backfillLegacyProblems(conn *gorm.DB) error {
	var rows []models.Problem
	if err := conn.Preload("CurrentVersion").Preload("PublishedVersion").Find(&rows).Error; err != nil {
		return err
	}

	for _, problem := range rows {
		if problem.CurrentVersionID != nil || problem.PublishedVersionID != nil {
			continue
		}

		legacyContent, legacyConstraints, legacyEditorial, legacyTimeLimit, legacyMemoryLimit, legacyOutputLimit, legacySamples, legacyTests, legacyTemplates := loadLegacyProblemSnapshot(conn, problem.ID)
		now := time.Now().UTC()
		version := models.ProblemVersion{
			ProblemID:       problem.ID,
			VersionNo:       1,
			Title:           problem.Title,
			Difficulty:      problem.Difficulty,
			DifficultyScore: problem.DifficultyScore,
			Tags:            problem.Tags,
			Content:         legacyContent,
			Constraints:     legacyConstraints,
			Source:          problem.Source,
			TimeLimit:       legacyTimeLimit,
			MemoryLimit:     legacyMemoryLimit,
			OutputLimitKB:   legacyOutputLimit,
			Editorial:       legacyEditorial,
			PublishedAt:     &now,
		}
		if err := conn.Create(&version).Error; err != nil {
			return err
		}

		for i := range legacySamples {
			legacySamples[i].VersionID = version.ID
		}
		for i := range legacyTests {
			legacyTests[i].VersionID = version.ID
		}
		for i := range legacyTemplates {
			legacyTemplates[i].VersionID = version.ID
		}
		if len(legacySamples) > 0 {
			if err := conn.Create(&legacySamples).Error; err != nil {
				return err
			}
		}
		if len(legacyTests) > 0 {
			if err := conn.Create(&legacyTests).Error; err != nil {
				return err
			}
		}
		if len(legacyTemplates) > 0 {
			if err := conn.Create(&legacyTemplates).Error; err != nil {
				return err
			}
		}

		status := problem.Status
		if status == "" {
			status = models.ProblemStatusPublished
		}
		if status == models.ProblemStatusDraft {
			status = models.ProblemStatusPublished
		}
		if err := conn.Model(&models.Problem{}).Where("id = ?", problem.ID).Updates(map[string]any{
			"status":                status,
			"current_version_id":    version.ID,
			"published_version_id":  version.ID,
			"published_at":          now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func loadLegacyProblemSnapshot(conn *gorm.DB, problemID uint64) (string, string, string, int, int, int32, []models.ProblemSample, []models.ProblemTestCase, []models.ProblemTemplate) {
	type legacyRow struct {
		Content       string
		TimeLimit     int
		MemoryLimit   int
		OutputLimitKB int32
		TestCases     models.TestCases
	}

	var row legacyRow
	// Best-effort read from pre-versioned columns; if they no longer exist this query will simply fail and defaults will be used.
	err := conn.Raw("SELECT content, time_limit, memory_limit, output_limit_kb, test_cases FROM problems WHERE id = ?", problemID).Scan(&row).Error
	if err != nil {
		row.TimeLimit = 1000
		row.MemoryLimit = 256
		row.OutputLimitKB = 1024
	}

	if row.Content == "" {
		row.Content = "# 题目内容暂缺\n\n该题目来自旧版本数据，管理员可在后台补全题面。"
	}
	if row.TimeLimit <= 0 {
		row.TimeLimit = 1000
	}
	if row.MemoryLimit <= 0 {
		row.MemoryLimit = 256
	}
	if row.OutputLimitKB <= 0 {
		row.OutputLimitKB = 1024
	}

	samples := make([]models.ProblemSample, 0, localMin(2, len(row.TestCases)))
	tests := make([]models.ProblemTestCase, 0, len(row.TestCases))
	for i, item := range row.TestCases {
		if i < 2 {
			samples = append(samples, models.ProblemSample{
				CaseNo:      i + 1,
				Input:       item.Input,
				Expected:    item.Expected,
				Explanation: "",
			})
		}
		tests = append(tests, models.ProblemTestCase{
			CaseNo:   i + 1,
			Input:    item.Input,
			Expected: item.Expected,
			IsHidden: i >= 2,
		})
	}
	if len(samples) == 0 {
		samples = append(samples, models.ProblemSample{
			CaseNo:   1,
			Input:    "",
			Expected: "",
		})
	}
	if len(tests) == 0 {
		tests = append(tests, models.ProblemTestCase{
			CaseNo:   1,
			Input:    "",
			Expected: "",
			IsHidden: false,
		})
	}

	return row.Content, "", "", row.TimeLimit, row.MemoryLimit, row.OutputLimitKB, samples, tests, legacyDefaultTemplates()
}

func legacyDefaultTemplates() []models.ProblemTemplate {
	return []models.ProblemTemplate{
		{Language: "cpp", Code: "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    return 0;\n}\n"},
		{Language: "python", Code: "import sys\ninput = sys.stdin.readline\n\ndef solve():\n    pass\n\nsolve()\n"},
		{Language: "go", Code: "package main\n\nfunc main() {\n}\n"},
	}
}

func localMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
