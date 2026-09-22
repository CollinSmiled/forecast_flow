package verification

import (
	"testing"
	"time"
)

func TestSnapshotValidate(t *testing.T) {
	snapshot := validSnapshot()

	if err := snapshot.Validate(); err != nil {
		t.Fatalf("validate snapshot: %v", err)
	}
}

func TestSnapshotValidateRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Snapshot)
	}{
		{
			name: "missing location",
			mutate: func(snapshot *Snapshot) {
				snapshot.LocationID = 0
			},
		},
		{
			name: "unsupported reference kind",
			mutate: func(snapshot *Snapshot) {
				snapshot.ReferenceKind = "sensor"
			},
		},
		{
			name: "unordered hours",
			mutate: func(snapshot *Snapshot) {
				snapshot.Hourly = append(snapshot.Hourly, snapshot.Hourly[0])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := validSnapshot()
			test.mutate(&snapshot)

			if err := snapshot.Validate(); err == nil {
				t.Fatal("validation error = nil, want an error")
			}
		})
	}
}

func validSnapshot() Snapshot {
	validAt := time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC)

	return Snapshot{
		LocationID:    3,
		Source:        "open_meteo_archive_best_match",
		ReferenceKind: ReferenceKindReanalysis,
		RetrievedAt:   validAt.Add(48 * time.Hour),
		Timezone:      "Asia/Jakarta",
		Hourly: []HourlyWeather{
			{ValidAt: validAt},
			{ValidAt: validAt.Add(time.Hour)},
		},
	}
}
