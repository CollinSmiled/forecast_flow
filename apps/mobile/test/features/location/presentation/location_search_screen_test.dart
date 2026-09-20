import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/location/data/models/location_result.dart';
import 'package:forecast_flow_mobile/features/location/presentation/location_search_controller.dart';
import 'package:forecast_flow_mobile/features/location/presentation/location_search_screen.dart';

void main() {
  testWidgets('searches for and selects a city', (tester) async {
    final candidate = _location();
    final saved = _location(locationId: 4);
    LocationResult? selected;
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) async {
        expect(query, 'Tokyo');
        return [candidate];
      },
      saveLocation: (_) async => saved,
    );

    await tester.pumpWidget(
      MaterialApp(
        home: LocationSearchScreen(
          controller: controller,
          onLocationSelected: (location) => selected = location,
        ),
      ),
    );

    expect(
      find.byKey(const ValueKey('location-scene-background')),
      findsOneWidget,
    );
    expect(find.byKey(const ValueKey('location-search-form')), findsOneWidget);
    expect(
      find.byKey(const ValueKey('location-results-panel')),
      findsOneWidget,
    );
    expect(find.text('Find your city'), findsOneWidget);
    await tester.enterText(find.byType(TextField), 'Tokyo');
    await tester.tap(find.text('Search cities'));
    await tester.pumpAndSettle();

    expect(find.text('1 result'), findsOneWidget);
    final resultTile = find.byKey(const ValueKey('location-1850147'));
    expect(resultTile, findsOneWidget);

    await tester.tap(resultTile);
    await tester.pumpAndSettle();

    expect(selected?.locationId, 4);
  });

  testWidgets('validates a blank city without calling the API', (tester) async {
    var searched = false;
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) async {
        searched = true;
        return [];
      },
      saveLocation: (_) async => _location(locationId: 4),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: LocationSearchScreen(
          controller: controller,
          onLocationSelected: (_) {},
        ),
      ),
    );
    await tester.tap(find.text('Search cities'));
    await tester.pump();

    expect(find.text('Enter a city name'), findsOneWidget);
    expect(searched, isFalse);
  });
}

LocationResult _location({int? locationId}) {
  return LocationResult(
    locationId: locationId,
    openMeteoLocationId: 1850147,
    city: 'Tokyo',
    country: 'Japan',
    countryCode: 'JP',
    latitude: 35.6762,
    longitude: 139.6503,
    timezone: 'Asia/Tokyo',
    administrativeArea: 'Tokyo',
  );
}
