import '../data/models/latest_forecast.dart';

class CurrentConditionsViewData {
  const CurrentConditionsViewData({required this.current, this.nearestHourly});

  final WeatherMetrics current;
  final WeatherMetrics? nearestHourly;

  factory CurrentConditionsViewData.fromForecast(LatestForecast forecast) {
    return CurrentConditionsViewData(
      current: forecast.current.weather,
      nearestHourly: _nearestHourlyMetrics(forecast),
    );
  }

  double? get precipitationProbability =>
      current.precipitationProbability ??
      nearestHourly?.precipitationProbability;

  double? get uvIndex => current.uvIndex ?? nearestHourly?.uvIndex;

  double? get visibility => current.visibility ?? nearestHourly?.visibility;

  double? get cloudCover => current.cloudCover ?? nearestHourly?.cloudCover;

  double? get pressureMsl => current.pressureMsl ?? nearestHourly?.pressureMsl;

  double? get windGusts10M =>
      current.windGusts10M ?? nearestHourly?.windGusts10M;
}

WeatherMetrics? _nearestHourlyMetrics(LatestForecast forecast) {
  HourlyForecast? nearest;
  int? nearestDistanceSeconds;

  for (final hourly in forecast.hourly) {
    final distanceSeconds = hourly.validAt
        .difference(forecast.current.validAt)
        .inSeconds
        .abs();
    if (nearestDistanceSeconds == null ||
        distanceSeconds < nearestDistanceSeconds) {
      nearest = hourly;
      nearestDistanceSeconds = distanceSeconds;
    }
  }

  return nearest?.weather;
}
