// async_intelligence.go - Asynchronous intelligence processing architecture
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// AsyncIntelligenceProcessor handles non-blocking intelligence operations
type AsyncIntelligenceProcessor struct {
	workQueue      chan IntelligenceWork
	resultStream   chan IntelligenceResult
	workers        []IntelligenceWorker
	workerCount    int
	program        *tea.Program
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	priorityQueue  *PriorityWorkQueue
	throttle       *IntelligenceThrottle
}

// IntelligenceWork represents a unit of intelligence processing
type IntelligenceWork struct {
	ID          string
	Type        IntelligenceWorkType
	Priority    int // 1-10 (10 = highest)
	Input       string
	Context     *KubeContextSummary
	Callback    func(IntelligenceResult)
	Timestamp   time.Time
	Timeout     time.Duration
	MaxRetries  int
}

// IntelligenceWorkType defines the type of intelligence work
type IntelligenceWorkType string

const (
	WorkTypeAnalysis     IntelligenceWorkType = "analysis"
	WorkTypePrediction   IntelligenceWorkType = "prediction"
	WorkTypeValidation   IntelligenceWorkType = "validation"
	WorkTypeOptimization IntelligenceWorkType = "optimization"
	WorkTypeDiagnostic   IntelligenceWorkType = "diagnostic"
)

// IntelligenceResult represents the result of intelligence processing
type IntelligenceResult struct {
	ID            string
	Type          IntelligenceWorkType
	Success       bool
	Data          interface{}
	Error         error
	ProcessingTime time.Duration
	Cached        bool
	Confidence    float64
	Metadata      map[string]interface{}
}

// IntelligenceWorker processes intelligence work items
type IntelligenceWorker struct {
	id           int
	workChan     chan IntelligenceWork
	engine       *IntelligenceEngine
	resultStream chan IntelligenceResult
	ctx          context.Context
}

// PriorityWorkQueue manages work items by priority
type PriorityWorkQueue struct {
	items []IntelligenceWork
	mutex sync.RWMutex
}

// IntelligenceThrottle prevents overwhelming the LLM with requests
type IntelligenceThrottle struct {
	requests      map[string]time.Time
	maxRPS        int
	windowSize    time.Duration
	mutex         sync.RWMutex
}

// NewAsyncIntelligenceProcessor creates a new async intelligence processor
func NewAsyncIntelligenceProcessor(workerCount int, program *tea.Program) *AsyncIntelligenceProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	processor := &AsyncIntelligenceProcessor{
		workQueue:     make(chan IntelligenceWork, 100),
		resultStream:  make(chan IntelligenceResult, 100),
		workers:       make([]IntelligenceWorker, workerCount),
		workerCount:   workerCount,
		program:       program,
		ctx:           ctx,
		cancel:        cancel,
		priorityQueue: NewPriorityWorkQueue(),
		throttle:      NewIntelligenceThrottle(5, time.Second), // 5 requests per second
	}

	return processor
}

// Start begins the async intelligence processing
func (aip *AsyncIntelligenceProcessor) Start() {
	// Start priority queue processor
	go aip.processPriorityQueue()

	// Start workers
	for i := 0; i < aip.workerCount; i++ {
		worker := IntelligenceWorker{
			id:           i,
			workChan:     aip.workQueue,
			engine:       Intelligence,
			resultStream: aip.resultStream,
			ctx:          aip.ctx,
		}
		aip.workers[i] = worker
		aip.wg.Add(1)
		go worker.start(&aip.wg)
	}

	// Start result processor
	go aip.processResults()
}

// Stop gracefully shuts down the async processor
func (aip *AsyncIntelligenceProcessor) Stop() {
	aip.cancel()
	close(aip.workQueue)
	aip.wg.Wait()
	close(aip.resultStream)
}

// SubmitWork submits intelligence work for async processing
func (aip *AsyncIntelligenceProcessor) SubmitWork(work IntelligenceWork) {
	work.Timestamp = time.Now()
	aip.priorityQueue.Add(work)
}

// SubmitHighPriorityWork submits high-priority work that bypasses the queue
func (aip *AsyncIntelligenceProcessor) SubmitHighPriorityWork(work IntelligenceWork) {
	work.Priority = 10
	work.Timestamp = time.Now()

	select {
	case aip.workQueue <- work:
	case <-aip.ctx.Done():
	}
}

// processPriorityQueue processes work items from the priority queue
func (aip *AsyncIntelligenceProcessor) processPriorityQueue() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-aip.ctx.Done():
			return
		case <-ticker.C:
			work := aip.priorityQueue.Pop()
			if work != nil {
				// Apply throttling
				if aip.throttle.ShouldThrottle(work.Type) {
					// Re-queue with delay
					go func() {
						time.Sleep(200 * time.Millisecond)
						aip.priorityQueue.Add(*work)
					}()
					continue
				}

				select {
				case aip.workQueue <- *work:
				case <-aip.ctx.Done():
					return
				default:
					// Queue full, re-add to priority queue
					aip.priorityQueue.Add(*work)
				}
			}
		}
	}
}

// processResults processes intelligence results and sends them to the UI
func (aip *AsyncIntelligenceProcessor) processResults() {
	for {
		select {
		case <-aip.ctx.Done():
			return
		case result := <-aip.resultStream:
			// Send result to UI
			if aip.program != nil {
				aip.program.Send(asyncIntelligenceResultMsg{result: result})
			}
		}
	}
}

// start begins worker processing
func (worker *IntelligenceWorker) start(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-worker.ctx.Done():
			return
		case work := <-worker.workChan:
			result := worker.processWork(work)

			// Send result back
			select {
			case worker.resultStream <- result:
			case <-worker.ctx.Done():
				return
			}
		}
	}
}

// processWork processes a single intelligence work item
func (worker *IntelligenceWorker) processWork(work IntelligenceWork) IntelligenceResult {
	startTime := time.Now()

	result := IntelligenceResult{
		ID:            work.ID,
		Type:          work.Type,
		Success:       false,
		ProcessingTime: 0,
		Metadata:      make(map[string]interface{}),
	}

	defer func() {
		result.ProcessingTime = time.Since(startTime)
	}()

	switch work.Type {
	case WorkTypeAnalysis:
		session, err := worker.engine.AnalyzeIntelligently(work.Input, work.Context)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
			result.Data = session
			result.Confidence = session.Confidence
		}

	case WorkTypePrediction:
		// Future: Implement predictive analysis
		result.Success = true
		result.Data = "prediction_placeholder"

	case WorkTypeValidation:
		// Future: Implement async validation
		result.Success = true
		result.Data = "validation_placeholder"

	case WorkTypeOptimization:
		// Future: Implement async optimization
		result.Success = true
		result.Data = "optimization_placeholder"

	case WorkTypeDiagnostic:
		// Future: Implement async diagnostics
		result.Success = true
		result.Data = "diagnostic_placeholder"

	default:
		result.Error = fmt.Errorf("unknown work type: %s", work.Type)
	}

	return result
}

// NewPriorityWorkQueue creates a new priority work queue
func NewPriorityWorkQueue() *PriorityWorkQueue {
	return &PriorityWorkQueue{
		items: make([]IntelligenceWork, 0),
	}
}

// Add adds a work item to the priority queue
func (pq *PriorityWorkQueue) Add(work IntelligenceWork) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// Insert in priority order (highest priority first)
	inserted := false
	for i, item := range pq.items {
		if work.Priority > item.Priority {
			pq.items = append(pq.items[:i], append([]IntelligenceWork{work}, pq.items[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		pq.items = append(pq.items, work)
	}
}

// Pop removes and returns the highest priority work item
func (pq *PriorityWorkQueue) Pop() *IntelligenceWork {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	work := pq.items[0]
	pq.items = pq.items[1:]
	return &work
}

// NewIntelligenceThrottle creates a new intelligence throttle
func NewIntelligenceThrottle(maxRPS int, windowSize time.Duration) *IntelligenceThrottle {
	return &IntelligenceThrottle{
		requests:   make(map[string]time.Time),
		maxRPS:     maxRPS,
		windowSize: windowSize,
	}
}

// ShouldThrottle checks if the request should be throttled
func (it *IntelligenceThrottle) ShouldThrottle(workType IntelligenceWorkType) bool {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	now := time.Now()
	key := string(workType)

	if lastRequest, exists := it.requests[key]; exists {
		if now.Sub(lastRequest) < it.windowSize {
			return true
		}
	}

	it.requests[key] = now
	return false
}

// asyncIntelligenceResultMsg is sent when async intelligence processing completes
type asyncIntelligenceResultMsg struct {
	result IntelligenceResult
}