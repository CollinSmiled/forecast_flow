import 'dart:async';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

import '../core/config/app_config.dart';
import '../core/theme/app_theme.dart';
import '../features/location/data/location_api_client.dart';
import '../features/location/data/models/location_result.dart';
import '../features/location/data/selected_location_store.dart';
import '../features/location/presentation/location_search_controller.dart';
import '../features/location/presentation/location_search_screen.dart';
import '../features/weather/data/forecast_api_client.dart';
import '../features/weather/presentation/weather_controller.dart';
import '../features/weather/presentation/weather_home_screen.dart';
import '../features/weather/presentation/weather_scene_resolver.dart';

class ForecastFlowApp extends StatefulWidget {
  const ForecastFlowApp({
    this.httpClient,
    this.apiBaseUri,
    this.selectedLocationStore,
    super.key,
  });

  final http.Client? httpClient;
  final Uri? apiBaseUri;
  final SelectedLocationStore? selectedLocationStore;

  @override
  State<ForecastFlowApp> createState() => _ForecastFlowAppState();
}

class _ForecastFlowAppState extends State<ForecastFlowApp> {
  late final http.Client _httpClient;
  late final bool _ownsHttpClient;
  late final LocationSearchController _locationController;
  late final WeatherController _weatherController;
  late final SelectedLocationStore _selectedLocationStore;

  Future<void> _pendingStorageWrite = Future.value();
  int? _selectedLocationId;
  List<LocationResult> _recentLocations = const [];
  bool _isRestoringLocation = true;

  @override
  void initState() {
    super.initState();
    _ownsHttpClient = widget.httpClient == null;
    _httpClient = widget.httpClient ?? http.Client();
    _selectedLocationStore =
        widget.selectedLocationStore ??
        SharedPreferencesSelectedLocationStore();
    final baseUri = widget.apiBaseUri ?? AppConfig.apiBaseUri;

    final locationApiClient = LocationApiClient(
      httpClient: _httpClient,
      baseUri: baseUri,
    );
    final forecastApiClient = ForecastApiClient(
      httpClient: _httpClient,
      baseUri: baseUri,
    );

    _locationController = LocationSearchController(
      searchAvailable: locationApiClient.searchAvailable,
      saveLocation: locationApiClient.add,
    );
    _weatherController = WeatherController(
      loadForecast: forecastApiClient.getLatest,
      pendingForecastRetries: 8,
      pendingForecastRetryDelay: const Duration(seconds: 10),
    );
    unawaited(_restoreSelectedLocation());
  }

  Future<void> _restoreSelectedLocation() async {
    int? locationId;
    List<LocationResult> recentLocations = const [];
    try {
      locationId = await _selectedLocationStore.readLocationId();
    } on Object {
      // Local preferences should never prevent the app from starting.
    }
    try {
      recentLocations = await _selectedLocationStore.readRecentLocations();
    } on Object {
      // A corrupt recent list should not affect the selected city.
    }

    if (!mounted) {
      return;
    }

    setState(() {
      _selectedLocationId = locationId;
      _recentLocations = recentLocations;
      _isRestoringLocation = false;
    });
    if (locationId != null) {
      unawaited(_weatherController.load(locationId));
    }
  }

  void _selectLocation(LocationResult location) {
    final locationId = location.locationId;
    if (locationId == null) {
      return;
    }

    setState(() {
      _selectedLocationId = locationId;
      _recentLocations = [
        location,
        ..._recentLocations.where((item) => item.locationId != locationId),
      ].take(5).toList(growable: false);
    });
    _queueStoredLocation(location);
    unawaited(_weatherController.load(locationId));
  }

  void _chooseAnotherLocation() {
    _locationController.reset();
    setState(() => _selectedLocationId = null);
    _queueStoredLocation(null);
  }

  void _queueStoredLocation(LocationResult? location) {
    _pendingStorageWrite = _pendingStorageWrite.then(
      (_) => _writeStoredLocation(location),
    );
  }

  Future<void> _writeStoredLocation(LocationResult? location) async {
    try {
      final locationId = location?.locationId;
      if (location == null || locationId == null) {
        await _selectedLocationStore.clearLocationId();
      } else {
        await _selectedLocationStore.saveLocationId(locationId);
        await _selectedLocationStore.saveRecentLocation(location);
      }
    } on Object {
      // Forecast loading remains usable if device storage is unavailable.
    }
  }

  @override
  void dispose() {
    _locationController.dispose();
    _weatherController.dispose();
    if (_ownsHttpClient) {
      _httpClient.close();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Forecast Flow',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      home: _isRestoringLocation
          ? const _RestoringLocationScreen()
          : _selectedLocationId == null
          ? LocationSearchScreen(
              controller: _locationController,
              recentLocations: _recentLocations,
              backgroundAssetResolver: _locationSearchBackgroundAsset,
              onLocationSelected: _selectLocation,
            )
          : WeatherHomeScreen(
              controller: _weatherController,
              onChooseLocation: _chooseAnotherLocation,
            ),
    );
  }

  String _locationSearchBackgroundAsset() {
    final state = _weatherController.state;
    if (state case WeatherLoaded(:final forecast)) {
      return WeatherSceneResolver.resolveForForecast(forecast);
    }

    return WeatherSceneResolver.resolveForLocalTime(DateTime.now());
  }
}

class _RestoringLocationScreen extends StatelessWidget {
  const _RestoringLocationScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: CircularProgressIndicator(
          key: ValueKey('restoring-selected-location'),
        ),
      ),
    );
  }
}
