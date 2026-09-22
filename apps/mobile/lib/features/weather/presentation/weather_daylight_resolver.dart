import '../../../core/time/app_time.dart';
import '../data/models/latest_forecast.dart';

abstract final class WeatherDaylightResolver {
  static bool resolve(LatestForecast forecast, {DateTime? now}) {
    final instant = (now ?? DateTime.now()).toUtc();
    final cityNow = AppTime.tryAtLocation(instant, forecast.timezone);
    if (cityNow == null) {
      return forecast.current.weather.isDay ?? true;
    }

    final cityDate = _dateKey(cityNow);
    for (final daily in forecast.daily) {
      if (daily.date != cityDate) {
        continue;
      }

      final sunrise = daily.sunrise;
      final sunset = daily.sunset;
      if (sunrise != null && sunset != null) {
        return !instant.isBefore(sunrise.toUtc()) &&
            instant.isBefore(sunset.toUtc());
      }
      break;
    }

    return cityNow.hour >= 6 && cityNow.hour < 18;
  }

  static String _dateKey(DateTime value) {
    final month = value.month.toString().padLeft(2, '0');
    final day = value.day.toString().padLeft(2, '0');
    return '${value.year}-$month-$day';
  }
}
