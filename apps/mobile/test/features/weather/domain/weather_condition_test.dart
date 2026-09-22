import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/domain/weather_condition.dart';

void main() {
  test('maps the supported WMO codes to app conditions', () {
    const examples = {
      0: WeatherCondition.clear,
      1: WeatherCondition.mainlyClear,
      2: WeatherCondition.partlyCloudy,
      3: WeatherCondition.overcast,
      45: WeatherCondition.fog,
      48: WeatherCondition.fog,
      53: WeatherCondition.drizzle,
      57: WeatherCondition.freezingDrizzle,
      63: WeatherCondition.rain,
      67: WeatherCondition.freezingRain,
      73: WeatherCondition.snow,
      77: WeatherCondition.snowGrains,
      81: WeatherCondition.rainShowers,
      86: WeatherCondition.snowShowers,
      95: WeatherCondition.thunderstorm,
      99: WeatherCondition.thunderstormWithHail,
    };

    for (final MapEntry(key: code, value: condition) in examples.entries) {
      expect(weatherConditionFromWmoCode(code), condition);
    }
  });

  test('maps unsupported codes to unknown', () {
    expect(weatherConditionFromWmoCode(-1), WeatherCondition.unknown);
    expect(weatherConditionFromWmoCode(100), WeatherCondition.unknown);
  });
}
