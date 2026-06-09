package stats

import (
	"sync"

	"remote_judge/internal/domain"
)

// Snapshot 包含当前系统计数器。
type Snapshot struct {
	TotalSubmissions int
	ActiveWorkers    int
	MaxWorkers       int
	BusyRejects      int
	StatusCounts     map[string]int
}

// Collector 在内存中记录系统指标。
type Collector struct {
	mu           sync.RWMutex
	total        int
	active       int
	maxWorkers   int
	busyRejects  int
	statusCounts map[domain.SubmissionStatus]int
}

// NewCollector 创建新的统计采集器。
func NewCollector() *Collector {
	return &Collector{
		statusCounts: make(map[domain.SubmissionStatus]int),
	}
}

// SetMaxWorkers 记录已配置的 Worker 上限。
func (c *Collector) SetMaxWorkers(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxWorkers = n
}

// RecordSubmission 记录一次新的提交创建。
func (c *Collector) RecordSubmission() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.total++
}

// RecordStatus 记录一个终态判题状态。
func (c *Collector) RecordStatus(status domain.SubmissionStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statusCounts[status]++
}

// WorkerStarted 增加活跃 Worker 计数。
func (c *Collector) WorkerStarted() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.active++
}

// WorkerFinished 减少活跃 Worker 计数。
func (c *Collector) WorkerFinished() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active > 0 {
		c.active--
	}
}

// RecordBusyReject 记录因 Worker 池满而被拒绝的提交。
func (c *Collector) RecordBusyReject() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.busyRejects++
}

// Snapshot 返回当前指标的副本。
func (c *Collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]int, len(c.statusCounts))
	for k, v := range c.statusCounts {
		out[string(k)] = v
	}
	return Snapshot{
		TotalSubmissions: c.total,
		ActiveWorkers:    c.active,
		MaxWorkers:       c.maxWorkers,
		BusyRejects:      c.busyRejects,
		StatusCounts:     out,
	}
}
