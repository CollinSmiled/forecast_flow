import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/domain/weather_condition.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_asset_resolver.dart';

void main() {
  test('uses different clear-weather artwork for day and night', () {
    final dayAsset = WeatherAssetResolver.resolve(
      WeatherCondition.clear,
      isDay: true,
    );
    final nightAsset = WeatherAssetResolver.resolve(
      WeatherCondition.clear,
      isDay: false,
    );

    expect(dayAsset, endsWith('clear_day.png'));
    expect(nightAsset, endsWith('clear_night.png'));
  });

  test('uses a safe fallback for unknown weather codes', () {
    final asset = WeatherAssetResolver.resolve(
      WeatherCondition.unknown,
      isDay: true,
    );

    expect(asset, endsWith('overcast.png'));
  });
}
