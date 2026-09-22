import '../../../core/time/app_time.dart';
import '../data/models/latest_forecast.dart';

enum WeatherScenePeriod { morning, day, evening, night }

abstract final class WeatherSceneResolver {
  static const morningAsset = 'assets/scenes/morning.png';
  static const dayAsset = 'assets/scenes/day.png';
  static const eveningAsset = 'assets/scenes/evening.png';
  static const nightAsset = 'assets/scenes/night.png';

  static String resolve(WeatherScenePeriod period) {
    return switch (period) {
      WeatherScenePeriod.morning => morningAsset,
      WeatherScenePeriod.day => dayAsset,
      WeatherScenePeriod.evening => eveningAsset,
      WeatherScenePeriod.night => nightAsset,
    };
  }

  static String resolveForForecast(LatestForecast forecast, {DateTime? now}) {
    return resolve(periodForForecast(forecast, now: now));
  }

  static WeatherScenePeriod periodForForecast(
    LatestForecast forecast, {
    DateTime? now,
  }) {
    final instant = (now ?? DateTime.now()).toUtc();
    final cityNow = AppTime.tryAtLocation(instant, forecast.timezone);
    if (cityNow == null) {
      return forecast.current.weather.isDay ?? true
          ? WeatherScenePeriod.day
          : WeatherScenePeriod.night;
    }

    final cityDate = _dateKey(cityNow);
    for (final daily in forecast.daily) {
      if (daily.date != cityDate) {
        continue;
      }

      final sunrise = daily.sunrise?.toUtc();
      final sunset = daily.sunset?.toUtc();
      if (sunrise != null && sunset != null) {
        return _periodAroundSun(instant, sunrise: sunrise, sunset: sunset);
      }
      break;
    }

    return periodForLocalTime(cityNow);
  }

  static String resolveForLocalTime(DateTime localTime) {
    return resolve(periodForLocalTime(localTime));
  }

  static WeatherScenePeriod periodForLocalTime(DateTime localTime) {
    return switch (localTime.hour) {
      >= 5 && < 9 => WeatherScenePeriod.morning,
      >= 9 && < 17 => WeatherScenePeriod.day,
      >= 17 && < 20 => WeatherScenePeriod.evening,
      _ => WeatherScenePeriod.night,
    };
  }

  static WeatherScenePeriod _periodAroundSun(
    DateTime instant, {
    required DateTime sunrise,
    required DateTime sunset,
  }) {
    final morningStart = sunrise.subtract(const Duration(minutes: 90));
    final morningEnd = sunrise.add(const Duration(hours: 2));
    final eveningStart = sunset.subtract(const Duration(hours: 2));
    final eveningEnd = sunset.add(const Duration(minutes: 90));

    if (!instant.isBefore(morningStart) && instant.isBefore(morningEnd)) {
      return WeatherScenePeriod.morning;
    }
    if (!instant.isBefore(eveningStart) && instant.isBefore(eveningEnd)) {
      return WeatherScenePeriod.evening;
    }
    if (!instant.isBefore(morningEnd) && instant.isBefore(eveningStart)) {
      return WeatherScenePeriod.day;
    }
    return WeatherScenePeriod.night;
  }

  static String _dateKey(DateTime value) {
    final month = value.month.toString().padLeft(2, '0');
    final day = value.day.toString().padLeft(2, '0');
    return '${value.year}-$month-$day';
  }
}
