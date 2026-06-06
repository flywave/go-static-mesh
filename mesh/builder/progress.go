package builder

type ProgressCallback interface {
	OnStageStart(stage string, totalSteps uint64)
	OnProgress(step uint64, totalSteps uint64)
	OnStageComplete(stage string)
	OnProgressError(err error)
}


