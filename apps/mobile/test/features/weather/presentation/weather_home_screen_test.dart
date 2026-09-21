import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_controller.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_home_screen.dart';

void main() {
  testWidgets('refreshes a loaded forecast on resume and while open', (
    tester,
  ) async {
    var requests = 0;
    final controller = WeatherController(
      loadForecast: (locationId) async {
        requests++;
        return _forecast(locationId, request: requests);
      },
    );
    await controller.load(4);

    await tester.pumpWidget(
      MaterialApp(
        home: WeatherHomeScreen(
          controller: controller,
          onChooseLocation: () {},
        ),
      ),
    );

    expect(requests, 1);

    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump();

    expect(requests, 2);

    await tester.pump(const Duration(minutes: 15));
    await tester.pump();

    expect(requests, 3);

    await tester.pumpWidget(const SizedBox.shrink());
    controller.dispose();
  });
}

LatestForecast _forecast(int locationId, {required int request}) {
  return LatestForecast(
    forecastId: 'forecast-$locationId-$request',
    location: ForecastLocation(
      locationId: locationId,
      openMeteoLocationId: 1850147,
      city: 'Tokyo',
      country: 'Japan',
      countryCode: 'JP',
      latitude: 35.6762,
      longitude: 139.6503,
      timezone: 'Asia/Tokyo',
    ),
    source: 'open_meteo_best_match',
    retrievedAt: DateTime.utc(2026, 9, 20, request),
    timezone: 'Asia/Tokyo',
    units: const ForecastUnits(
      temperature: 'celsius',
      precipitation: 'mm',
      windSpeed: 'km/h',
      pressure: 'hPa',
      visibility: 'm',
    ),
    current: CurrentForecast(
      validAt: DateTime.utc(2026, 9, 20, request),
      intervalSeconds: 900,
      weather: const WeatherMetrics(
        temperature2M: 24,
        apparentTemperature: 25,
        relativeHumidity2M: 60,
        weatherCode: 0,
        windSpeed10M: 8,
        isDay: true,
      ),
    ),
    hourly: const [],
    daily: const [],
  );
}
