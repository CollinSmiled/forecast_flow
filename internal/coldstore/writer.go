package coldstore

import "context"

type Writer interface {
	AppendOperationalForecasts(
		ctx context.Context,
		rows []OperationalForecastRow,
	) error

	AppendModelRuns(
		ctx context.Context,
		rows []ModelRunRow,
	) error
}
