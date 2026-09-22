import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/location/data/models/location_result.dart';
import 'package:forecast_flow_mobile/features/location/domain/supported_country.dart';
import 'package:forecast_flow_mobile/features/location/presentation/location_search_controller.dart';
import 'package:forecast_flow_mobile/features/location/presentation/location_search_screen.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_scene_resolver.dart';

void main() {
  testWidgets('searches for and selects a city', (tester) async {
    final candidate = _location();
    final saved = _location(locationId: 4);
    LocationResult? selected;
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) async {
        expect(query, 'Tokyo');
        expect(country, SupportedCountry.malaysia);
        return [candidate];
      },
      saveLocation: (_) async => saved,
    );

    await tester.pumpWidget(
      MaterialApp(
        home: LocationSearchScreen(
          controller: controller,
          backgroundAssetResolver: () => WeatherSceneResolver.eveningAsset,
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
    final scene = tester.widget<Image>(
      find.byKey(const ValueKey('location-scene-background')),
    );
    expect(
      (scene.image as AssetImage).assetName,
      WeatherSceneResolver.eveningAsset,
    );
    expect(find.text('Find the forecast that matters to you.'), findsNothing);
    expect(find.text('Find your city'), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('country-selector')));
    await tester.pumpAndSettle();

    expect(find.byKey(const ValueKey('country-picker-sheet')), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('country-option-MY')));
    await tester.pumpAndSettle();

    expect(find.text('Malaysia'), findsOneWidget);
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

  testWidgets('does not overflow when the keyboard opens', (tester) async {
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) async =>
          [],
      saveLocation: (_) async => _location(locationId: 4),
    );
    tester.view.physicalSize = const Size(480, 844);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    addTearDown(tester.view.resetViewInsets);

    await tester.pumpWidget(
      MaterialApp(
        home: LocationSearchScreen(
          controller: controller,
          recentLocations: [_location(locationId: 4)],
          onLocationSelected: (_) {},
        ),
      ),
    );
    await tester.tap(find.byType(TextField));
    tester.view.viewInsets = const FakeViewPadding(bottom: 320);
    await tester.pump();

    expect(tester.takeException(), isNull);
  });

  testWidgets('styles recent cities with the weather screen typography', (
    tester,
  ) async {
    final controller = LocationSearchController(
      searchAvailable: ({required query, required country, limit = 10}) async =>
          [],
      saveLocation: (_) async => _location(locationId: 4),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: LocationSearchScreen(
          controller: controller,
          recentLocations: [_location(locationId: 4)],
          onLocationSelected: (_) {},
        ),
      ),
    );

    final heading = tester.widget<Text>(find.text('Recent cities'));
    expect(heading.style?.fontSize, 18);
    expect(heading.style?.fontWeight, FontWeight.w700);
    expect(find.byKey(const ValueKey('recent-location-4')), findsOneWidget);
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
