import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/location/data/models/location_result.dart';

void main() {
  test('round-trips stable location metadata through JSON', () {
    const location = LocationResult(
      locationId: 4,
      openMeteoLocationId: 1850147,
      city: 'Tokyo',
      country: 'Japan',
      countryCode: 'JP',
      latitude: 35.6762,
      longitude: 139.6503,
      timezone: 'Asia/Tokyo',
      elevation: 40,
      population: 14094034,
      administrativeArea: 'Tokyo',
    );

    final restored = LocationResult.fromJson(location.toJson());

    expect(restored.locationId, location.locationId);
    expect(restored.openMeteoLocationId, location.openMeteoLocationId);
    expect(restored.city, location.city);
    expect(restored.country, location.country);
    expect(restored.countryCode, location.countryCode);
    expect(restored.latitude, location.latitude);
    expect(restored.longitude, location.longitude);
    expect(restored.timezone, location.timezone);
    expect(restored.elevation, location.elevation);
    expect(restored.population, location.population);
    expect(restored.administrativeArea, location.administrativeArea);
  });
}
