import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/presentation/weather_scene_resolver.dart';

void main() {
  test('selects the daytime scene', () {
    expect(
      WeatherSceneResolver.resolve(isDay: true),
      WeatherSceneResolver.dayAsset,
    );
  });

  test('selects the nighttime scene', () {
    expect(
      WeatherSceneResolver.resolve(isDay: false),
      WeatherSceneResolver.nightAsset,
    );
  });
}
