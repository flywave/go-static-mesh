package mesh

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorType int

const (
	ErrTypeNetwork ErrorType = iota
	ErrTypeTile
	ErrTypeProvider
	ErrTypeTIN
	ErrTypeBounds
	ErrTypeCoordinate
	ErrTypeMemory
	ErrTypeFile
	ErrTypeDecode
	ErrTypeUnknown
)

func (e ErrorType) String() string {
	switch e {
	case ErrTypeNetwork:
		return "NETWORK"
	case ErrTypeTile:
		return "TILE"
	case ErrTypeProvider:
		return "PROVIDER"
	case ErrTypeTIN:
		return "TIN"
	case ErrTypeBounds:
		return "BOUNDS"
	case ErrTypeCoordinate:
		return "COORDINATE"
	case ErrTypeMemory:
		return "MEMORY"
	case ErrTypeFile:
		return "FILE"
	case ErrTypeDecode:
		return "DECODE"
	default:
		return "UNKNOWN"
	}
}

type MeshError struct {
	Type       ErrorType
	Operation  string
	TileCoord  [3]int
	Underlying error
	Context    map[string]interface{}
}

func (e *MeshError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Type, e.Operation)
	if e.TileCoord != [3]int{} {
		msg += fmt.Sprintf(" (tile %d/%d/%d)", e.TileCoord[2], e.TileCoord[0], e.TileCoord[1])
	}
	if len(e.Context) > 0 {
		var ctx []string
		for k, v := range e.Context {
			ctx = append(ctx, fmt.Sprintf("%s=%v", k, v))
		}
		if len(ctx) > 0 {
			msg += fmt.Sprintf(" [%s]", strings.Join(ctx, ", "))
		}
	}
	if e.Underlying != nil {
		msg += fmt.Sprintf(": %v", e.Underlying)
	}
	return msg
}

func (e *MeshError) Unwrap() error {
	return e.Underlying
}

func NewNetworkError(op string, tile [3]int, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeNetwork,
		Operation:  op,
		TileCoord:  tile,
		Underlying: err,
	}
}

func NewTileError(op string, tile [3]int, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeTile,
		Operation:  op,
		TileCoord:  tile,
		Underlying: err,
	}
}

func NewProviderError(op string, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeProvider,
		Operation:  op,
		Underlying: err,
	}
}

func NewTINError(op string, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeTIN,
		Operation:  op,
		Underlying: err,
	}
}

func NewBoundsError(op string, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeBounds,
		Operation:  op,
		Underlying: err,
	}
}

func NewCoordinateError(op string, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeCoordinate,
		Operation:  op,
		Underlying: err,
	}
}

func NewMemoryError(op string, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeMemory,
		Operation:  op,
		Underlying: err,
	}
}

func NewFileError(op string, filepath string, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeFile,
		Operation:  op,
		Underlying: err,
		Context: map[string]interface{}{
			"filepath": filepath,
		},
	}
}

func NewDecodeError(op string, format string, err error) *MeshError {
	return &MeshError{
		Type:       ErrTypeDecode,
		Operation:  op,
		Underlying: err,
		Context: map[string]interface{}{
			"format": format,
		},
	}
}

var (
	ErrNoProviderSet           = errors.New("no raster or tin mesh provider set")
	ErrBoundsNotSet            = errors.New("bounds not set properly")
	ErrNoTINGenerated          = errors.New("failed to generate TIN mesh")
	ErrFailedToCloseMesh       = errors.New("failed to close mesh")
	ErrFailedToAddGeoData      = errors.New("failed to add geo data to mesh")
	ErrFailedToGenerateTexture = errors.New("failed to generate texture")
	ErrInvalidProviderType     = errors.New("invalid provider type")
	ErrProviderNotSupported    = errors.New("provider does not support required method")
)

type BuildError struct {
	Stage   string
	Step    string
	Err     error
	Context map[string]interface{}
}

func (e *BuildError) Error() string {
	if e.Err == nil {
		return e.Stage
	}
	if e.Step != "" {
		return fmt.Sprintf("%s: %s: %v", e.Stage, e.Step, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

func (e *BuildError) Unwrap() error {
	return e.Err
}
