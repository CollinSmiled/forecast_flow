import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_scene_resolver.dart';

void main() {
  test('maps every time period to its scene asset', () {
    expect(
      WeatherSceneResolver.resolve(WeatherScenePeriod.morning),
      WeatherSceneResolver.morningAsset,
    );
    expect(
      WeatherSceneResolver.resolve(WeatherScenePeriod.day),
      WeatherSceneResolver.dayAsset,
    );
    expect(
      WeatherSceneResolver.resolve(WeatherScenePeriod.evening),
      WeatherSceneResolver.eveningAsset,
    );
    expect(
      WeatherSceneResolver.resolve(WeatherScenePeriod.night),
      WeatherSceneResolver.nightAsset,
    );
  });

  test('uses four fallback periods from the local clock', () {
    expect(
      WeatherSceneResolver.periodForLocalTime(DateTime(2026, 9, 20, 6)),
      WeatherScenePeriod.morning,
    );
    expect(
      WeatherSceneResolver.periodForLocalTime(DateTime(2026, 9, 20, 12)),
      WeatherScenePeriod.day,
    );
    expect(
      WeatherSceneResolver.periodForLocalTime(DateTime(2026, 9, 20, 18)),
      WeatherScenePeriod.evening,
    );
    expect(
      WeatherSceneResolver.periodForLocalTime(DateTime(2026, 9, 20, 22)),
      WeatherScenePeriod.night,
    );
  });

  test('uses city sunrise and sunset to choose all four periods', () {
    final forecast = _forecast();

    expect(
      WeatherSceneResolver.periodForForecast(
        forecast,
        now: DateTime.utc(2026, 9, 19, 23, 30),
      ),
      WeatherScenePeriod.morning,
    );
    expect(
      WeatherSceneResolver.periodForForecast(
        forecast,
        now: DateTime.utc(2026, 9, 20, 5),
      ),
      WeatherScenePeriod.day,
    );
    expect(
      WeatherSceneResolver.periodForForecast(
        forecast,
        now: DateTime.utc(2026, 9, 20, 10),
      ),
      WeatherScenePeriod.evening,
    );
    expect(
      WeatherSceneResolver.periodForForecast(
        forecast,
        now: DateTime.utc(2026, 9, 20, 15, 35),
      ),
      WeatherScenePeriod.night,
    );
  });
}

LatestForecast _forecast() {
  const timezone = 'Asia/Jakarta';
  return LatestForecast(
    forecastId: 'forecast-jakarta',
    location: const ForecastLocation(
      locationId: 4,
      openMeteoLocationId: 1642911,
      city: 'Jakarta',
      country: 'Indonesia',
      countryCode: 'ID',
      latitude: -6.2,
      longitude: 106.8,
      timezone: timezone,
    ),
    source: 'open_meteo_best_match',
    retrievedAt: DateTime.utc(2026, 9, 20, 7, 20),
    timezone: timezone,
    units: const ForecastUnits(
      temperature: 'celsius',
      precipitation: 'mm',
      windSpeed: 'km/h',
      pressure: 'hPa',
      visibility: 'm',
    ),
    current: CurrentForecast(
      validAt: DateTime.utc(2026, 9, 20, 7, 15),
      intervalSeconds: 900,
      weather: const WeatherMetrics(isDay: true),
    ),
    hourly: const [],
    daily: [
      DailyForecast(
        date: '2026-09-20',
        sunrise: DateTime.utc(2026, 9, 19, 22, 43),
        sunset: DateTime.utc(2026, 9, 20, 10, 49),
      ),
    ],
  );
}
