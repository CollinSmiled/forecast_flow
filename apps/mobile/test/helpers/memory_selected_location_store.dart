import 'package:forecast_flow_mobile/features/location/data/selected_location_store.dart';
import 'package:forecast_flow_mobile/features/location/data/models/location_result.dart';

class MemorySelectedLocationStore implements SelectedLocationStore {
  MemorySelectedLocationStore({this.locationId, List<LocationResult>? recent})
    : recentLocations = [...?recent];

  int? locationId;
  final List<LocationResult> recentLocations;
  int readCount = 0;
  int recentReadCount = 0;
  int saveCount = 0;
  int recentSaveCount = 0;
  int clearCount = 0;

  @override
  Future<int?> readLocationId() async {
    readCount++;
    return locationId;
  }

  @override
  Future<List<LocationResult>> readRecentLocations() async {
    recentReadCount++;
    return List.unmodifiable(recentLocations);
  }

  @override
  Future<void> saveLocationId(int locationId) async {
    saveCount++;
    this.locationId = locationId;
  }

  @override
  Future<void> saveRecentLocation(LocationResult location) async {
    recentSaveCount++;
    recentLocations.removeWhere(
      (item) => item.locationId == location.locationId,
    );
    recentLocations.insert(0, location);
    if (recentLocations.length > 5) {
      recentLocations.removeRange(5, recentLocations.length);
    }
  }

  @override
  Future<void> clearLocationId() async {
    clearCount++;
    locationId = null;
  }
}
