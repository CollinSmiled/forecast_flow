import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/forecast_freshness.dart';

void main() {
  final localizations = DefaultMaterialLocalizations();

  test('formats a fresh update from today in the city timezone', () {
    final freshness = ForecastFreshness.fromForecast(
      _forecast(retrievedAt: DateTime.utc(2026, 9, 20, 7, 20)),
      now: DateTime.utc(2026, 9, 20, 8),
    );

    expect(freshness.isStale, isFalse);
    expect(freshness.updatedLabel(localizations), 'Updated 2:20 PM');
  });

  test('calls out an update from yesterday', () {
    final freshness = ForecastFreshness.fromForecast(
      _forecast(retrievedAt: DateTime.utc(2026, 9, 20, 7, 20)),
      now: DateTime.utc(2026, 9, 21, 3),
    );

    expect(freshness.isStale, isTrue);
    expect(
      freshness.updatedLabel(localizations),
      'Updated yesterday at 2:20 PM',
    );
  });

  test('includes the date for an older update', () {
    final freshness = ForecastFreshness.fromForecast(
      _forecast(retrievedAt: DateTime.utc(2026, 9, 20, 7, 20)),
      now: DateTime.utc(2026, 9, 23, 3),
    );

    expect(
      freshness.updatedLabel(localizations),
      'Updated Sun, Sep 20 at 2:20 PM',
    );
  });

  test('marks data stale only after the two hour threshold', () {
    final forecast = _forecast(retrievedAt: DateTime.utc(2026, 9, 20, 7, 20));

    expect(
      ForecastFreshness.fromForecast(
        forecast,
        now: DateTime.utc(2026, 9, 20, 9, 20),
      ).isStale,
      isFalse,
    );
    expect(
      ForecastFreshness.fromForecast(
        forecast,
        now: DateTime.utc(2026, 9, 20, 9, 21),
      ).isStale,
      isTrue,
    );
  });
}

LatestForecast _forecast({required DateTime retrievedAt}) {
  const timezone = 'Asia/Jakarta';
  return LatestForecast(
    forecastId: 'forecast-jakarta',
    location: const ForecastLocation(
      locationId: 3,
      openMeteoLocationId: 1642911,
      city: 'Jakarta',
      country: 'Indonesia',
      countryCode: 'ID',
      latitude: -6.2,
      longitude: 106.8,
      timezone: timezone,
    ),
    source: 'open_meteo_best_match',
    retrievedAt: retrievedAt,
    timezone: timezone,
    units: const ForecastUnits(
      temperature: 'celsius',
      precipitation: 'mm',
      windSpeed: 'km/h',
      pressure: 'hPa',
      visibility: 'm',
    ),
    current: CurrentForecast(
      validAt: retrievedAt,
      intervalSeconds: 900,
      weather: const WeatherMetrics(),
    ),
    hourly: const [],
    daily: const [],
  );
}
