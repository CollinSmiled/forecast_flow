import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/current_conditions_view_data.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_details_section.dart';

void main() {
  testWidgets('renders atmospheric details with display units', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: WeatherDetailsSection(
            conditions: CurrentConditionsViewData(
              current: WeatherMetrics(
                cloudCover: 88,
                pressureMsl: 1008,
                windGusts10M: 28,
              ),
              nearestHourly: WeatherMetrics(visibility: 12000),
            ),
            units: ForecastUnits(
              temperature: 'celsius',
              precipitation: 'mm',
              windSpeed: 'km/h',
              pressure: 'hPa',
              visibility: 'm',
            ),
          ),
        ),
      ),
    );

    expect(find.text('88%'), findsOneWidget);
    expect(find.text('1008 hPa'), findsOneWidget);
    expect(find.text('12 km'), findsOneWidget);
    expect(find.text('28.0 km/h'), findsOneWidget);
  });
}
