package backend

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// WARNING: this was produced by Opus 4 (web) as sample best-practices golang code
// for a worker pool, one per db partition. This is raw from Opus, and needs to be
// adapted and integrated into the rest of the code.  This is non-reachable at this
// point.

// Message represents a queue message
type Message struct {
	ID      string
	Type    string
	Payload []byte
}

// Result represents the processing result from a worker
type Result struct {
	MessageID string
	WorkerID  int
	Success   bool
	Error     error
	Duration  time.Duration
}

// Stats tracks processing statistics
type Stats struct {
	mu        sync.RWMutex
	processed int64
	succeeded int64
	failed    int64
	totalTime time.Duration
}

func (s *Stats) record(result Result) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processed++
	s.totalTime += result.Duration

	if result.Success {
		s.succeeded++
	} else {
		s.failed++
	}
}

func (s *Stats) snapshot() (processed, succeeded, failed int64, avgTime time.Duration) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	processed = s.processed
	succeeded = s.succeeded
	failed = s.failed

	if processed > 0 {
		avgTime = time.Duration(int64(s.totalTime) / processed)
	}

	return
}

// Worker represents a single worker
type Worker struct {
	id       int
	msgChan  chan Message
	resultCh chan<- Result
	quit     chan struct{}
	wg       *sync.WaitGroup
}

func NewWorker(id int, resultCh chan<- Result, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:       id,
		msgChan:  make(chan Message, 1), // Buffered to prevent blocking
		resultCh: resultCh,
		quit:     make(chan struct{}),
		wg:       wg,
	}
}

func (w *Worker) start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		log.Printf("Worker %d started", w.id)

		for {
			select {
			case <-ctx.Done():
				log.Printf("Worker %d shutting down due to context cancellation", w.id)
				return

			case <-w.quit:
				log.Printf("Worker %d shutting down", w.id)
				return

			case msg := <-w.msgChan:
				start := time.Now()
				err := w.processMessage(ctx, msg)

				result := Result{
					MessageID: msg.ID,
					WorkerID:  w.id,
					Success:   err == nil,
					Error:     err,
					Duration:  time.Since(start),
				}

				// Non-blocking send to prevent deadlock during shutdown
				select {
				case w.resultCh <- result:
				case <-ctx.Done():
					log.Printf("Worker %d: context cancelled while sending result", w.id)
					return
				}
			}
		}
	}()
}

func (w *Worker) processMessage(ctx context.Context, msg Message) error {
	// Simulate processing with context awareness
	log.Printf("Worker %d processing message %s", w.id, msg.ID)

	// Simulate some IO-bound work
	select {
	case <-time.After(time.Millisecond * 100): // Simulate work
		// Simulate occasional errors
		if msg.ID[0] == 'E' {
			return fmt.Errorf("simulated error for message %s", msg.ID)
		}
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *Worker) stop() {
	close(w.quit)
}

// WorkerPool manages all workers
type WorkerPool struct {
	workers  []*Worker
	resultCh chan Result
	stats    *Stats
	wg       sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	wp := &WorkerPool{
		workers:  make([]*Worker, numWorkers),
		resultCh: make(chan Result, numWorkers*2), // Buffered to prevent blocking
		stats:    &Stats{},
	}

	for i := 0; i < numWorkers; i++ {
		wp.workers[i] = NewWorker(i, wp.resultCh, &wp.wg)
	}

	return wp
}

func (wp *WorkerPool) start(ctx context.Context) {
	// Start all workers
	for _, w := range wp.workers {
		w.start(ctx)
	}

	// Start result collector
	wp.wg.Add(1)
	go wp.collectResults(ctx)

	// Start stats reporter
	wp.wg.Add(1)
	go wp.reportStats(ctx)
}

func (wp *WorkerPool) collectResults(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("Result collector shutting down")
			// Drain remaining results
			for {
				select {
				case result := <-wp.resultCh:
					wp.handleResult(result)
				default:
					return
				}
			}

		case result := <-wp.resultCh:
			wp.handleResult(result)
		}
	}
}

func (wp *WorkerPool) handleResult(result Result) {
	wp.stats.record(result)

	if result.Error != nil {
		log.Printf("Message %s failed on worker %d: %v",
			result.MessageID, result.WorkerID, result.Error)
	} else {
		log.Printf("Message %s succeeded on worker %d in %v",
			result.MessageID, result.WorkerID, result.Duration)
	}
}

func (wp *WorkerPool) reportStats(ctx context.Context) {
	defer wp.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Final stats report
			wp.printStats()
			return

		case <-ticker.C:
			wp.printStats()
		}
	}
}

func (wp *WorkerPool) printStats() {
	processed, succeeded, failed, avgTime := wp.stats.snapshot()
	log.Printf("Stats - Processed: %d, Succeeded: %d, Failed: %d, Avg Time: %v",
		processed, succeeded, failed, avgTime)
}

func (wp *WorkerPool) dispatch(msg Message) error {
	// Route message to specific worker based on payload
	workerIndex := wp.selectWorker(msg)
	worker := wp.workers[workerIndex]

	// Non-blocking send to prevent deadlock
	select {
	case worker.msgChan <- msg:
		return nil
	default:
		return fmt.Errorf("worker %d queue is full", workerIndex)
	}
}

func (wp *WorkerPool) selectWorker(msg Message) int {
	// Example: Use message type to determine worker
	// This ensures same type always goes to same worker
	hash := 0
	for _, b := range []byte(msg.Type) {
		hash = (hash * 31) + int(b)
	}
	return abs(hash) % len(wp.workers)
}

func (wp *WorkerPool) shutdown() {
	log.Println("Shutting down worker pool...")

	// Stop all workers
	for _, w := range wp.workers {
		w.stop()
	}

	// Wait for all goroutines to finish
	wp.wg.Wait()

	// Close result channel after all workers are done
	close(wp.resultCh)

	log.Println("Worker pool shutdown complete")
}

// Main application structure
type App struct {
	pool     *WorkerPool
	shutdown chan struct{}
	done     chan struct{}
}

func NewApp() *App {
	return &App{
		pool:     NewWorkerPool(10),
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (app *App) run(ctx context.Context) {
	defer close(app.done)

	// Start worker pool
	app.pool.start(ctx)

	// Simulate message processing
	msgCount := int64(0)
	ticker := time.NewTicker(10 * time.Millisecond) // Simulate message arrival rate
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Main processor shutting down")
			app.pool.shutdown()
			return

		case <-app.shutdown:
			log.Println("Shutdown signal received")
			app.pool.shutdown()
			return

		case <-ticker.C:
			// Simulate receiving a message from queue
			msgID := fmt.Sprintf("MSG-%d", atomic.AddInt64(&msgCount, 1))

			// Add some variety to message types
			msgType := "typeA"
			if msgCount%3 == 0 {
				msgType = "typeB"
			} else if msgCount%5 == 0 {
				msgType = "typeC"
			}

			// Simulate error messages
			if msgCount%13 == 0 {
				msgID = "E" + msgID // Will trigger error in worker
			}

			msg := Message{
				ID:      msgID,
				Type:    msgType,
				Payload: []byte(fmt.Sprintf("Payload for %s", msgID)),
			}

			if err := app.pool.dispatch(msg); err != nil {
				log.Printf("Failed to dispatch message %s: %v", msg.ID, err)
			}
		}
	}
}

func main() {
	// Setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Create root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and start application
	app := NewApp()

	// Start main processing in background
	go app.run(ctx)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal: %v", sig)

	// Cancel context to trigger graceful shutdown
	cancel()

	// Wait for app to finish with timeout
	shutdownComplete := make(chan struct{})
	go func() {
		<-app.done
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		log.Println("Graceful shutdown completed")
	case <-time.After(30 * time.Second):
		log.Println("Shutdown timeout exceeded")
		os.Exit(1)
	}
}

// Helper functions
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
