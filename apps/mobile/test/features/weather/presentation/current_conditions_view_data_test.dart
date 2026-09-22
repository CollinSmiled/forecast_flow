import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/current_conditions_view_data.dart';

void main() {
  test('fills unavailable current display values from the nearest hour', () {
    final data = CurrentConditionsViewData.fromForecast(_forecast());

    expect(data.precipitationProbability, 40);
    expect(data.uvIndex, 6.2);
    expect(data.visibility, 12000);
  });

  test('prefers an available current value over the hourly fallback', () {
    final data = CurrentConditionsViewData(
      current: const WeatherMetrics(cloudCover: 75),
      nearestHourly: const WeatherMetrics(cloudCover: 20),
    );

    expect(data.cloudCover, 75);
  });
}

LatestForecast _forecast() {
  return LatestForecast(
    forecastId: 'forecast-3',
    location: const ForecastLocation(
      locationId: 3,
      openMeteoLocationId: 1642911,
      city: 'Jakarta',
      country: 'Indonesia',
      countryCode: 'ID',
      latitude: -6.2,
      longitude: 106.8,
      timezone: 'Asia/Jakarta',
    ),
    source: 'open_meteo_best_match',
    retrievedAt: DateTime.utc(2026, 9, 20),
    timezone: 'Asia/Jakarta',
    units: const ForecastUnits(
      temperature: 'celsius',
      precipitation: 'mm',
      windSpeed: 'km/h',
      pressure: 'hPa',
      visibility: 'm',
    ),
    current: CurrentForecast(
      validAt: DateTime.utc(2026, 9, 20, 0, 10),
      intervalSeconds: 900,
      weather: const WeatherMetrics(),
    ),
    hourly: [
      HourlyForecast(
        validAt: DateTime.utc(2026, 9, 20),
        weather: const WeatherMetrics(
          precipitationProbability: 40,
          uvIndex: 6.2,
          visibility: 12000,
        ),
      ),
      HourlyForecast(
        validAt: DateTime.utc(2026, 9, 20, 1),
        weather: const WeatherMetrics(
          precipitationProbability: 80,
          uvIndex: 8,
          visibility: 8000,
        ),
      ),
    ],
    daily: const [],
  );
}
