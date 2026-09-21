import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/app/app.dart';
import 'package:forecast_flow_mobile/features/location/data/models/location_result.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_scene_resolver.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import '../helpers/memory_selected_location_store.dart';

void main() {
  testWidgets('selects a city and displays its live forecast response', (
    tester,
  ) async {
    final selectedLocationStore = MemorySelectedLocationStore();
    final httpClient = MockClient((request) async {
      if (request.method == 'GET' &&
          request.url.path == '/api/v1/locations/search') {
        return http.Response(
          jsonEncode({
            'data': [_locationJson(saved: false)],
          }),
          200,
        );
      }
      if (request.method == 'POST' && request.url.path == '/api/v1/locations') {
        return http.Response(
          jsonEncode({'data': _locationJson(saved: true)}),
          201,
        );
      }
      if (request.method == 'GET' &&
          request.url.path == '/api/v1/locations/4/forecast') {
        return http.Response(jsonEncode(_forecastEnvelope()), 200);
      }

      return http.Response('not found', 404);
    });

    await tester.pumpWidget(
      ForecastFlowApp(
        httpClient: httpClient,
        apiBaseUri: Uri.parse('http://api.example.test:8080'),
        selectedLocationStore: selectedLocationStore,
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(find.byType(TextField), 'Tokyo');
    await tester.tap(find.text('Search cities'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('location-1850147')));
    await tester.pumpAndSettle();

    expect(find.text('Tokyo, Japan'), findsOneWidget);
    final scene = tester.widget<Image>(
      find.byKey(const ValueKey('weather-scene-background')),
    );
    expect(
      (scene.image as AssetImage).assetName,
      isIn([
        WeatherSceneResolver.morningAsset,
        WeatherSceneResolver.dayAsset,
        WeatherSceneResolver.eveningAsset,
        WeatherSceneResolver.nightAsset,
      ]),
    );
    final currentTemperature = tester.widget<Text>(
      find.byKey(const ValueKey('current-temperature')),
    );
    expect(currentTemperature.data, '24°');
    expect(find.byKey(const ValueKey('forecast-updated-at')), findsOneWidget);
    expect(find.byKey(const ValueKey('stale-forecast-notice')), findsOneWidget);
    expect(find.byType(RefreshIndicator), findsOneWidget);
    expect(find.text('Clear sky'), findsOneWidget);
    expect(find.text('60%'), findsOneWidget);
    expect(find.text('8.5 km/h'), findsOneWidget);
    expect(find.text('Hourly forecast'), findsOneWidget);
    expect(find.text('10 AM'), findsOneWidget);
    expect(find.text('Weather details'), findsOneWidget);
    expect(find.text('25%'), findsOneWidget);
    expect(find.text('1012 hPa'), findsOneWidget);
    expect(find.text('10 km'), findsOneWidget);
    expect(find.text('18.0 km/h'), findsOneWidget);
    expect(find.text('Sun and daylight'), findsOneWidget);
    expect(find.text('1-day forecast'), findsOneWidget);
    expect(selectedLocationStore.locationId, 4);
    expect(selectedLocationStore.saveCount, 1);

    await tester.tap(find.byKey(const ValueKey('choose-location')));
    await tester.pumpAndSettle();

    expect(find.text('Choose a city'), findsOneWidget);
    final searchScene = tester.widget<Image>(
      find.byKey(const ValueKey('location-scene-background')),
    );
    expect(
      (searchScene.image as AssetImage).assetName,
      (scene.image as AssetImage).assetName,
    );
    expect(find.text('Recent cities'), findsOneWidget);
    expect(find.byKey(const ValueKey('recent-location-4')), findsOneWidget);
    expect(selectedLocationStore.locationId, isNull);
    expect(selectedLocationStore.clearCount, 1);

    await tester.tap(find.byKey(const ValueKey('recent-location-4')));
    await tester.pumpAndSettle();

    expect(find.text('Tokyo, Japan'), findsOneWidget);
    expect(selectedLocationStore.locationId, 4);
    expect(selectedLocationStore.recentLocations.single.city, 'Tokyo');
    expect(selectedLocationStore.recentSaveCount, 2);
  });

  testWidgets('restores the last selected city when the app starts', (
    tester,
  ) async {
    final selectedLocationStore = MemorySelectedLocationStore(locationId: 4);
    var forecastRequests = 0;
    final httpClient = MockClient((request) async {
      if (request.method == 'GET' &&
          request.url.path == '/api/v1/locations/4/forecast') {
        forecastRequests++;
        return http.Response(jsonEncode(_forecastEnvelope()), 200);
      }

      return http.Response('not found', 404);
    });

    await tester.pumpWidget(
      ForecastFlowApp(
        httpClient: httpClient,
        apiBaseUri: Uri.parse('http://api.example.test:8080'),
        selectedLocationStore: selectedLocationStore,
      ),
    );

    expect(
      find.byKey(const ValueKey('restoring-selected-location')),
      findsOneWidget,
    );
    await tester.pumpAndSettle();

    expect(find.text('Tokyo, Japan'), findsOneWidget);
    final currentTemperature = tester.widget<Text>(
      find.byKey(const ValueKey('current-temperature')),
    );
    expect(currentTemperature.data, '24°');
    expect(selectedLocationStore.readCount, 1);
    expect(forecastRequests, 1);
  });

  testWidgets('restores recent cities for quick switching after restart', (
    tester,
  ) async {
    final recentTokyo = LocationResult.fromJson(_locationJson(saved: true));
    final selectedLocationStore = MemorySelectedLocationStore(
      recent: [recentTokyo],
    );
    final httpClient = MockClient((request) async {
      if (request.method == 'GET' &&
          request.url.path == '/api/v1/locations/4/forecast') {
        return http.Response(jsonEncode(_forecastEnvelope()), 200);
      }

      return http.Response('not found', 404);
    });

    await tester.pumpWidget(
      ForecastFlowApp(
        httpClient: httpClient,
        apiBaseUri: Uri.parse('http://api.example.test:8080'),
        selectedLocationStore: selectedLocationStore,
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Recent cities'), findsOneWidget);
    expect(find.byKey(const ValueKey('recent-location-4')), findsOneWidget);

    await tester.tap(find.byKey(const ValueKey('recent-location-4')));
    await tester.pumpAndSettle();

    expect(find.text('Tokyo, Japan'), findsOneWidget);
    expect(selectedLocationStore.locationId, 4);
  });
}

Map<String, dynamic> _locationJson({required bool saved}) {
  return {
    if (saved) 'location_id': 4,
    'open_meteo_location_id': 1850147,
    'city': 'Tokyo',
    'country': 'Japan',
    'country_code': 'JP',
    'latitude': 35.6762,
    'longitude': 139.6503,
    'timezone': 'Asia/Tokyo',
    'elevation': 40,
    'population': 14094034,
    'administrative_area': 'Tokyo',
  };
}

Map<String, dynamic> _forecastEnvelope() {
  return {
    'data': {
      'forecast_id': 'forecast-tokyo',
      'location': _locationJson(saved: true),
      'source': 'open_meteo_best_match',
      'retrieved_at': '2026-09-20T00:00:00Z',
      'timezone': 'Asia/Tokyo',
      'units': {
        'temperature': 'celsius',
        'precipitation': 'mm',
        'wind_speed': 'km/h',
        'pressure': 'hPa',
        'visibility': 'm',
      },
      'current': {
        'valid_at': '2026-09-20T01:00:00Z',
        'interval_seconds': 900,
        'temperature_2m': 24.2,
        'apparent_temperature': 25.1,
        'relative_humidity_2m': 60,
        'weather_code': 0,
        'cloud_cover': 25,
        'pressure_msl': 1012,
        'wind_speed_10m': 8.5,
        'wind_gusts_10m': 18,
        'is_day': true,
      },
      'hourly': [
        {
          'valid_at': '2026-09-20T01:00:00Z',
          'temperature_2m': 24.2,
          'precipitation_probability': 10,
          'weather_code': 0,
          'visibility': 10000,
          'uv_index': 4.1,
          'is_day': true,
        },
      ],
      'daily': [
        {
          'date': '2026-09-20',
          'weather_code': 0,
          'temperature_2m_max': 27,
          'temperature_2m_min': 20,
          'precipitation_probability_max': 10,
          'sunrise': '2026-09-19T20:28:00Z',
          'sunset': '2026-09-20T08:40:00Z',
          'daylight_duration_seconds': 43920,
        },
      ],
    },
  };
}
