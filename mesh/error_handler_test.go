package mesh

import (
	"errors"
	"testing"
	"time"
)

func TestMeshError(t *testing.T) {
	tests := []struct {
		name string
		err  *MeshError
		want string
	}{
		{
			name: "network error with tile coord",
			err:  NewNetworkError("FetchTile", [3]int{10, 20, 15}, errors.New("connection timeout")),
			want: "[NETWORK] FetchTile (tile 15/10/20): connection timeout",
		},
		{
			name: "tile error",
			err:  NewTileError("GetTileData", [3]int{5, 10, 12}, errors.New("not found")),
			want: "[TILE] GetTileData (tile 12/5/10): not found",
		},
		{
			name: "provider error",
			err:  NewProviderError("GetElevationGrid", errors.New("provider not initialized")),
			want: "[PROVIDER] GetElevationGrid: provider not initialized",
		},
		{
			name: "TIN error",
			err:  NewTINError("GenerateFromRaster", errors.New("insufficient points")),
			want: "[TIN] GenerateFromRaster: insufficient points",
		},
		{
			name: "bounds error",
			err:  NewBoundsError("SetBounds", errors.New("min >= max")),
			want: "[BOUNDS] SetBounds: min >= max",
		},
		{
			name: "file error with context",
			err:  NewFileError("ReadFile", "/path/to/file.tif", errors.New("no such file")),
			want: "[FILE] ReadFile [filepath=/path/to/file.tif]: no such file",
		},
		{
			name: "decode error with context",
			err:  NewDecodeError("Decode", "quantized-mesh", errors.New("invalid header")),
			want: "[DECODE] Decode [format=quantized-mesh]: invalid header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("MeshError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errType ErrorType
		want    string
	}{
		{ErrTypeNetwork, "NETWORK"},
		{ErrTypeTile, "TILE"},
		{ErrTypeProvider, "PROVIDER"},
		{ErrTypeTIN, "TIN"},
		{ErrTypeBounds, "BOUNDS"},
		{ErrTypeCoordinate, "COORDINATE"},
		{ErrTypeMemory, "MEMORY"},
		{ErrTypeFile, "FILE"},
		{ErrTypeDecode, "DECODE"},
		{ErrTypeUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.errType.String(); got != tt.want {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultErrorPolicy(t *testing.T) {
	policy := DefaultErrorPolicy()

	if policy == nil {
		t.Fatal("DefaultErrorPolicy() returned nil")
	}

	tests := []struct {
		errType ErrorType
		want    ErrorRecoveryStrategy
		retries int
		delay   time.Duration
	}{
		{ErrTypeNetwork, RecoveryRetry, 3, 1 * time.Second},
		{ErrTypeTile, RecoverySkip, 3, 1 * time.Second},
		{ErrTypeProvider, RecoveryAbort, 3, 1 * time.Second},
		{ErrTypeTIN, RecoveryAbort, 3, 1 * time.Second},
		{ErrTypeBounds, RecoveryAbort, 3, 1 * time.Second},
		{ErrTypeCoordinate, RecoverySkip, 3, 1 * time.Second},
		{ErrTypeMemory, RecoveryAbort, 3, 1 * time.Second},
		{ErrTypeFile, RecoveryAbort, 3, 1 * time.Second},
		{ErrTypeDecode, RecoverySkip, 3, 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.want.String(), func(t *testing.T) {
			if got := policy.Strategies[tt.errType]; got != tt.want {
				t.Errorf("Error policy for %v = %v, want %v", tt.errType, got, tt.want)
			}
			if policy.MaxRetries != tt.retries {
				t.Errorf("MaxRetries = %v, want %v", policy.MaxRetries, tt.retries)
			}
			if policy.RetryDelay != tt.delay {
				t.Errorf("RetryDelay = %v, want %v", policy.RetryDelay, tt.delay)
			}
		})
	}
}

func TestLenientErrorPolicy(t *testing.T) {
	policy := LenientErrorPolicy()

	if policy.Strategies[ErrTypeNetwork] != RecoverySkip {
		t.Error("Lenient policy should skip network errors")
	}
	if policy.Strategies[ErrTypeTile] != RecoverySkip {
		t.Error("Lenient policy should skip tile errors")
	}
}

func TestStrictErrorPolicy(t *testing.T) {
	policy := StrictErrorPolicy()

	if policy.Strategies[ErrTypeTile] != RecoveryAbort {
		t.Error("Strict policy should abort on tile errors")
	}
	if policy.Strategies[ErrTypeDecode] != RecoveryAbort {
		t.Error("Strict policy should abort on decode errors")
	}
}

func TestErrorHandler(t *testing.T) {
	handler := NewErrorHandler(nil)

	if handler == nil {
		t.Fatal("NewErrorHandler() returned nil")
	}

	t.Run("HandleNetworkError", func(t *testing.T) {
		err := NewNetworkError("FetchTile", [3]int{10, 20, 15}, errors.New("timeout"))
		strategy := handler.HandleError(err, nil)

		if strategy != RecoveryRetry {
			t.Errorf("Expected Retry strategy for network error, got %v", strategy)
		}

		count := handler.GetErrorCount(ErrTypeNetwork)
		if count != 1 {
			t.Errorf("Expected error count 1, got %d", count)
		}
	})

	t.Run("HandleTileError", func(t *testing.T) {
		err := NewTileError("GetTileData", [3]int{5, 10, 12}, errors.New("not found"))
		strategy := handler.HandleError(err, nil)

		if strategy != RecoverySkip {
			t.Errorf("Expected Skip strategy for tile error, got %v", strategy)
		}
	})

	t.Run("HandleProviderError", func(t *testing.T) {
		err := NewProviderError("GetElevationGrid", errors.New("not initialized"))
		strategy := handler.HandleError(err, nil)

		if strategy != RecoveryAbort {
			t.Errorf("Expected Abort strategy for provider error, got %v", strategy)
		}
	})

	t.Run("HandleMultipleErrors", func(t *testing.T) {
		handler.Reset()

		for i := 0; i < 5; i++ {
			err := NewNetworkError("FetchTile", [3]int{10, 20, 15}, errors.New("timeout"))
			handler.HandleError(err, nil)
		}

		if handler.GetErrorCount(ErrTypeNetwork) != 5 {
			t.Errorf("Expected error count 5, got %d", handler.GetErrorCount(ErrTypeNetwork))
		}

		if handler.GetTotalErrorCount() != 5 {
			t.Errorf("Expected total error count 5, got %d", handler.GetTotalErrorCount())
		}
	})

	t.Run("GetLastErrorTime", func(t *testing.T) {
		handler.Reset()

		err := NewNetworkError("FetchTile", [3]int{10, 20, 15}, errors.New("timeout"))
		handler.HandleError(err, nil)

		lastTime := handler.GetLastErrorTime(ErrTypeNetwork)
		if lastTime.IsZero() {
			t.Error("Expected non-zero last error time")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		handler.Reset()

		handler.HandleError(NewNetworkError("Fetch", [3]int{}, errors.New("err")), nil)
		handler.HandleError(NewTileError("Get", [3]int{}, errors.New("err")), nil)

		stats := handler.GetStats()
		if stats == nil {
			t.Fatal("GetStats() returned nil")
		}

		if total, ok := stats["total"].(int); !ok || total != 2 {
			t.Errorf("Expected total stats = 2, got %v", stats["total"])
		}
	})

	t.Run("CustomPolicy", func(t *testing.T) {
		customPolicy := &ErrorPolicy{
			Strategies: map[ErrorType]ErrorRecoveryStrategy{
				ErrTypeNetwork: RecoverySkip,
			},
			MaxRetries: 5,
			RetryDelay: 2 * time.Second,
		}

		customHandler := NewErrorHandler(customPolicy)
		err := NewNetworkError("Fetch", [3]int{}, errors.New("timeout"))
		strategy := customHandler.HandleError(err, nil)

		if strategy != RecoverySkip {
			t.Errorf("Expected Skip strategy, got %v", strategy)
		}
	})

	t.Run("CustomOnError", func(t *testing.T) {
		var capturedError error
		customPolicy := &ErrorPolicy{
			Strategies: map[ErrorType]ErrorRecoveryStrategy{
				ErrTypeNetwork: RecoveryRetry,
			},
			MaxRetries: 3,
			RetryDelay: 1 * time.Second,
			OnError: func(err error) {
				capturedError = err
			},
		}

		handler := NewErrorHandler(customPolicy)
		err := NewNetworkError("Fetch", [3]int{}, errors.New("timeout"))
		handler.HandleError(err, nil)

		if capturedError == nil {
			t.Error("Expected OnError to be called")
		}
	})
}

func TestErrorRecoveryStrategyString(t *testing.T) {
	tests := []struct {
		strategy ErrorRecoveryStrategy
		want     string
	}{
		{RecoverySkip, "SKIP"},
		{RecoveryRetry, "RETRY"},
		{RecoveryUseDefault, "USE_DEFAULT"},
		{RecoveryAbort, "ABORT"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.strategy.String(); got != tt.want {
				t.Errorf("ErrorRecoveryStrategy.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorClassification(t *testing.T) {
	handler := NewErrorHandler(nil)

	tests := []struct {
		name     string
		err      error
		wantType ErrorType
	}{
		{"timeout error", errors.New("connection timeout"), ErrTypeNetwork},
		{"404 error", errors.New("tile not found (404)"), ErrTypeTile},
		{"bounds error", errors.New("invalid bounds: min >= max"), ErrTypeBounds},
		{"coordinate error", errors.New("projection failed"), ErrTypeCoordinate},
		{"TIN error", errors.New("triangulation failed"), ErrTypeTIN},
		{"memory error", errors.New("memory allocation failed"), ErrTypeMemory},
		{"file error", errors.New("no such file or directory"), ErrTypeFile},
		{"decode error", errors.New("failed to decode"), ErrTypeDecode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType := handler.classifyError(tt.err)
			if gotType != tt.wantType {
				t.Errorf("classifyError() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	meshErr := NewNetworkError("operation", [3]int{}, underlying)

	if meshErr.Unwrap() != underlying {
		t.Error("MeshError.Unwrap() should return the underlying error")
	}

	if !errors.Is(meshErr, underlying) {
		t.Error("errors.Is should return true for underlying error")
	}
}

func TestErrorWithContext(t *testing.T) {
	err := NewFileError("ReadFile", "/path/to/file", errors.New("not found"))

	if _, ok := err.Context["filepath"]; !ok {
		t.Error("Expected filepath in context")
	}
	if err.Context["filepath"] != "/path/to/file" {
		t.Errorf("Expected filepath = /path/to/file, got %v", err.Context["filepath"])
	}
}
