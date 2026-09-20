import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/sun_cycle_section.dart';

void main() {
  testWidgets('renders sun times in the forecast city timezone', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: SunCycleSection(
            timezone: 'Asia/Jakarta',
            forecast: DailyForecast(
              date: '2026-09-20',
              sunrise: DateTime.utc(2026, 9, 19, 22, 42),
              sunset: DateTime.utc(2026, 9, 20, 10, 48),
              daylightDurationSeconds: 43560,
            ),
          ),
        ),
      ),
    );

    expect(find.text('5:42 AM'), findsOneWidget);
    expect(find.text('5:48 PM'), findsOneWidget);
    expect(find.text('12h 6m'), findsOneWidget);
  });

  testWidgets('renders unavailable sun values safely', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: SunCycleSection(
            timezone: 'Invalid/Timezone',
            forecast: DailyForecast(date: '2026-09-20'),
          ),
        ),
      ),
    );

    expect(find.text('--'), findsNWidgets(3));
  });
}
