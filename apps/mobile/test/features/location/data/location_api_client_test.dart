import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/location/data/location_api_client.dart';
import 'package:forecast_flow_mobile/features/location/domain/supported_country.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  const baseUri = 'http://api.example.test:8080';

  test('searches available locations with encoded query parameters', () async {
    final client = LocationApiClient(
      httpClient: MockClient((request) async {
        expect(request.url.path, '/api/v1/locations/search');
        expect(request.url.queryParameters['q'], 'Ho Chi Minh');
        expect(request.url.queryParameters['country_code'], 'VN');
        expect(request.url.queryParameters['limit'], '5');
        expect(request.headers['Accept'], 'application/json');

        return http.Response(jsonEncode(_locationList(saved: false)), 200);
      }),
      baseUri: Uri.parse(baseUri),
    );

    final results = await client.searchAvailable(
      query: '  Ho Chi Minh  ',
      country: SupportedCountry.vietnam,
      limit: 5,
    );

    expect(results, hasLength(1));
    expect(results.single.city, 'Ho Chi Minh City');
    expect(results.single.locationId, isNull);
    expect(results.single.isSaved, isFalse);
  });

  test('searches saved locations through the saved endpoint', () async {
    final client = LocationApiClient(
      httpClient: MockClient((request) async {
        expect(request.url.path, '/api/v1/locations');
        return http.Response(jsonEncode(_locationList(saved: true)), 200);
      }),
      baseUri: Uri.parse(baseUri),
    );

    final results = await client.searchSaved(
      query: 'Ho Chi Minh',
      country: SupportedCountry.vietnam,
    );

    expect(results.single.locationId, 8);
    expect(results.single.isSaved, isTrue);
  });

  test('adds a geocoding result to the saved location catalog', () async {
    final client = LocationApiClient(
      httpClient: MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/v1/locations');
        expect(request.headers['Content-Type'], 'application/json');
        expect(jsonDecode(request.body), {'open_meteo_location_id': 1566083});

        final location = _locationList(saved: true)['data'] as List<dynamic>;
        return http.Response(jsonEncode({'data': location.single}), 201);
      }),
      baseUri: Uri.parse(baseUri),
    );

    final saved = await client.add(1566083);

    expect(saved.locationId, 8);
    expect(saved.openMeteoLocationId, 1566083);
  });

  test('preserves a structured backend error message', () async {
    final client = LocationApiClient(
      httpClient: MockClient(
        (_) async => http.Response(
          jsonEncode({
            'error': {
              'code': 'invalid_request',
              'message': 'country is not supported',
            },
          }),
          400,
        ),
      ),
      baseUri: Uri.parse(baseUri),
    );

    expect(
      () => client.searchAvailable(
        query: 'Delhi',
        country: SupportedCountry.indonesia,
      ),
      throwsA(
        isA<LocationHttpException>().having(
          (error) => error.message,
          'message',
          'country is not supported',
        ),
      ),
    );
  });

  test('rejects an empty search before making a request', () async {
    var requested = false;
    final client = LocationApiClient(
      httpClient: MockClient((_) async {
        requested = true;
        return http.Response('{}', 200);
      }),
      baseUri: Uri.parse(baseUri),
    );

    expect(
      () => client.searchAvailable(
        query: '   ',
        country: SupportedCountry.indonesia,
      ),
      throwsArgumentError,
    );
    expect(requested, isFalse);
  });
}

Map<String, dynamic> _locationList({required bool saved}) {
  return {
    'data': [
      {
        if (saved) 'location_id': 8,
        'open_meteo_location_id': 1566083,
        'city': 'Ho Chi Minh City',
        'country': 'Vietnam',
        'country_code': 'VN',
        'latitude': 10.8231,
        'longitude': 106.6297,
        'timezone': 'Asia/Ho_Chi_Minh',
        'elevation': 19,
        'population': 8990000,
        'administrative_area': 'Ho Chi Minh',
      },
    ],
  };
}
