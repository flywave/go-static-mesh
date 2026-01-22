package mesh

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type ErrorRecoveryStrategy int

const (
	RecoverySkip ErrorRecoveryStrategy = iota
	RecoveryRetry
	RecoveryUseDefault
	RecoveryAbort
)

func (s ErrorRecoveryStrategy) String() string {
	switch s {
	case RecoverySkip:
		return "SKIP"
	case RecoveryRetry:
		return "RETRY"
	case RecoveryUseDefault:
		return "USE_DEFAULT"
	case RecoveryAbort:
		return "ABORT"
	default:
		return "UNKNOWN"
	}
}

type ErrorPolicy struct {
	Strategies map[ErrorType]ErrorRecoveryStrategy
	MaxRetries int
	RetryDelay time.Duration
	OnError    func(error)
	OnRecover  func(ErrorType, ErrorRecoveryStrategy)
}

func DefaultErrorPolicy() *ErrorPolicy {
	return &ErrorPolicy{
		Strategies: map[ErrorType]ErrorRecoveryStrategy{
			ErrTypeNetwork:    RecoveryRetry,
			ErrTypeTile:       RecoverySkip,
			ErrTypeProvider:   RecoveryAbort,
			ErrTypeTIN:        RecoveryAbort,
			ErrTypeBounds:     RecoveryAbort,
			ErrTypeCoordinate: RecoverySkip,
			ErrTypeMemory:     RecoveryAbort,
			ErrTypeFile:       RecoveryAbort,
			ErrTypeDecode:     RecoverySkip,
		},
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		OnError: func(err error) {
			log.Printf("Mesh error: %v", err)
		},
		OnRecover: func(errType ErrorType, strategy ErrorRecoveryStrategy) {
			log.Printf("Error recovery: [%s] -> %s", errType, strategy)
		},
	}
}

func LenientErrorPolicy() *ErrorPolicy {
	policy := DefaultErrorPolicy()
	policy.Strategies[ErrTypeNetwork] = RecoverySkip
	policy.Strategies[ErrTypeTile] = RecoverySkip
	policy.Strategies[ErrTypeDecode] = RecoverySkip
	return policy
}

func StrictErrorPolicy() *ErrorPolicy {
	policy := DefaultErrorPolicy()
	policy.Strategies[ErrTypeTile] = RecoveryAbort
	policy.Strategies[ErrTypeDecode] = RecoveryAbort
	policy.Strategies[ErrTypeCoordinate] = RecoveryAbort
	return policy
}

type ErrorHandler struct {
	policy     *ErrorPolicy
	errorCount map[ErrorType]int
	lastErrors map[ErrorType]time.Time
	errorMutex sync.RWMutex
}

func NewErrorHandler(policy *ErrorPolicy) *ErrorHandler {
	if policy == nil {
		policy = DefaultErrorPolicy()
	}
	return &ErrorHandler{
		policy:     policy,
		errorCount: make(map[ErrorType]int),
		lastErrors: make(map[ErrorType]time.Time),
	}
}

func (h *ErrorHandler) HandleError(err error, context map[string]interface{}) ErrorRecoveryStrategy {
	h.recordError(err)

	errType := h.classifyError(err)
	strategy := h.policy.Strategies[errType]

	if h.policy.OnError != nil {
		fullError := err
		if len(context) > 0 {
			fullError = fmt.Errorf("%w (context: %v)", err, context)
		}
		h.policy.OnError(fullError)
	}

	if h.shouldAbort(errType) {
		return RecoveryAbort
	}

	if h.policy.OnRecover != nil {
		h.policy.OnRecover(errType, strategy)
	}

	return strategy
}

func (h *ErrorHandler) recordError(err error) {
	errType := h.classifyError(err)
	now := time.Now()

	h.errorMutex.Lock()
	defer h.errorMutex.Unlock()

	h.errorCount[errType]++
	h.lastErrors[errType] = now
}

func (h *ErrorHandler) classifyError(err error) ErrorType {
	if meshErr, ok := err.(*MeshError); ok {
		return meshErr.Type
	}

	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection refused"):
		return ErrTypeNetwork
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "404"):
		return ErrTypeTile
	case strings.Contains(errStr, "bounds") || strings.Contains(errStr, "invalid range"):
		return ErrTypeBounds
	case strings.Contains(errStr, "coordinate") || strings.Contains(errStr, "projection"):
		return ErrTypeCoordinate
	case strings.Contains(errStr, "TIN") || strings.Contains(errStr, "triangulation"):
		return ErrTypeTIN
	case strings.Contains(errStr, "memory") || strings.Contains(errStr, "allocation"):
		return ErrTypeMemory
	case strings.Contains(errStr, "file") || strings.Contains(errStr, "no such file"):
		return ErrTypeFile
	case strings.Contains(errStr, "decode") || strings.Contains(errStr, "parse"):
		return ErrTypeDecode
	default:
		return ErrTypeUnknown
	}
}

func (h *ErrorHandler) shouldAbort(errType ErrorType) bool {
	h.errorMutex.RLock()
	defer h.errorMutex.RUnlock()

	count := h.errorCount[errType]

	if count >= 100 {
		return true
	}

	switch h.policy.Strategies[errType] {
	case RecoveryAbort:
		return count >= 5
	case RecoveryRetry:
		return count >= h.policy.MaxRetries*2
	default:
		return false
	}
}

func (h *ErrorHandler) GetErrorCount(errType ErrorType) int {
	h.errorMutex.RLock()
	defer h.errorMutex.RUnlock()
	return h.errorCount[errType]
}

func (h *ErrorHandler) GetTotalErrorCount() int {
	h.errorMutex.RLock()
	defer h.errorMutex.RUnlock()

	total := 0
	for _, count := range h.errorCount {
		total += count
	}
	return total
}

func (h *ErrorHandler) GetLastErrorTime(errType ErrorType) time.Time {
	h.errorMutex.RLock()
	defer h.errorMutex.RUnlock()
	return h.lastErrors[errType]
}

func (h *ErrorHandler) Reset() {
	h.errorMutex.Lock()
	defer h.errorMutex.Unlock()

	h.errorCount = make(map[ErrorType]int)
	h.lastErrors = make(map[ErrorType]time.Time)
}

func (h *ErrorHandler) GetStats() map[string]interface{} {
	h.errorMutex.RLock()
	defer h.errorMutex.RUnlock()

	stats := make(map[string]interface{})
	total := 0

	for errType, count := range h.errorCount {
		stats[errType.String()] = map[string]interface{}{
			"count": count,
			"last":  h.lastErrors[errType],
		}
		total += count
	}

	stats["total"] = total
	return stats
}
