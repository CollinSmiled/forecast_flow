import 'package:forecast_flow_mobile/features/location/data/selected_location_store.dart';

class MemorySelectedLocationStore implements SelectedLocationStore {
  MemorySelectedLocationStore({this.locationId});

  int? locationId;
  int readCount = 0;
  int saveCount = 0;
  int clearCount = 0;

  @override
  Future<int?> readLocationId() async {
    readCount++;
    return locationId;
  }

  @override
  Future<void> saveLocationId(int locationId) async {
    saveCount++;
    this.locationId = locationId;
  }

  @override
  Future<void> clearLocationId() async {
    clearCount++;
    locationId = null;
  }
}
