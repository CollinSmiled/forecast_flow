import 'package:shared_preferences/shared_preferences.dart';

abstract interface class SelectedLocationStore {
  Future<int?> readLocationId();

  Future<void> saveLocationId(int locationId);

  Future<void> clearLocationId();
}

class SharedPreferencesSelectedLocationStore implements SelectedLocationStore {
  SharedPreferencesSelectedLocationStore({SharedPreferencesAsync? preferences})
    : _preferences = preferences ?? SharedPreferencesAsync();

  static const _locationIdKey = 'selected_location_id';

  final SharedPreferencesAsync _preferences;

  @override
  Future<int?> readLocationId() async {
    final locationId = await _preferences.getInt(_locationIdKey);
    return locationId != null && locationId > 0 ? locationId : null;
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
  Future<void> clearLocationId() {
    return _preferences.remove(_locationIdKey);
  }
}
