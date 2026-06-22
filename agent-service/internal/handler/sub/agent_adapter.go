package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AgentAdapter 封装了原 main.go 中的 Agent 功能
type AgentAdapter struct {
	knowledgeDB   *KnowledgeDB
	recordsFile   string
	knowledgeFile string
	embeddingFile string
}

// NewAgentAdapter 创建 Agent 适配器
func NewAgentAdapter(knowledgeFile, embeddingFile, recordsFile string) (*AgentAdapter, error) {
	adapter := &AgentAdapter{
		knowledgeFile: knowledgeFile,
		embeddingFile: embeddingFile,
		recordsFile:   recordsFile,
	}

	// 检查并初始化知识库
	if err := adapter.ensureKnowledgeDB(); err != nil {
		return nil, fmt.Errorf("初始化知识库失败: %v", err)
	}

	// 加载知识库
	db, err := LoadKnowledgeWithEmbedding(embeddingFile)
	if err != nil {
		return nil, fmt.Errorf("加载知识库失败: %v", err)
	}
	adapter.knowledgeDB = db

	return adapter, nil
}

// ensureKnowledgeDB 确保知识库已初始化
func (a *AgentAdapter) ensureKnowledgeDB() error {
	// 检查 embedding 文件是否存在
	if _, err := os.Stat(a.embeddingFile); os.IsNotExist(err) {
		// 检查原始知识库文件
		if _, err := os.Stat(a.knowledgeFile); os.IsNotExist(err) {
			return fmt.Errorf("知识库文件不存在: %s", a.knowledgeFile)
		}

		// 加载原始知识库并计算向量
		db, err := LoadKnowledge(a.knowledgeFile)
		if err != nil {
			return err
		}

		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(a.embeddingFile), 0755); err != nil {
			return err
		}

		// 预计算向量
		if err := PrecomputeEmbeddings(db, a.embeddingFile); err != nil {
			return err
		}
	}
	return nil
}

// GenerateTutorial 生成辅导输出
func (a *AgentAdapter) GenerateTutorial(cfg StartConfig) (*TutorialOutput, error) {
	// 构建 Submission
	submission := Submission{
		Code:         cfg.Code,
		Language:     cfg.Language,
		ProblemTitle: cfg.ProblemTitle,
		ProblemDesc:  cfg.ProblemDesc,
	}

	// 构建 JudgeResult
	judgeResult := JudgeResult{}

	// 创建 Agent 并运行
	agent := DeepAgent{
		Memory:      []Memory{},
		Submission:  submission,
		JudgeResult: judgeResult,
		KnowledgeDB: a.knowledgeDB,
	}

	output, err := agent.Run()
	if err != nil {
		return nil, err
	}

	// 保存记录
	if cfg.UserID != "" && cfg.ProblemTitle != "" {
		_ = SaveRecord(cfg.UserID, cfg.ProblemTitle)
	}

	return output, nil
}

// Close 释放资源
func (a *AgentAdapter) Close() error {
	return nil
}

// ConvertToStartConfig 将 GenerateSolutionPayload 转换为 StartConfig
func ConvertToStartConfig(req *GenerateSolutionPayload) StartConfig {
	return StartConfig{
		UserID:       "agent_user", // 可以从上下文获取
		Code:         req.Code,
		Language:     req.Language,
		ProblemTitle: req.ProblemTitle,
		ProblemDesc:  req.ProblemDesc,
	}
}

// ConvertTutorialToSolutionResponse 将 TutorialOutput 转换为题解响应格式
func ConvertTutorialToSolutionResponse(tutorial *TutorialOutput, rawResponse string) map[string]interface{} {
	// 构建题解内容
	content := fmt.Sprintf(`## 一、问题分析
%s

## 二、原因分析
%s

## 三、改进方向
%s

## 四、拓展学习
%s

## 五、总结
本题核心问题已定位，通过上述改进可以提升代码质量。`,
		tutorial.ProblemLocate,
		tutorial.ReasonAnalysis,
		tutorial.ImproveDirection,
		tutorial.ExtendLearn,
	)

	// 提取算法标签
	tags := extractTagsFromContent(tutorial.ProblemLocate + " " + tutorial.ReasonAnalysis)

	// 格式化时间复杂度和空间复杂度（添加 Markdown 加粗）
	timeComplexity := tutorial.TimeComplexity
	spaceComplexity := tutorial.SpaceComplexity
	if timeComplexity != "" && !strings.HasPrefix(timeComplexity, "**") {
		timeComplexity = "**" + timeComplexity + "**"
	}
	if spaceComplexity != "" && !strings.HasPrefix(spaceComplexity, "**") {
		spaceComplexity = "**" + spaceComplexity + "**"
	}

	result := map[string]interface{}{
		"title":         "题解分析报告",
		"content":       content,
		"algorithmTags": tags,
		"complexity": map[string]string{
			"time":  timeComplexity,
			"space": spaceComplexity,
		},
		"rawMarkdown": rawResponse,
		"provider":    "agent-service",
		"confidence":  tutorial.Confidence,
	}

	return result
}

// extractTagsFromContent 从内容中提取算法标签
func extractTagsFromContent(content string) []string {
	// 简单实现：检查常见标签
	allTags := strings.Split(CandidateTagDict, "、")
	foundTags := []string{}

	for _, tag := range allTags {
		tag = strings.TrimSpace(tag)
		if tag != "" && strings.Contains(strings.ToLower(content), strings.ToLower(tag)) {
			foundTags = append(foundTags, tag)
			if len(foundTags) >= 5 {
				break
			}
		}
	}

	if len(foundTags) == 0 {
		foundTags = []string{"算法分析"}
	}
	return foundTags
}
