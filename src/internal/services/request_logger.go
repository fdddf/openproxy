package services

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2/log"
)

const (
	// requestLogQueueSize bounds how many records may be waiting to be written.
	// Past this the proxy drops log records rather than slowing down requests.
	requestLogQueueSize = 512
	// requestLogBatchSize is the most records committed in one transaction.
	requestLogBatchSize = 64
	// requestLogFlushInterval bounds how long a record waits before being written.
	requestLogFlushInterval = time.Second
	// requestLogSweepInterval is how often expired records are pruned.
	requestLogSweepInterval = time.Hour
)

// RequestLogger persists proxied request records off the request path.
//
// Writing inline used to block every proxied call on a database round trip
// carrying the full request and response body — on SQLite, whose writes are
// serialised, that put a global lock in front of the proxy. Records are now
// queued and committed in batches by a single worker.
type RequestLogger interface {
	// Record queues a request record. It never blocks: if the queue is full the
	// record is dropped and counted.
	Record(entry *models.Request)
	// Start launches the writer and the retention sweeper.
	Start()
	// Stop flushes everything already queued and shuts the worker down.
	Stop(ctx context.Context) error
}

type requestLogger struct {
	dbService     DatabaseService
	configService ConfigService

	queue   chan *models.Request
	done    chan struct{}
	stopped chan struct{}
	wg      sync.WaitGroup

	startOnce sync.Once
	stopOnce  sync.Once

	dropped atomic.Int64
}

// NewRequestLogger creates the asynchronous request log writer.
func NewRequestLogger(dbService DatabaseService, configService ConfigService) RequestLogger {
	return &requestLogger{
		dbService:     dbService,
		configService: configService,
		queue:         make(chan *models.Request, requestLogQueueSize),
		done:          make(chan struct{}),
		stopped:       make(chan struct{}),
	}
}

func (r *requestLogger) config() common.RequestLogConfig {
	return r.configService.GetConfig().RequestLog
}

func (r *requestLogger) Record(entry *models.Request) {
	if entry == nil || !r.config().LoggingEnabled() {
		return
	}

	if max := r.config().MaxBodyBytes; max > 0 {
		entry.RequestBody = truncateBody(entry.RequestBody, max)
		entry.ResponseBody = truncateBody(entry.ResponseBody, max)
	}

	select {
	case r.queue <- entry:
	default:
		// Losing a log record is preferable to stalling a proxied request.
		if n := r.dropped.Add(1); n == 1 || n%100 == 0 {
			log.Warnf("request log queue full, dropped %d record(s)", n)
		}
	}
}

func (r *requestLogger) Start() {
	r.startOnce.Do(func() {
		r.wg.Add(2)
		go r.writeLoop()
		go r.sweepLoop()
	})
}

func (r *requestLogger) Stop(ctx context.Context) error {
	r.stopOnce.Do(func() { close(r.done) })

	finished := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *requestLogger) writeLoop() {
	defer r.wg.Done()

	ticker := time.NewTicker(requestLogFlushInterval)
	defer ticker.Stop()

	batch := make([]*models.Request, 0, requestLogBatchSize)

	for {
		select {
		case entry := <-r.queue:
			batch = append(batch, entry)
			if len(batch) >= requestLogBatchSize {
				r.flush(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				r.flush(batch)
				batch = batch[:0]
			}
		case <-r.done:
			// Drain whatever is still queued before exiting.
			for {
				select {
				case entry := <-r.queue:
					batch = append(batch, entry)
					if len(batch) >= requestLogBatchSize {
						r.flush(batch)
						batch = batch[:0]
					}
				default:
					if len(batch) > 0 {
						r.flush(batch)
					}
					return
				}
			}
		}
	}
}

func (r *requestLogger) flush(batch []*models.Request) {
	if len(batch) == 0 {
		return
	}

	dao := r.dbService.GetDAO()
	if dao == nil {
		return
	}

	if err := dao.Request.CreateInBatches(batch, requestLogBatchSize); err != nil {
		log.Errorf("record %d request log(s) failed: %v", len(batch), err)
	}
}

func (r *requestLogger) sweepLoop() {
	defer r.wg.Done()

	ticker := time.NewTicker(requestLogSweepInterval)
	defer ticker.Stop()

	r.sweep()
	for {
		select {
		case <-ticker.C:
			r.sweep()
		case <-r.done:
			return
		}
	}
}

// sweep deletes request records past the configured retention window.
func (r *requestLogger) sweep() {
	days := r.config().RetentionDays
	if days <= 0 {
		return
	}

	dao := r.dbService.GetDAO()
	if dao == nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	result, err := dao.Request.Unscoped().Where(dao.Request.CreatedAt.Lt(cutoff)).Delete()
	if err != nil {
		log.Errorf("prune request logs older than %d day(s): %v", days, err)
		return
	}
	if result.RowsAffected > 0 {
		log.Infof("pruned %d request log(s) older than %d day(s)", result.RowsAffected, days)
	}
}

// truncateBody caps a stored body, leaving a marker so the record is not
// mistaken for a genuinely short payload.
func truncateBody(body string, max int) string {
	if max <= 0 || len(body) <= max {
		return body
	}
	return body[:max] + "\n...[truncated]"
}
