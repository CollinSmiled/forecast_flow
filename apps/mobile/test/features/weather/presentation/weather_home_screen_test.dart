import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/forecast_api_client.dart';
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

  testWidgets('does not overflow on a narrow phone with larger text', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(320, 640);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final controller = WeatherController(
      loadForecast: (_) async => _completeForecast(),
    );
    await controller.load(4);

    await tester.pumpWidget(
      MaterialApp(
        builder: (context, child) => MediaQuery(
          data: MediaQuery.of(
            context,
          ).copyWith(textScaler: const TextScaler.linear(1.3)),
          child: child!,
        ),
        home: WeatherHomeScreen(
          controller: controller,
          onChooseLocation: () {},
        ),
      ),
    );
    await tester.pump();

    expect(tester.takeException(), isNull);

    await tester.pumpWidget(const SizedBox.shrink());
    controller.dispose();
  });

  testWidgets('retries a recoverable failure when the app resumes', (
    tester,
  ) async {
    var requests = 0;
    final controller = WeatherController(
      loadForecast: (locationId) async {
        requests++;
        if (requests == 1) {
          throw const ForecastNetworkException(
            'service unreachable',
            cause: 'offline',
          );
        }

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

    expect(controller.state, isA<WeatherFailure>());
    expect(requests, 1);

    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump();

    expect(controller.state, isA<WeatherLoaded>());
    expect(requests, 2);

    await tester.pumpWidget(const SizedBox.shrink());
    controller.dispose();
  });

  testWidgets('does not retry a non-recoverable failure on resume', (
    tester,
  ) async {
    var requests = 0;
    final controller = WeatherController(
      loadForecast: (_) async {
        requests++;
        throw const ForecastHttpException(
          'invalid request',
          statusCode: 400,
        );
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

    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump();

    expect(controller.state, isA<WeatherFailure>());
    expect(requests, 1);

    await tester.pumpWidget(const SizedBox.shrink());
    controller.dispose();
  });
}

LatestForecast _completeForecast() {
  final now = DateTime.now().toUtc();
  final tomorrow = now.add(const Duration(days: 1));
  final date = [
    tomorrow.year.toString().padLeft(4, '0'),
    tomorrow.month.toString().padLeft(2, '0'),
    tomorrow.day.toString().padLeft(2, '0'),
  ].join('-');

  return LatestForecast(
    forecastId: 'forecast-complete',
    location: const ForecastLocation(
      locationId: 4,
      openMeteoLocationId: 1850147,
      city: 'Tokyo',
      country: 'Japan',
      countryCode: 'JP',
      latitude: 35.6762,
      longitude: 139.6503,
      timezone: 'Asia/Tokyo',
    ),
    source: 'open_meteo_best_match',
    retrievedAt: now,
    timezone: 'Asia/Tokyo',
    units: const ForecastUnits(
      temperature: 'celsius',
      precipitation: 'mm',
      windSpeed: 'km/h',
      pressure: 'hPa',
      visibility: 'm',
    ),
    current: CurrentForecast(
      validAt: now,
      intervalSeconds: 900,
      weather: const WeatherMetrics(
        temperature2M: 24,
        apparentTemperature: 25,
        relativeHumidity2M: 60,
        precipitationProbability: 25,
        weatherCode: 1,
        cloudCover: 35,
        pressureMsl: 1012,
        visibility: 10000,
        windSpeed10M: 8.5,
        windGusts10M: 18,
        uvIndex: 4.1,
        isDay: true,
      ),
    ),
    hourly: [
      HourlyForecast(
        validAt: now.add(const Duration(hours: 1)),
        weather: const WeatherMetrics(
          temperature2M: 25,
          precipitationProbability: 20,
          weatherCode: 1,
          isDay: true,
        ),
      ),
    ],
    daily: [
      DailyForecast(
        date: date,
        weatherCode: 1,
        temperature2MMax: 29,
        temperature2MMin: 20,
        precipitationProbabilityMax: 30,
        sunrise: tomorrow,
        sunset: tomorrow.add(const Duration(hours: 12)),
        daylightDurationSeconds: 43200,
      ),
    ],
  );
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
