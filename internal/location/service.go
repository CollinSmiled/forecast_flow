package location

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCountryCodeRequired = errors.New(
		"country code is required",
	)
	ErrCountryNotAllowed = errors.New(
		"country is not allowed",
	)
	ErrInvalidLocationID = errors.New(
		"Open-Meteo location ID must be positive",
	)
)

type Geocoder interface {
	SearchLocations(
		ctx context.Context,
		searchTerm string,
		countryCode string,
		limit int,
	) ([]Location, error)

	GetLocation(
		ctx context.Context,
		openMeteoLocationID int64,
	) (Location, error)
}

type Store interface {
	Upsert(
		ctx context.Context,
		candidate Location,
	) (Location, error)

	Search(
		ctx context.Context,
		searchTerm string,
		countryCode string,
		limit int,
	) ([]Location, error)
}

type Service struct {
	geocoder            Geocoder
	store               Store
	allowedCountryCodes map[string]struct{}
}

func NewService(
	geocoder Geocoder,
	store Store,
	allowedCountryCodes []string,
) (*Service, error) {
	if geocoder == nil {
		return nil, errors.New(
			"create location service: geocoder is required",
		)
	}

	if store == nil {
		return nil, errors.New(
			"create location service: store is required",
		)
	}

	if len(allowedCountryCodes) == 0 {
		return nil, errors.New(
			"create location service: at least one country is required",
		)
	}

	allowed := make(map[string]struct{}, len(allowedCountryCodes))

	for _, countryCode := range allowedCountryCodes {
		countryCode = strings.ToUpper(
			strings.TrimSpace(countryCode),
		)

		if len(countryCode) != 2 {
			return nil, fmt.Errorf(
				"create location service: invalid country code %q",
				countryCode,
			)
		}

		allowed[countryCode] = struct{}{}
	}

	return &Service{
		geocoder:            geocoder,
		store:               store,
		allowedCountryCodes: allowed,
	}, nil
}

func (service *Service) SearchAvailable(
	ctx context.Context,
	searchTerm string,
	countryCode string,
	limit int,
) ([]Location, error) {
	countryCode, err := service.allowedCountryCode(countryCode)
	if err != nil {
		return nil, err
	}

	candidates, err := service.geocoder.SearchLocations(
		ctx,
		searchTerm,
		countryCode,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"search available locations: %w",
			err,
		)
	}

	filtered := make([]Location, 0, len(candidates))

	for _, candidate := range candidates {
		if strings.EqualFold(
			candidate.CountryCode,
			countryCode,
		) {
			filtered = append(filtered, candidate)
		}
	}

	return filtered, nil
}

func (service *Service) Add(
	ctx context.Context,
	openMeteoLocationID int64,
) (Location, error) {
	if openMeteoLocationID <= 0 {
		return Location{}, ErrInvalidLocationID
	}

	candidate, err := service.geocoder.GetLocation(
		ctx,
		openMeteoLocationID,
	)
	if err != nil {
		return Location{}, fmt.Errorf(
			"get selected location: %w",
			err,
		)
	}

	if _, err := service.allowedCountryCode(
		candidate.CountryCode,
	); err != nil {
		return Location{}, err
	}

	saved, err := service.store.Upsert(ctx, candidate)
	if err != nil {
		return Location{}, fmt.Errorf(
			"save selected location: %w",
			err,
		)
	}

	return saved, nil
}

func (service *Service) SearchSaved(
	ctx context.Context,
	searchTerm string,
	countryCode string,
	limit int,
) ([]Location, error) {
	countryCode, err := service.allowedCountryCode(countryCode)
	if err != nil {
		return nil, err
	}

	saved, err := service.store.Search(
		ctx,
		searchTerm,
		countryCode,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"search saved locations: %w",
			err,
		)
	}

	return saved, nil
}

func (service *Service) allowedCountryCode(
	countryCode string,
) (string, error) {
	countryCode = strings.ToUpper(
		strings.TrimSpace(countryCode),
	)

	if countryCode == "" {
		return "", ErrCountryCodeRequired
	}

	if _, allowed := service.allowedCountryCodes[countryCode]; !allowed {
		return "", fmt.Errorf(
			"%w: %s",
			ErrCountryNotAllowed,
			countryCode,
		)
	}

	return countryCode, nil
}
