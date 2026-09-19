package coldstore

import "context"

type Writer interface {
	AppendOperationalForecast(
		ctx context.Context,
		row OperationalForecastRow,
	) error

	AppendModelRun(
		ctx context.Context,
		row ModelRunRow,
	) error
}
