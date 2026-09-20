import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/hourly_forecast_section.dart';

void main() {
  testWidgets('renders upcoming hours in the forecast city timezone', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: HourlyForecastSection(
            currentValidAt: DateTime.utc(2026, 9, 20, 1),
            timezone: 'Asia/Tokyo',
            forecasts: [
              _hour(DateTime.utc(2026, 9, 20), temperature: 20),
              _hour(DateTime.utc(2026, 9, 20, 1), temperature: 21),
              _hour(DateTime.utc(2026, 9, 20, 2), temperature: 22),
            ],
          ),
        ),
      ),
    );

    expect(find.text('Hourly forecast'), findsOneWidget);
    expect(find.text('9 AM'), findsNothing);
    expect(find.text('10 AM'), findsOneWidget);
    expect(find.text('11 AM'), findsOneWidget);
    expect(find.text('20°'), findsNothing);
    expect(find.text('21°'), findsOneWidget);
    expect(find.text('40%'), findsNWidgets(2));
  });

  testWidgets('uses placeholders for invalid timezone and missing values', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: HourlyForecastSection(
            currentValidAt: DateTime.utc(2026, 9, 20),
            timezone: 'Invalid/Timezone',
            forecasts: [
              HourlyForecast(
                validAt: DateTime.utc(2026, 9, 20),
                weather: const WeatherMetrics(),
              ),
            ],
          ),
        ),
      ),
    );

    expect(find.text('--°'), findsOneWidget);
    expect(find.text('--'), findsNWidgets(2));
  });
}

HourlyForecast _hour(DateTime validAt, {required double temperature}) {
  return HourlyForecast(
    validAt: validAt,
    weather: WeatherMetrics(
      temperature2M: temperature,
      precipitationProbability: 40,
      weatherCode: 1,
      isDay: true,
    ),
  );
}
