package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// 大模型推荐

func RecommendByLLM(db *KnowledgeDB, userID string) error {
	records, err := LoadRecords(userID)
	if err != nil {
		return fmt.Errorf("加载历史失败: %v", err)
	}

	if len(records) == 0 {
		fmt.Println("推荐: two_sum, binary_search")
		return nil
	}

	// 构建历史摘要
	var history strings.Builder
	for _, r := range records {
		status := "AC"
		if r.Status != "AC" && r.Status != "Accepted" && r.Status != "正确" && r.Status != "✅" {
			status = "WA"
		}
		history.WriteString(fmt.Sprintf("- %s: %s\n", r.QuestionID, status))
	}

	// 构建知识点列表
	var kpList strings.Builder
	for _, kp := range db.Topics {
		kpList.WriteString(fmt.Sprintf("- ID:%d, %s\n", kp.ID, kp.Name))
	}

	// 第一次调用：薄弱知识点
	systemPrompt1 := `你是算法学习规划师。根据用户做题记录，从知识库中选出3个薄弱知识点。
只输出JSON：{"kp_ids": [1, 2, 3]}`

	userPrompt1 := fmt.Sprintf("知识库：\n%s\n用户历史：\n%s", kpList.String(), history.String())

	fmt.Print("分析薄弱知识点...")
	response1, err := CallLLM(systemPrompt1, userPrompt1, true)
	if err != nil {
		return fmt.Errorf("知识点分析失败: %v", err)
	}

	var kpResult struct {
		KpIDs []int `json:"kp_ids"`
	}
	if err := json.Unmarshal([]byte(response1), &kpResult); err != nil || len(kpResult.KpIDs) == 0 {
		fmt.Println(" 使用默认推荐")
		fmt.Println("推荐: two_sum, binary_search")
		return nil
	}
	fmt.Println(" 完成")

	// 图谱扩展
	priorityMap := make(map[int]int)
	for i, id := range kpResult.KpIDs {
		priorityMap[id] = i + 1
	}

	type kpWithPriority struct {
		ID       int
		Priority int
		Name     string
	}
	allKP := []kpWithPriority{}

	for _, id := range kpResult.KpIDs {
		allKP = append(allKP, kpWithPriority{ID: id, Priority: priorityMap[id], Name: getKPName(id, db)})
	}

	for _, id := range kpResult.KpIDs {
		kp := getKPByID(id, db)
		if kp == nil {
			continue
		}
		basePriority := priorityMap[id]
		for _, preID := range kp.Prerequisites {
			if _, exists := priorityMap[preID]; !exists {
				priorityMap[preID] = basePriority + 10
				allKP = append(allKP, kpWithPriority{ID: preID, Priority: basePriority + 10, Name: getKPName(preID, db)})
			}
		}
		for _, relID := range kp.Related {
			if _, exists := priorityMap[relID]; !exists {
				priorityMap[relID] = basePriority + 20
				allKP = append(allKP, kpWithPriority{ID: relID, Priority: basePriority + 20, Name: getKPName(relID, db)})
			}
		}
		for _, varID := range kp.Variants {
			if _, exists := priorityMap[varID]; !exists {
				priorityMap[varID] = basePriority + 30
				allKP = append(allKP, kpWithPriority{ID: varID, Priority: basePriority + 30, Name: getKPName(varID, db)})
			}
		}
	}

	sort.Slice(allKP, func(i, j int) bool { return allKP[i].Priority < allKP[j].Priority })
	if len(allKP) > 3 {
		allKP = allKP[:3]
	}

	names := []string{}
	for _, kp := range allKP {
		names = append(names, kp.Name)
	}
	fmt.Printf("薄弱知识点: %s\n", strings.Join(names, "、"))

	// 构建知识点详情
	var kpDetails strings.Builder
	for _, kp := range allKP {
		kpDetail := getKPByID(kp.ID, db)
		if kpDetail == nil {
			continue
		}
		kpDetails.WriteString(fmt.Sprintf("\n【%s】%s\n", kpDetail.Name, kpDetail.Content))
	}

	// 第二次调用：推荐题目
	systemPrompt2 := `根据知识点推荐3道题目。只输出JSON：{"questions": [{"name": "题目名", "reason": "理由"}]}`

	userPrompt2 := fmt.Sprintf("知识点：%s", kpDetails.String())

	fmt.Print("推荐题目...")
	response2, err := CallLLM(systemPrompt2, userPrompt2, true)
	if err != nil {
		return fmt.Errorf("题目推荐失败: %v", err)
	}

	var qResult struct {
		Questions []struct {
			Name   string `json:"name"`
			Reason string `json:"reason"`
		} `json:"questions"`
	}
	if err := json.Unmarshal([]byte(response2), &qResult); err != nil || len(qResult.Questions) == 0 {
		fmt.Println(" 使用默认推荐")
		fmt.Println("推荐: two_sum, binary_search")
		return nil
	}
	fmt.Println(" 完成")

	fmt.Println("\n推荐题目：")
	for i, q := range qResult.Questions {
		fmt.Printf("  %d. %s\n", i+1, q.Name)
		fmt.Printf("     %s\n", q.Reason)
	}

	return nil
}

// 配置结构

// StartConfig 启动配置参数
type StartConfig struct {
	UserID       string
	Code         string
	Language     string
	ProblemTitle string
	ProblemDesc  string
}

// DefaultStartConfig 返回默认配置
func DefaultStartConfig() StartConfig {
	return StartConfig{
		UserID: "student_004",
		Code: `#include<iostream>
using namespace std;
int n; unsigned long long ans=1;
int main() { cin>>n; for(int i=2;i<=n;i++){ ans*=i; while(ans%10000==0)ans/=10000; ans%=10000;} while(ans%10==0)ans/=10; ans%=10; cout<<ans; }`,
		Language:     "cpp",
		ProblemTitle: "阶乘问题",
		ProblemDesc:  "计算 N! 最右边的非零位",
	}
}

//返回结构

// StartResult start函数返回的结果
type StartResult struct {
	ProblemLocate    string  `json:"problem_locate"`
	ReasonAnalysis   string  `json:"reason_analysis"`
	ImproveDirection string  `json:"improve_direction"`
	ExtendLearn      string  `json:"extend_learn"`
	Confidence       float64 `json:"confidence"`
	RecordSaved      bool    `json:"record_saved"`
	Error            string  `json:"error,omitempty"`
}

//主函数

func start(cfg StartConfig) (string, error) {
	result := StartResult{}

	fmt.Println("=== 智能OJ平台 ===")

	if _, err := os.Stat(KnowledgeWithEmbFile); os.IsNotExist(err) {
		fmt.Println("初始化知识库...")
		db, err := LoadKnowledge(KnowledgeFile)
		if err != nil {
			fmt.Printf("加载失败: %v\n", err)
			result.Error = fmt.Sprintf("加载知识库失败: %v", err)
			return marshalResult(result)
		}
		if err := PrecomputeEmbeddings(db, KnowledgeWithEmbFile); err != nil {
			fmt.Printf("向量计算失败: %v\n", err)
			result.Error = fmt.Sprintf("向量计算失败: %v", err)
			return marshalResult(result)
		}
	}

	db, err := LoadKnowledgeWithEmbedding(KnowledgeWithEmbFile)
	if err != nil {
		fmt.Printf("加载知识库失败: %v\n", err)
		result.Error = fmt.Sprintf("加载知识库失败: %v", err)
		return marshalResult(result)
	}
	fmt.Printf("知识库: %d 条\n", len(db.Topics))

	// Agent辅导
	fmt.Println("\n--- Agent辅导 ---")
	submission := Submission{
		Code:         cfg.Code,
		Language:     cfg.Language,
		ProblemTitle: cfg.ProblemTitle,
		ProblemDesc:  cfg.ProblemDesc,
	}
	judgeResult := JudgeResult{}

	agent := DeepAgent{
		Memory:      []Memory{},
		Submission:  submission,
		JudgeResult: judgeResult,
		KnowledgeDB: db,
	}
	output, err := agent.Run()
	if err != nil {
		fmt.Printf("Agent失败: %v\n", err)
		result.Error = fmt.Sprintf("Agent失败: %v", err)
	} else {
		fmt.Println("\n=== 辅导输出 ===")
		fmt.Printf("问题定位: %s\n", output.ProblemLocate)
		fmt.Printf("原因分析: %s\n", output.ReasonAnalysis)
		fmt.Printf("改进方向: %s\n", output.ImproveDirection)
		fmt.Printf("拓展学习: %s\n", output.ExtendLearn)

		// 填充结果
		result.ProblemLocate = output.ProblemLocate
		result.ReasonAnalysis = output.ReasonAnalysis
		result.ImproveDirection = output.ImproveDirection
		result.ExtendLearn = output.ExtendLearn
		result.Confidence = output.Confidence

		if err := SaveRecord(cfg.UserID, submission.ProblemTitle); err == nil {
			fmt.Printf("已保存: %s - %s\n", submission.ProblemTitle)
			result.RecordSaved = true
		}
	}

	/*	// 大模型推荐
		fmt.Println("\n--- 大模型推荐 ---")
		if err := RecommendByLLM(db, cfg.UserID); err != nil {
			fmt.Printf("推荐失败: %v\n", err)
		}

		// 历史记录
		fmt.Println("\n--- 历史记录 ---")
		records, _ := LoadRecords(cfg.UserID)
		fmt.Printf("共 %d 条记录\n", len(records))
	*/
	fmt.Println("\n完成")

	return marshalResult(result)
}

// marshalResult 将结果序列化为JSON
func marshalResult(result StartResult) (string, error) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化结果失败: %v", err)
	}
	return string(jsonData), nil
}

func main() {
	result, err := start(DefaultStartConfig())
	if err != nil {
		fmt.Printf("执行失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n=== 返回结果 ===")
	fmt.Println(result)
}
