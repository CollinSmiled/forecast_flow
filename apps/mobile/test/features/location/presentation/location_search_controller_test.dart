import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/location/data/location_api_client.dart';
import 'package:forecast_flow_mobile/features/location/data/models/location_result.dart';
import 'package:forecast_flow_mobile/features/location/domain/supported_country.dart';
import 'package:forecast_flow_mobile/features/location/presentation/location_search_controller.dart';

void main() {
  test('moves from searching to results', () async {
    final completer = Completer<List<LocationResult>>();
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) {
        expect(query, 'Tokyo');
        expect(country, SupportedCountry.japan);
        expect(limit, 10);
        return completer.future;
      },
      saveLocation: (_) async => _location(saved: true),
    );

    final search = controller.search(
      query: '  Tokyo  ',
      country: SupportedCountry.japan,
    );
    expect(controller.state.status, LocationSearchStatus.searching);

    completer.complete([_location()]);
    await search;

    expect(controller.state.status, LocationSearchStatus.results);
    expect(controller.state.results.single.city, 'Tokyo');
  });

  test('represents an empty result without treating it as an error', () async {
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) async {
        return [];
      },
      saveLocation: (_) async => _location(saved: true),
    );

    await controller.search(query: 'Unknown', country: SupportedCountry.japan);

    expect(controller.state.status, LocationSearchStatus.empty);
    expect(controller.state.errorMessage, isNull);
  });

  test('saves a candidate and replaces it with the catalog location', () async {
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) async {
        return [_location()];
      },
      saveLocation: (_) async => _location(saved: true),
    );
    await controller.search(query: 'Tokyo', country: SupportedCountry.japan);

    final saved = await controller.save(controller.state.results.single);

    expect(saved?.locationId, 4);
    expect(controller.state.results.single.locationId, 4);
    expect(controller.state.isSaving, isFalse);
  });

  test('preserves a save error alongside the search results', () async {
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) async {
        return [_location()];
      },
      saveLocation: (_) async {
        throw const LocationNetworkException(
          'service unreachable',
          cause: 'offline',
        );
      },
    );
    await controller.search(query: 'Tokyo', country: SupportedCountry.japan);

    final saved = await controller.save(controller.state.results.single);

    expect(saved, isNull);
    expect(controller.state.status, LocationSearchStatus.results);
    expect(controller.state.results, hasLength(1));
    expect(controller.state.errorMessage, 'service unreachable');
  });

  test('an older search cannot replace newer results', () async {
    final tokyo = Completer<List<LocationResult>>();
    final osaka = Completer<List<LocationResult>>();
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) {
        return query == 'Tokyo' ? tokyo.future : osaka.future;
      },
      saveLocation: (_) async => _location(saved: true),
    );

    final first = controller.search(
      query: 'Tokyo',
      country: SupportedCountry.japan,
    );
    final second = controller.search(
      query: 'Osaka',
      country: SupportedCountry.japan,
    );
    osaka.complete([_location(city: 'Osaka')]);
    await second;
    tokyo.complete([_location()]);
    await first;

    expect(controller.state.query, 'Osaka');
    expect(controller.state.results.single.city, 'Osaka');
  });
}

LocationResult _location({String city = 'Tokyo', bool saved = false}) {
  return LocationResult(
    locationId: saved ? 4 : null,
    openMeteoLocationId: 1850147,
    city: city,
    country: 'Japan',
    countryCode: 'JP',
    latitude: 35.6762,
    longitude: 139.6503,
    timezone: 'Asia/Tokyo',
    administrativeArea: 'Tokyo',
  );
}
