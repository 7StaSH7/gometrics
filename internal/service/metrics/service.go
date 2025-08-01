package metrics

type MetricsService interface {
	UpdateMetric(mType, name string, value any) error
}

type metricsService struct {
}

func New() MetricsService {
	return &metricsService{}
}
