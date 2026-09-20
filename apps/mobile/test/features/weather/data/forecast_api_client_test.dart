import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/forecast_api_client.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  const baseUri = 'http://api.example.test:8080';

  test('requests and parses the latest forecast', () async {
    final httpClient = MockClient((request) async {
      expect(request.url, Uri.parse('$baseUri/api/v1/locations/3/forecast'));
      expect(request.method, 'GET');
      expect(request.headers['Accept'], 'application/json');

      return http.Response(
        jsonEncode(_forecastEnvelope()),
        200,
        headers: {'content-type': 'application/json'},
      );
    });
    final client = ForecastApiClient(
      httpClient: httpClient,
      baseUri: Uri.parse(baseUri),
    );

    final forecast = await client.getLatest(3);

    expect(forecast.location.city, 'Jakarta');
    expect(forecast.current.weather.temperature2M, 31.1);
  });

  test('translates a 404 API response into a not-found exception', () async {
    final client = ForecastApiClient(
      httpClient: MockClient(
        (_) async => http.Response(
          jsonEncode({
            'error': {
              'code': 'forecast_not_found',
              'message': 'no forecast is available for this location',
            },
          }),
          404,
        ),
      ),
      baseUri: Uri.parse(baseUri),
    );

    expect(
      () => client.getLatest(99),
      throwsA(
        isA<ForecastNotFoundException>().having(
          (error) => error.message,
          'message',
          'no forecast is available for this location',
        ),
      ),
    );
  });

  test('translates transport failures into a network exception', () async {
    final client = ForecastApiClient(
      httpClient: MockClient((_) async {
        throw http.ClientException('connection refused');
      }),
      baseUri: Uri.parse(baseUri),
    );

    expect(() => client.getLatest(3), throwsA(isA<ForecastNetworkException>()));
  });

  test('rejects malformed successful responses', () async {
    final client = ForecastApiClient(
      httpClient: MockClient(
        (_) async => http.Response('{"data":"invalid"}', 200),
      ),
      baseUri: Uri.parse(baseUri),
    );

    expect(
      () => client.getLatest(3),
      throwsA(isA<InvalidForecastResponseException>()),
    );
  });

  test('rejects invalid location IDs before making a request', () async {
    var requested = false;
    final client = ForecastApiClient(
      httpClient: MockClient((_) async {
        requested = true;
        return http.Response('{}', 200);
      }),
      baseUri: Uri.parse(baseUri),
    );

    expect(() => client.getLatest(0), throwsArgumentError);
    expect(requested, isFalse);
  });
}

Map<String, dynamic> _forecastEnvelope() {
  return {
    'data': {
      'forecast_id': 'forecast-123',
      'location': {
        'location_id': 3,
        'open_meteo_location_id': 1642911,
        'city': 'Jakarta',
        'country': 'Indonesia',
        'country_code': 'ID',
        'latitude': -6.2146,
        'longitude': 106.8451,
        'timezone': 'Asia/Jakarta',
        'elevation': 8,
        'population': 10560000,
        'administrative_area': 'Jakarta',
      },
      'source': 'open_meteo_best_match',
      'retrieved_at': '2026-09-20T00:00:00Z',
      'timezone': 'Asia/Jakarta',
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
        'temperature_2m': 31.1,
        'weather_code': 2,
        'is_day': true,
      },
      'hourly': <Map<String, dynamic>>[],
      'daily': <Map<String, dynamic>>[],
    },
  };
}
