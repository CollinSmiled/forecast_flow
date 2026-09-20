import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/forecast_api_client.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_controller.dart';

void main() {
  test('moves from initial to loading to loaded', () async {
    final forecast = _forecast(locationId: 3, city: 'Jakarta');
    final completer = Completer<LatestForecast>();
    final controller = WeatherController(loadForecast: (_) => completer.future);

    expect(controller.state, isA<WeatherInitial>());

    final loading = controller.load(3);
    expect(controller.state, isA<WeatherLoading>());

    completer.complete(forecast);
    await loading;

    expect(
      controller.state,
      isA<WeatherLoaded>().having(
        (state) => state.forecast.location.city,
        'city',
        'Jakarta',
      ),
    );
  });

  test(
    'represents a missing forecast separately from other failures',
    () async {
      final controller = WeatherController(
        loadForecast: (_) async {
          throw const ForecastNotFoundException('forecast unavailable');
        },
      );

      await controller.load(99);

      expect(
        controller.state,
        isA<WeatherNotFound>().having(
          (state) => state.message,
          'message',
          'forecast unavailable',
        ),
      );
    },
  );

  test('marks network failures as retryable', () async {
    final controller = WeatherController(
      loadForecast: (_) async {
        throw const ForecastNetworkException(
          'service unreachable',
          cause: 'connection refused',
        );
      },
    );

    await controller.load(3);

    expect(
      controller.state,
      isA<WeatherFailure>()
          .having((state) => state.message, 'message', 'service unreachable')
          .having((state) => state.canRetry, 'canRetry', isTrue),
    );
  });

  test('retry uses the most recently selected location', () async {
    var attempts = 0;
    final controller = WeatherController(
      loadForecast: (locationId) async {
        attempts++;
        if (attempts == 1) {
          throw const ForecastNetworkException(
            'temporary failure',
            cause: 'offline',
          );
        }

        return _forecast(locationId: locationId, city: 'Jakarta');
      },
    );

    await controller.load(3);
    await controller.retry();

    expect(attempts, 2);
    expect(controller.state, isA<WeatherLoaded>());
  });

  test('an older response cannot replace a newer city selection', () async {
    final jakarta = Completer<LatestForecast>();
    final tokyo = Completer<LatestForecast>();
    final controller = WeatherController(
      loadForecast: (locationId) {
        return switch (locationId) {
          3 => jakarta.future,
          4 => tokyo.future,
          _ => throw StateError('unexpected location'),
        };
      },
    );

    final firstRequest = controller.load(3);
    final secondRequest = controller.load(4);

    tokyo.complete(_forecast(locationId: 4, city: 'Tokyo'));
    await secondRequest;
    jakarta.complete(_forecast(locationId: 3, city: 'Jakarta'));
    await firstRequest;

    expect(
      controller.state,
      isA<WeatherLoaded>().having(
        (state) => state.forecast.location.city,
        'city',
        'Tokyo',
      ),
    );
  });
}

LatestForecast _forecast({required int locationId, required String city}) {
  return LatestForecast(
    forecastId: 'forecast-$locationId',
    location: ForecastLocation(
      locationId: locationId,
      openMeteoLocationId: 1000 + locationId,
      city: city,
      country: locationId == 3 ? 'Indonesia' : 'Japan',
      countryCode: locationId == 3 ? 'ID' : 'JP',
      latitude: 0,
      longitude: 0,
      timezone: 'UTC',
    ),
    source: 'open_meteo_best_match',
    retrievedAt: DateTime.utc(2026, 9, 20),
    timezone: 'UTC',
    units: const ForecastUnits(
      temperature: 'celsius',
      precipitation: 'mm',
      windSpeed: 'km/h',
      pressure: 'hPa',
      visibility: 'm',
    ),
    current: CurrentForecast(
      validAt: DateTime.utc(2026, 9, 20),
      intervalSeconds: 900,
      weather: const WeatherMetrics(temperature2M: 31, weatherCode: 2),
    ),
    hourly: const [],
    daily: const [],
  );
}
