package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//数据结构

type KnowledgeItem struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Category        string    `json:"category"`
	Content         string    `json:"content"`
	TimeComplexity  string    `json:"time_complexity"`
	SpaceComplexity string    `json:"space_complexity"`
	Keywords        []string  `json:"keywords"`
	CommonErrors    []string  `json:"common_errors"`
	Prerequisites   []int     `json:"prerequisites"`
	Related         []int     `json:"related"`
	Variants        []int     `json:"variants"`
	Questions       []string  `json:"questions"`
	Embedding       []float64 `json:"embedding,omitempty"`
}

type KnowledgeDB struct {
	Topics []KnowledgeItem `json:"topics"`
}

type Submission struct {
	Code         string
	Language     string
	ProblemTitle string
	ProblemDesc  string
}

type JudgeResult struct {
	Status       string
	TimeUsed     int
	MemoryUsed   int
	ErrorMessage string
}

type UserRecord struct {
	QuestionID string `json:"question_id"`
	Status     string `json:"status"`
}

type MasteryLevel struct {
	Name       string
	TotalCount int
	ACCount    int
	Accuracy   float64
	Level      string
}

type UserProfile struct {
	UserID  string
	Records []UserRecord
	Mastery map[int]*MasteryLevel
}

type Recommendation struct {
	QuestionID string
	Reason     string
	Priority   int
	Source     string
}

type LearningAdvice struct {
	WeakAnalysis   string   `json:"weak_analysis"`
	RootCause      string   `json:"root_cause"`
	LearningPath   []string `json:"learning_path"`
	SpecificAdvice []string `json:"specific_advice"`
}

func LoadKnowledge(filename string) (*KnowledgeDB, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var db KnowledgeDB
	err = json.Unmarshal(data, &db)
	return &db, err
}

func LoadKnowledgeWithEmbedding(filename string) (*KnowledgeDB, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var db KnowledgeDB
	err = json.Unmarshal(data, &db)
	return &db, err
}

func getQuestionKPMapping(db *KnowledgeDB) map[string][]int {

	result := make(map[string][]int)
	for _, topic := range db.Topics {
		for _, qid := range topic.Questions {
			result[qid] = append(result[qid], topic.ID)
		}
	}
	return result
}

func getKPName(id int, db *KnowledgeDB) string {
	for _, t := range db.Topics {
		if t.ID == id {
			return t.Name
		}
	}
	return fmt.Sprintf("知识点%d", id)
}

func getKPByID(id int, db *KnowledgeDB) *KnowledgeItem {
	for i := range db.Topics {
		if db.Topics[i].ID == id {
			return &db.Topics[i]
		}
	}
	return nil
}

func getQuestionsByKP(kpID int, qkMap map[string][]int) []string {
	result := []string{}
	for qid, ids := range qkMap {
		for _, id := range ids {
			if id == kpID {
				result = append(result, qid)
				break
			}
		}
	}
	return result
}

func getNames(kpIDs []int, db *KnowledgeDB) string {
	names := []string{}
	for _, id := range kpIDs {
		names = append(names, getKPName(id, db))
	}
	return strings.Join(names, "、")
}

//掌握程度计算

func computeMastery(records []UserRecord, qkMap map[string][]int) map[int]*MasteryLevel {
	stats := make(map[int]*struct{ total, ac int })

	for _, r := range records {
		kpIDs, ok := qkMap[r.QuestionID]
		if !ok {
			continue
		}

		//判断是否 AC
		isAC := false
		status := strings.ToLower(r.Status)
		if status == "AC" ||
			status == "Accepted" ||
			status == "正确" ||
			status == "通过" ||
			status == "✅" {
			isAC = true
		}

		for _, kpID := range kpIDs {
			if stats[kpID] == nil {
				stats[kpID] = &struct{ total, ac int }{0, 0}
			}
			stats[kpID].total++
			if isAC {
				stats[kpID].ac++
			}
		}
	}

	result := make(map[int]*MasteryLevel)
	for kpID, s := range stats {
		acc := float64(s.ac) / float64(s.total) * 100
		level := "weak"
		if acc >= 70 {
			level = "mastered"
		} else if acc >= 40 {
			level = "consolidating"
		}
		result[kpID] = &MasteryLevel{
			Name:       fmt.Sprintf("知识点%d", kpID),
			TotalCount: s.total,
			ACCount:    s.ac,
			Accuracy:   acc,
			Level:      level,
		}
	}
	return result
}

func getWeakPoints(mastery map[int]*MasteryLevel) []int {
	weak := []int{}
	for id, m := range mastery {
		if m.Level == "weak" {
			weak = append(weak, id)
		}
	}
	return weak
}

func getConsolidatingPoints(mastery map[int]*MasteryLevel) []int {
	result := []int{}
	for id, m := range mastery {
		if m.Level == "consolidating" {
			result = append(result, id)
		}
	}
	return result
}

func getMasteredPoints(mastery map[int]*MasteryLevel) []int {
	result := []int{}
	for id, m := range mastery {
		if m.Level == "mastered" {
			result = append(result, id)
		}
	}
	return result
}
