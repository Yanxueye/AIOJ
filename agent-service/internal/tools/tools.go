package tools

import "encoding/json"

// ToolDef describes a tool available to the LLM (matches ai.ToolDefinition).
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"parameters"` // JSON Schema
}

// queryUserProblemsSchema is the JSON Schema for the query_user_problems tool.
var queryUserProblemsSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"tags": {
			"type": "array",
			"items": {"type": "string"},
			"description": "Filter by algorithm tags (e.g., [\"哈希表\", \"动态规划\"]). Leave empty for all."
		},
		"status": {
			"type": "string",
			"enum": ["solved", "attempted", "untried", ""],
			"description": "Filter by solve status: solved, attempted, untried, or empty for all"
		},
		"difficulty": {
			"type": "string",
			"enum": ["简单", "中等", "困难", ""],
			"description": "Filter by difficulty level, or empty for all"
		}
	},
	"required": []
}`)

// submitCodeSchema is the JSON Schema for the submit_code tool.
var submitCodeSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"problem_id": {
			"type": "integer",
			"description": "The problem ID to submit code for"
		},
		"code": {
			"type": "string",
			"description": "The complete source code to judge"
		},
		"language": {
			"type": "string",
			"description": "Programming language (e.g., cpp17, python3, java)",
			"enum": ["cpp17", "cpp20", "c11", "python3", "go122", "java21"]
		}
	},
	"required": ["problem_id", "code", "language"]
}`)

// retrieveKnowledgeSchema is the JSON Schema for the retrieve_knowledge tool.
var retrieveKnowledgeSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"tags": {
			"type": "array",
			"items": {"type": "string"},
			"description": "Algorithm tags to search knowledge for (e.g., [\"二分查找\", \"线段树\"])"
		},
		"query": {
			"type": "string",
			"description": "Optional free-text search query for more specific context"
		}
	},
	"required": ["tags"]
}`)

var getUserCodeSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"problem_id": {
			"type": "integer",
			"description": "The problem ID to query code for"
		}
	},
	"required": ["problem_id"]
}`)

// Definitions maps tool names to their definitions.
var Definitions = map[string]ToolDef{
	"search_problems": {
		Name:        "search_problems",
		Description: "根据用户提到的题目名称模糊搜索题库，返回匹配的题目ID、标题、标签和难度。当用户用中文名提到题但不知道题号时调用。示例：search_problems({query:\"最大子数组和\"}) → [{id:1007, title:\"最大子数组和\", tags:[\"数组\",\"动态规划\"], difficulty:\"中等\"}]",
		Schema:      searchProblemsSchema,
	},
	"query_user_problems": {
		Name:        "query_user_problems",
		Description: "查询当前用户的做题记录和知识点统计。当用户让你讲解知识点、分析薄弱点、推荐题目时，先调用此工具了解他的掌握情况。返回该知识点下所有做过的题（含状态和尝试次数）和各标签的AC率。示例：query_user_problems({tags:[\"动态规划\"]}) → 返回用户做过的DP题列表和DP标签的AC率。有了数据后再结合用户实际掌握情况做讲解，而不是泛泛而谈。",
		Schema:      queryUserProblemsSchema,
	},
	"retrieve_knowledge": {
		Name:        "retrieve_knowledge",
		Description: "从 OI-Wiki 检索算法知识文档。获取关于特定算法的定义、性质、例题等参考资料。示例：retrieve_knowledge({tags:[\"动态规划\"], query:\"状态转移方程\"}) → 返回 OI-Wiki 中动态规划相关的知识块。",
		Schema:      retrieveKnowledgeSchema,
	},
	"get_user_code": {
		Name:        "get_user_code",
		Description: "获取用户对指定题目最近一次提交的代码。当用户问能否看到他的代码、让你分析他的提交、或讨论他之前的解法时调用。返回代码内容、语言和评测状态。示例：get_user_code({problem_id:1007}) → {code:\"#include...\", language:\"cpp\", status:\"Accepted\"}。若用户未提交过该题则返回 found:false。",
		Schema:      getUserCodeSchema,
	},
	"submit_code": {
		Name:        "submit_code",
		Description: "提交代码进行在线评测，返回判题结果（Accepted/Wrong Answer 等）、执行时间、内存消耗。用这个验证用户或你自己生成的代码是否正确。示例：submit_code({problem_id:1007, code:\"#include...\", language:\"cpp17\"})。",
		Schema:      submitCodeSchema,
	},
}

// ToolNames returns all registered tool names.
func ToolNames() []string {
	names := make([]string, 0, len(Definitions))
	for n := range Definitions {
		names = append(names, n)
	}
	return names
}

var searchProblemsSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"query": {
			"type": "string",
			"description": "Fuzzy search keyword for problem title (e.g., '两数之和', 'maximum subarray')"
		}
	},
	"required": ["query"]
}`)
