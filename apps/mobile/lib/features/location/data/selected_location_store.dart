import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

import 'models/location_result.dart';

abstract interface class SelectedLocationStore {
  Future<int?> readLocationId();

  Future<List<LocationResult>> readRecentLocations();

  Future<void> saveLocationId(int locationId);

  Future<void> saveRecentLocation(LocationResult location);

  Future<void> clearLocationId();
}

class SharedPreferencesSelectedLocationStore implements SelectedLocationStore {
  SharedPreferencesSelectedLocationStore({SharedPreferencesAsync? preferences})
    : _preferences = preferences ?? SharedPreferencesAsync();

  static const _locationIdKey = 'selected_location_id';
  static const _recentLocationsKey = 'recent_locations';
  static const _maximumRecentLocations = 5;

  final SharedPreferencesAsync _preferences;

  @override
  Future<int?> readLocationId() async {
    final locationId = await _preferences.getInt(_locationIdKey);
    return locationId != null && locationId > 0 ? locationId : null;
  }

  @override
  Future<List<LocationResult>> readRecentLocations() async {
    final encodedLocations =
        await _preferences.getStringList(_recentLocationsKey) ?? const [];
    final locations = <LocationResult>[];

    for (final encodedLocation in encodedLocations) {
      try {
        final decoded = jsonDecode(encodedLocation);
        if (decoded is Map<String, dynamic>) {
          final location = LocationResult.fromJson(decoded);
          if (location.isSaved) {
            locations.add(location);
          }
        }
      } on FormatException {
        // Ignore one corrupt entry instead of losing the whole recent list.
      }
    }

    return List.unmodifiable(locations);
  }

  @override
  Future<void> saveLocationId(int locationId) {
    if (locationId < 1) {
      throw ArgumentError.value(
        locationId,
        'locationId',
        'must be a positive integer',
      );
    }

    return _preferences.setInt(_locationIdKey, locationId);
  }

  @override
  Future<void> saveRecentLocation(LocationResult location) async {
    final locationId = location.locationId;
    if (locationId == null || locationId < 1) {
      throw ArgumentError.value(
        locationId,
        'location.locationId',
        'must be a positive integer',
      );
    }

    final existing = await readRecentLocations();
    final recent = [
      location,
      ...existing.where((item) => item.locationId != locationId),
    ].take(_maximumRecentLocations);
    await _preferences.setStringList(
      _recentLocationsKey,
      recent.map((item) => jsonEncode(item.toJson())).toList(growable: false),
    );
  }

  @override
  Future<void> clearLocationId() {
    return _preferences.remove(_locationIdKey);
  }
}
