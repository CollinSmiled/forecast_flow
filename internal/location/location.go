package location

import "time"

type Location struct {
	ID                  int64
	OpenMeteoLocationID int64

	City        string
	Country     string
	CountryCode string

	Latitude  float64
	Longitude float64
	Timezone  string

	Elevation          *float64
	Population         *int64
	AdministrativeArea *string

	CreatedAt time.Time
	UpdatedAt time.Time
}
