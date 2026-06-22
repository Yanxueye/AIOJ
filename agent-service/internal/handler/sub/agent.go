package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// 配置
const (
	DashScopeAPIKey = "sk-ed18064f710b4eeda5087292a6869439"

	ChatAPIURL = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	ChatModel  = "qwen3.6-flash"

	EmbeddingAPIURL = "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
	EmbeddingModel  = "text-embedding-v4"

	KnowledgeFile        = "knowledge.json"
	KnowledgeWithEmbFile = "knowledge_with_embedding.json"

	TopK          = 3
	KeywordWeight = 0.6
	VectorWeight  = 0.4

	TimeoutSec = 180
)

// Agent 结构

type Memory struct {
	Round   int
	Role    string
	Content string
}

type DeepAgent struct {
	Memory      []Memory
	Submission  Submission
	JudgeResult JudgeResult
	KnowledgeDB *KnowledgeDB
}

type HybridResult struct {
	ID         int
	Item       KnowledgeItem
	FinalScore float64
}

// 辅导输出结构

type TutorialOutput struct {
	ProblemLocate    string  `json:"problem_locate"`
	ReasonAnalysis   string  `json:"reason_analysis"`
	ImproveDirection string  `json:"improve_direction"`
	ExtendLearn      string  `json:"extend_learn"`
	TimeComplexity   string  `json:"time_complexity"`
	SpaceComplexity  string  `json:"space_complexity"`
	Confidence       float64 `json:"confidence"`
}

// Embedding API

func GetEmbedding(text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"model":      EmbeddingModel,
		"input":      map[string]interface{}{"texts": []string{text}},
		"parameters": map[string]string{"text_type": "query"},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", EmbeddingAPIURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DashScopeAPIKey)

	client := &http.Client{Timeout: TimeoutSec * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Output struct {
			Embeddings []struct {
				Embedding []float64 `json:"embedding"`
			} `json:"embeddings"`
		} `json:"output"`
	}
	json.Unmarshal(body, &result)

	if len(result.Output.Embeddings) == 0 {
		return nil, fmt.Errorf("没有返回向量")
	}
	return result.Output.Embeddings[0].Embedding, nil
}

// 预计算向量

func PrecomputeEmbeddings(db *KnowledgeDB, savePath string) error {
	fmt.Printf("预计算向量...\n")
	for i := range db.Topics {
		item := &db.Topics[i]
		text := item.Name + " " + item.Category + " " + item.Content + " " + strings.Join(item.Keywords, " ")
		fmt.Printf("  [%d/%d] %s\n", i+1, len(db.Topics), item.Name)

		embedding, err := GetEmbedding(text)
		if err != nil {
			return err
		}
		item.Embedding = embedding
		time.Sleep(500 * time.Millisecond)
	}

	data, _ := json.MarshalIndent(db, "", "  ")
	os.WriteFile(savePath, data, 0644)
	fmt.Printf("✅ 已保存到 %s\n", savePath)
	return nil
}

// 双通道检索

func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func KeywordSearch(db *KnowledgeDB, query string) []HybridResult {
	query = strings.ToLower(query)
	results := []HybridResult{}

	for _, item := range db.Topics {
		score := 0
		if strings.Contains(strings.ToLower(item.Name), query) {
			score += 20
		}
		if strings.Contains(strings.ToLower(item.Category), query) {
			score += 10
		}
		for _, kw := range item.Keywords {
			if strings.Contains(query, strings.ToLower(kw)) {
				score += 5
			}
		}
		if strings.Contains(strings.ToLower(item.Content), query) {
			score += 2
		}
		if score > 0 {
			results = append(results, HybridResult{
				ID:         item.ID,
				Item:       item,
				FinalScore: float64(score) / 100.0,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].FinalScore > results[j].FinalScore })
	return results
}

func VectorSearch(db *KnowledgeDB, query string) ([]HybridResult, error) {
	queryEmbedding, err := GetEmbedding(query)
	if err != nil {
		return nil, err
	}

	results := []HybridResult{}
	for _, item := range db.Topics {
		if len(item.Embedding) == 0 {
			continue
		}
		similarity := CosineSimilarity(queryEmbedding, item.Embedding)
		if similarity > 0.3 {
			results = append(results, HybridResult{
				ID:         item.ID,
				Item:       item,
				FinalScore: similarity,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].FinalScore > results[j].FinalScore })
	return results, nil
}

func HybridSearch(db *KnowledgeDB, query string) ([]HybridResult, error) {
	keywordResults := KeywordSearch(db, query)

	vectorResults, err := VectorSearch(db, query)
	if err != nil {
		if len(keywordResults) > TopK {
			return keywordResults[:TopK], nil
		}
		return keywordResults, nil
	}

	scoreMap := make(map[int]*HybridResult)
	for i := range keywordResults {
		scoreMap[keywordResults[i].ID] = &keywordResults[i]
	}
	for i := range vectorResults {
		if existing, ok := scoreMap[vectorResults[i].ID]; ok {
			existing.FinalScore = existing.FinalScore*KeywordWeight + vectorResults[i].FinalScore*VectorWeight
		} else {
			scoreMap[vectorResults[i].ID] = &vectorResults[i]
		}
	}

	results := []HybridResult{}
	for _, v := range scoreMap {
		results = append(results, *v)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].FinalScore > results[j].FinalScore })

	if len(results) > TopK {
		results = results[:TopK]
	}
	return results, nil
}

// AI 调用

func CallLLM(systemPrompt, userPrompt string, jsonMode bool) (string, error) {
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := callLLMOnce(systemPrompt, userPrompt, jsonMode)
		if err == nil {
			return result, nil
		}
		if strings.Contains(err.Error(), "429") ||
			strings.Contains(err.Error(), "Too Many Requests") ||
			strings.Contains(err.Error(), "timeout") {
			waitTime := time.Duration(attempt+1) * 2 * time.Second
			fmt.Printf("  ⚠️ 请求失败 (%v)，%v 后重试...\n", err, waitTime)
			time.Sleep(waitTime)
			continue
		}
		return "", err
	}
	return "", fmt.Errorf("重试 %d 次后仍失败", maxRetries)
}

func callLLMOnce(systemPrompt, userPrompt string, jsonMode bool) (string, error) {
	reqBody := map[string]interface{}{
		"model": ChatModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"stream":     false,
		"max_tokens": 800,
	}

	if jsonMode {
		reqBody["response_format"] = map[string]string{
			"type": "json_object",
		}
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", ChatAPIURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DashScopeAPIKey)

	client := &http.Client{Timeout: TimeoutSec * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return "", fmt.Errorf("API错误 [%s]: %s", errResp.Error.Code, errResp.Error.Message)
		}
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(body, &result)

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API返回空结果")
	}
	return result.Choices[0].Message.Content, nil
}

// 解析 JSON 输出

func ParseTutorialOutput(response string) (*TutorialOutput, error) {
	var output TutorialOutput
	err := json.Unmarshal([]byte(response), &output)
	if err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}
	return &output, nil
}

// ParseTutorialOutputEnhanced 增强版解析，自动清理 Markdown 和提取 JSON
func ParseTutorialOutputEnhanced(response string) (*TutorialOutput, error) {

	cleaned := strings.TrimSpace(response)

	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start != -1 && end != -1 && end > start {
		cleaned = cleaned[start : end+1]
	}

	var output TutorialOutput
	err := json.Unmarshal([]byte(cleaned), &output)
	if err != nil {
		fmt.Printf("⚠️ JSON解析失败，原始响应前500字符:\n%s\n", truncate(response, 500))
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}
	return &output, nil
}

// Deep Agent

func buildRound1Prompt(title, desc, code string) string {
	return fmt.Sprintf(`理解题目和代码：
题目：%s
描述：%s
代码：%s
回答：1.核心考点 2.用户方法 3.主要逻辑（300字内）`, title, desc, code)
}

func buildRound2Prompt(prevResult, knowledge string) string {
	return fmt.Sprintf(`上一步：%s
知识：%s
分析：1.核心考点 2.用户方法 3.主要逻辑（400字内）`,
		prevResult, knowledge)
}

func buildRound3Prompt(prevResult string) string {
	return fmt.Sprintf(`分析：%s
方案：1.修改方法 2.修改后复杂度 3.更好解法（400字内）`, prevResult)
}

func buildRound4Prompt(r1, r2, r3 string) string {
	return fmt.Sprintf(`理解：%s
分析：%s
方案：%s
请按以下 JSON 格式输出最终辅导，只输出 JSON，不要其他内容：
{
    "problem_locate": "问题定位",
    "reason_analysis": "原因分析",
    "improve_direction": "改进方向",
    "extend_learn": "拓展学习（相关知识点或类似题目）",
    "time_complexity": "时间复杂度，格式如 O(n)",
    "space_complexity": "空间复杂度，格式如 O(1)",
    "confidence": 0.95
}`, r1, r2, r3)
}

func RetrieveKnowledge(db *KnowledgeDB, problemTitle string) ([]HybridResult, string) {
	query := problemTitle + " 知识点 算法"

	knowledge, err := HybridSearch(db, query)
	if err != nil || len(knowledge) == 0 {
		return []HybridResult{}, ""
	}

	context := ""
	for i, k := range knowledge {
		context += fmt.Sprintf("\n【知识%d】%s\n- 核心：%s\n", i+1, k.Item.Name, k.Item.Content)
	}
	return knowledge, context
}

func (a *DeepAgent) Run() (*TutorialOutput, error) {
	fmt.Println("\n=== Deep Agent 启动 ===")

	_, knowledgeContext := RetrieveKnowledge(a.KnowledgeDB, a.Submission.ProblemTitle)

	var r1, r2, r3 string

	fmt.Print("Round 1...")
	p1 := buildRound1Prompt(a.Submission.ProblemTitle, a.Submission.ProblemDesc, a.Submission.Code)
	r1, _ = CallLLM("", p1, false)
	fmt.Println(" 完成")

	fmt.Print("Round 2...")
	p2 := buildRound2Prompt(r1, knowledgeContext)
	r2, _ = CallLLM("", p2, false)
	fmt.Println(" 完成")

	fmt.Print("Round 3...")
	p3 := buildRound3Prompt(r2)
	r3, _ = CallLLM("", p3, false)
	fmt.Println(" 完成")

	fmt.Print("Round 4...")
	p4 := buildRound4Prompt(r1, r2, r3)
	response, err := CallLLM("", p4, true)
	if err != nil {
		fmt.Println(" 失败")
		return nil, err
	}
	fmt.Println(" 完成")

	output, err := ParseTutorialOutputEnhanced(response)
	if err != nil {
		output, err2 := ParseTutorialOutput(response)
		if err2 != nil {
			return &TutorialOutput{
				ProblemLocate:    "解析失败",
				ReasonAnalysis:   response,
				ImproveDirection: "请查看原始输出",
				ExtendLearn:      "",
				TimeComplexity:   "O(n)",
				SpaceComplexity:  "O(1)",
				Confidence:       0.5,
			}, nil
		}
		return output, nil
	}
	return output, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
