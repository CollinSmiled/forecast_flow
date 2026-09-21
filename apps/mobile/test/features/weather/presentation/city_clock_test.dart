import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/city_clock.dart';

void main() {
  test('formats the same instant in each forecast city timezone', () {
    final instant = DateTime.utc(2026, 9, 20, 13, 35);

    expect(CityClock.label('Asia/Jakarta', now: instant), 'Local time 8:35 PM');
    expect(CityClock.label('Asia/Tokyo', now: instant), 'Local time 10:35 PM');
  });

  test('returns null when the forecast timezone is unknown', () {
    expect(
      CityClock.label('Invalid/Timezone', now: DateTime.utc(2026, 9, 20)),
      isNull,
    );
  });
}
