import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/daily_forecast_section.dart';

void main() {
  testWidgets('renders daily conditions and high-low temperatures', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: SingleChildScrollView(
            child: DailyForecastSection(
              forecasts: const [
                DailyForecast(
                  date: '2026-09-20',
                  weatherCode: 0,
                  temperature2MMax: 32,
                  temperature2MMin: 25,
                  precipitationProbabilityMax: 10,
                ),
                DailyForecast(
                  date: '2026-09-21',
                  weatherCode: 61,
                  temperature2MMax: 30,
                  temperature2MMin: 24,
                  precipitationProbabilityMax: 80,
                ),
              ],
            ),
          ),
        ),
      ),
    );

    expect(find.text('2-day forecast'), findsOneWidget);
    expect(find.text('Today'), findsOneWidget);
    expect(find.text('Mon'), findsOneWidget);
    expect(find.text('32°'), findsOneWidget);
    expect(find.text('25°'), findsOneWidget);
    expect(find.text('80%'), findsOneWidget);
  });

  testWidgets('renders missing daily measurements safely', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: DailyForecastSection(
            forecasts: [DailyForecast(date: '2026-09-20')],
          ),
        ),
      ),
    );

    expect(find.text('--°'), findsNWidgets(2));
    expect(find.text('--'), findsOneWidget);
  });
}
