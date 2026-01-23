package builder

type ProgressCallback interface {
	OnStageStart(stage string, totalSteps uint64)
	OnProgress(step uint64, totalSteps uint64)
	OnStageComplete(stage string)
	OnProgressError(err error)
}

type DefaultProgressCallback struct{}

func (d *DefaultProgressCallback) OnStageStart(stage string, totalSteps uint64) {}

func (d *DefaultProgressCallback) OnProgress(step, total uint64) {}

func (d *DefaultProgressCallback) OnStageComplete(stage string) {}

func (d *DefaultProgressCallback) OnProgressError(err error) {}
