import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_daylight_resolver.dart';

void main() {
  test('uses city sunrise and sunset instead of stale current is_day', () {
    final forecast = _forecast(
      isDay: true,
      daily: [
        DailyForecast(
          date: '2026-09-20',
          sunrise: DateTime.utc(2026, 9, 19, 22, 43),
          sunset: DateTime.utc(2026, 9, 20, 10, 49),
        ),
      ],
    );

    expect(
      WeatherDaylightResolver.resolve(
        forecast,
        now: DateTime.utc(2026, 9, 20, 15, 35),
      ),
      isFalse,
    );
  });

  test('recognizes daytime between city sunrise and sunset', () {
    final forecast = _forecast(
      isDay: false,
      daily: [
        DailyForecast(
          date: '2026-09-20',
          sunrise: DateTime.utc(2026, 9, 19, 22, 43),
          sunset: DateTime.utc(2026, 9, 20, 10, 49),
        ),
      ],
    );

    expect(
      WeatherDaylightResolver.resolve(
        forecast,
        now: DateTime.utc(2026, 9, 20, 5),
      ),
      isTrue,
    );
  });

  test('falls back to the city clock when sun data is unavailable', () {
    final forecast = _forecast(isDay: true);

    expect(
      WeatherDaylightResolver.resolve(
        forecast,
        now: DateTime.utc(2026, 9, 20, 15),
      ),
      isFalse,
    );
  });

  test('falls back to API is_day for an unknown timezone', () {
    final forecast = _forecast(isDay: false, timezone: 'Unknown/Timezone');

    expect(
      WeatherDaylightResolver.resolve(
        forecast,
        now: DateTime.utc(2026, 9, 20, 5),
      ),
      isFalse,
    );
  });
}

LatestForecast _forecast({
  required bool isDay,
  String timezone = 'Asia/Jakarta',
  List<DailyForecast> daily = const [],
}) {
  return LatestForecast(
    forecastId: 'forecast-jakarta',
    location: ForecastLocation(
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
      weather: WeatherMetrics(isDay: isDay),
    ),
    hourly: const [],
    daily: daily,
  );
}
