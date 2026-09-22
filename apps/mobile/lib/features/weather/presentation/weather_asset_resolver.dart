import '../domain/weather_condition.dart';

abstract final class WeatherAssetResolver {
  static const _root = 'assets/icons/weather';

  static String resolve(WeatherCondition condition, {required bool isDay}) {
    return switch (condition) {
      WeatherCondition.clear =>
        isDay ? '$_root/clear_day.png' : '$_root/clear_night.png',
      WeatherCondition.mainlyClear =>
        isDay
            ? '$_root/mainly_clear_day.png'
            : '$_root/partly_cloudy_night.png',
      WeatherCondition.partlyCloudy =>
        isDay
            ? '$_root/partly_cloudy_day.png'
            : '$_root/partly_cloudy_night.png',
      WeatherCondition.overcast => '$_root/overcast.png',
      WeatherCondition.fog =>
        isDay ? '$_root/fog_day.png' : '$_root/fog_night.png',
      WeatherCondition.drizzle => '$_root/drizzle.png',
      WeatherCondition.freezingDrizzle => '$_root/mixed_precipitation.png',
      WeatherCondition.rain => '$_root/rain.png',
      WeatherCondition.freezingRain => '$_root/mixed_precipitation.png',
      WeatherCondition.snow => '$_root/snow.png',
      WeatherCondition.snowGrains => '$_root/snow_grains.png',
      WeatherCondition.rainShowers =>
        isDay ? '$_root/showers_day.png' : '$_root/rain.png',
      WeatherCondition.snowShowers => '$_root/snow.png',
      WeatherCondition.thunderstorm =>
        isDay ? '$_root/thunderstorm_day.png' : '$_root/thunderstorm_night.png',
      WeatherCondition.thunderstormWithHail => '$_root/thunderstorm.png',
      WeatherCondition.unknown => '$_root/overcast.png',
    };
  }
}
