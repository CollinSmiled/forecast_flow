import 'dart:async';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

import '../core/config/app_config.dart';
import '../core/theme/app_theme.dart';
import '../features/location/data/location_api_client.dart';
import '../features/location/data/models/location_result.dart';
import '../features/location/presentation/location_search_controller.dart';
import '../features/location/presentation/location_search_screen.dart';
import '../features/weather/data/forecast_api_client.dart';
import '../features/weather/presentation/weather_controller.dart';
import '../features/weather/presentation/weather_home_screen.dart';

class ForecastFlowApp extends StatefulWidget {
  const ForecastFlowApp({this.httpClient, this.apiBaseUri, super.key});

  final http.Client? httpClient;
  final Uri? apiBaseUri;

  @override
  State<ForecastFlowApp> createState() => _ForecastFlowAppState();
}

class _ForecastFlowAppState extends State<ForecastFlowApp> {
  late final http.Client _httpClient;
  late final bool _ownsHttpClient;
  late final LocationSearchController _locationController;
  late final WeatherController _weatherController;

  LocationResult? _selectedLocation;

  @override
  void initState() {
    super.initState();
    _ownsHttpClient = widget.httpClient == null;
    _httpClient = widget.httpClient ?? http.Client();
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
    );
  }

  void _selectLocation(LocationResult location) {
    final locationId = location.locationId;
    if (locationId == null) {
      return;
    }

    setState(() => _selectedLocation = location);
    unawaited(_weatherController.load(locationId));
  }

  void _chooseAnotherLocation() {
    _locationController.reset();
    setState(() => _selectedLocation = null);
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
      home: _selectedLocation == null
          ? LocationSearchScreen(
              controller: _locationController,
              onLocationSelected: _selectLocation,
            )
          : WeatherHomeScreen(
              controller: _weatherController,
              onChooseLocation: _chooseAnotherLocation,
            ),
    );
  }
}
