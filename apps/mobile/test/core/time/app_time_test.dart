import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/core/time/app_time.dart';

void main() {
  test('converts an instant into the selected IANA timezone', () {
    final instant = DateTime.utc(2026, 9, 20);

    final tokyo = AppTime.tryAtLocation(instant, 'Asia/Tokyo');
    final jakarta = AppTime.tryAtLocation(instant, 'Asia/Jakarta');

    expect(tokyo?.hour, 9);
    expect(jakarta?.hour, 7);
  });

  test('returns null for an unknown timezone', () {
    final converted = AppTime.tryAtLocation(
      DateTime.utc(2026, 9, 20),
      'Asia/Not_A_Real_City',
    );

    expect(converted, isNull);
  });
}
