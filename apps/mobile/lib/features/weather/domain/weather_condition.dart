enum WeatherCondition {
  clear,
  mainlyClear,
  partlyCloudy,
  overcast,
  fog,
  drizzle,
  freezingDrizzle,
  rain,
  freezingRain,
  snow,
  snowGrains,
  rainShowers,
  snowShowers,
  thunderstorm,
  thunderstormWithHail,
  unknown,
}

WeatherCondition weatherConditionFromWmoCode(int code) {
  return switch (code) {
    0 => WeatherCondition.clear,
    1 => WeatherCondition.mainlyClear,
    2 => WeatherCondition.partlyCloudy,
    3 => WeatherCondition.overcast,
    45 || 48 => WeatherCondition.fog,
    51 || 53 || 55 => WeatherCondition.drizzle,
    56 || 57 => WeatherCondition.freezingDrizzle,
    61 || 63 || 65 => WeatherCondition.rain,
    66 || 67 => WeatherCondition.freezingRain,
    71 || 73 || 75 => WeatherCondition.snow,
    77 => WeatherCondition.snowGrains,
    80 || 81 || 82 => WeatherCondition.rainShowers,
    85 || 86 => WeatherCondition.snowShowers,
    95 => WeatherCondition.thunderstorm,
    96 || 99 => WeatherCondition.thunderstormWithHail,
    _ => WeatherCondition.unknown,
  };
}

extension WeatherConditionLabel on WeatherCondition {
  String get label {
    return switch (this) {
      WeatherCondition.clear => 'Clear sky',
      WeatherCondition.mainlyClear => 'Mainly clear',
      WeatherCondition.partlyCloudy => 'Partly cloudy',
      WeatherCondition.overcast => 'Overcast',
      WeatherCondition.fog => 'Foggy',
      WeatherCondition.drizzle => 'Drizzle',
      WeatherCondition.freezingDrizzle => 'Freezing drizzle',
      WeatherCondition.rain => 'Rain',
      WeatherCondition.freezingRain => 'Freezing rain',
      WeatherCondition.snow => 'Snowfall',
      WeatherCondition.snowGrains => 'Snow grains',
      WeatherCondition.rainShowers => 'Rain showers',
      WeatherCondition.snowShowers => 'Snow showers',
      WeatherCondition.thunderstorm => 'Thunderstorm',
      WeatherCondition.thunderstormWithHail => 'Thunderstorm with hail',
      WeatherCondition.unknown => 'Unknown conditions',
    };
  }
}
