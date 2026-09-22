class LatestForecast {
  const LatestForecast({
    required this.forecastId,
    required this.location,
    required this.source,
    required this.retrievedAt,
    required this.timezone,
    required this.units,
    required this.current,
    required this.hourly,
    required this.daily,
  });

  final String forecastId;
  final ForecastLocation location;
  final String source;
  final DateTime retrievedAt;
  final String timezone;
  final ForecastUnits units;
  final CurrentForecast current;
  final List<HourlyForecast> hourly;
  final List<DailyForecast> daily;

  factory LatestForecast.fromEnvelope(Map<String, dynamic> json) {
    return LatestForecast.fromJson(_requiredMap(json, 'data'));
  }

  factory LatestForecast.fromJson(Map<String, dynamic> json) {
    return LatestForecast(
      forecastId: _requiredString(json, 'forecast_id'),
      location: ForecastLocation.fromJson(_requiredMap(json, 'location')),
      source: _requiredString(json, 'source'),
      retrievedAt: _requiredDateTime(json, 'retrieved_at'),
      timezone: _requiredString(json, 'timezone'),
      units: ForecastUnits.fromJson(_requiredMap(json, 'units')),
      current: CurrentForecast.fromJson(_requiredMap(json, 'current')),
      hourly: _requiredList(json, 'hourly')
          .map((item) => HourlyForecast.fromJson(_asMap(item, 'hourly item')))
          .toList(growable: false),
      daily: _requiredList(json, 'daily')
          .map((item) => DailyForecast.fromJson(_asMap(item, 'daily item')))
          .toList(growable: false),
    );
  }
}

class ForecastLocation {
  const ForecastLocation({
    required this.locationId,
    required this.openMeteoLocationId,
    required this.city,
    required this.country,
    required this.countryCode,
    required this.latitude,
    required this.longitude,
    required this.timezone,
    this.elevation,
    this.population,
    this.administrativeArea,
  });

  final int locationId;
  final int openMeteoLocationId;
  final String city;
  final String country;
  final String countryCode;
  final double latitude;
  final double longitude;
  final String timezone;
  final double? elevation;
  final int? population;
  final String? administrativeArea;

  factory ForecastLocation.fromJson(Map<String, dynamic> json) {
    return ForecastLocation(
      locationId: _requiredInt(json, 'location_id'),
      openMeteoLocationId: _requiredInt(json, 'open_meteo_location_id'),
      city: _requiredString(json, 'city'),
      country: _requiredString(json, 'country'),
      countryCode: _requiredString(json, 'country_code'),
      latitude: _requiredDouble(json, 'latitude'),
      longitude: _requiredDouble(json, 'longitude'),
      timezone: _requiredString(json, 'timezone'),
      elevation: _optionalDouble(json, 'elevation'),
      population: _optionalInt(json, 'population'),
      administrativeArea: _optionalString(json, 'administrative_area'),
    );
  }
}

class ForecastUnits {
  const ForecastUnits({
    required this.temperature,
    required this.precipitation,
    required this.windSpeed,
    required this.pressure,
    required this.visibility,
  });

  final String temperature;
  final String precipitation;
  final String windSpeed;
  final String pressure;
  final String visibility;

  factory ForecastUnits.fromJson(Map<String, dynamic> json) {
    return ForecastUnits(
      temperature: _requiredString(json, 'temperature'),
      precipitation: _requiredString(json, 'precipitation'),
      windSpeed: _requiredString(json, 'wind_speed'),
      pressure: _requiredString(json, 'pressure'),
      visibility: _requiredString(json, 'visibility'),
    );
  }
}

class CurrentForecast {
  const CurrentForecast({
    required this.validAt,
    required this.intervalSeconds,
    required this.weather,
  });

  final DateTime validAt;
  final int intervalSeconds;
  final WeatherMetrics weather;

  factory CurrentForecast.fromJson(Map<String, dynamic> json) {
    return CurrentForecast(
      validAt: _requiredDateTime(json, 'valid_at'),
      intervalSeconds: _requiredInt(json, 'interval_seconds'),
      weather: WeatherMetrics.fromJson(json),
    );
  }
}

class HourlyForecast {
  const HourlyForecast({required this.validAt, required this.weather});

  final DateTime validAt;
  final WeatherMetrics weather;

  factory HourlyForecast.fromJson(Map<String, dynamic> json) {
    return HourlyForecast(
      validAt: _requiredDateTime(json, 'valid_at'),
      weather: WeatherMetrics.fromJson(json),
    );
  }
}

class WeatherMetrics {
  const WeatherMetrics({
    this.temperature2M,
    this.apparentTemperature,
    this.relativeHumidity2M,
    this.precipitation,
    this.precipitationProbability,
    this.rain,
    this.showers,
    this.snowfall,
    this.weatherCode,
    this.cloudCover,
    this.pressureMsl,
    this.visibility,
    this.windSpeed10M,
    this.windDirection10M,
    this.windGusts10M,
    this.uvIndex,
    this.isDay,
  });

  final double? temperature2M;
  final double? apparentTemperature;
  final double? relativeHumidity2M;
  final double? precipitation;
  final double? precipitationProbability;
  final double? rain;
  final double? showers;
  final double? snowfall;
  final int? weatherCode;
  final double? cloudCover;
  final double? pressureMsl;
  final double? visibility;
  final double? windSpeed10M;
  final double? windDirection10M;
  final double? windGusts10M;
  final double? uvIndex;
  final bool? isDay;

  factory WeatherMetrics.fromJson(Map<String, dynamic> json) {
    return WeatherMetrics(
      temperature2M: _optionalDouble(json, 'temperature_2m'),
      apparentTemperature: _optionalDouble(json, 'apparent_temperature'),
      relativeHumidity2M: _optionalDouble(json, 'relative_humidity_2m'),
      precipitation: _optionalDouble(json, 'precipitation'),
      precipitationProbability: _optionalDouble(
        json,
        'precipitation_probability',
      ),
      rain: _optionalDouble(json, 'rain'),
      showers: _optionalDouble(json, 'showers'),
      snowfall: _optionalDouble(json, 'snowfall'),
      weatherCode: _optionalInt(json, 'weather_code'),
      cloudCover: _optionalDouble(json, 'cloud_cover'),
      pressureMsl: _optionalDouble(json, 'pressure_msl'),
      visibility: _optionalDouble(json, 'visibility'),
      windSpeed10M: _optionalDouble(json, 'wind_speed_10m'),
      windDirection10M: _optionalDouble(json, 'wind_direction_10m'),
      windGusts10M: _optionalDouble(json, 'wind_gusts_10m'),
      uvIndex: _optionalDouble(json, 'uv_index'),
      isDay: _optionalBool(json, 'is_day'),
    );
  }
}

class DailyForecast {
  const DailyForecast({
    required this.date,
    this.weatherCode,
    this.temperature2MMax,
    this.temperature2MMin,
    this.apparentTemperatureMax,
    this.apparentTemperatureMin,
    this.precipitationSum,
    this.precipitationProbabilityMax,
    this.precipitationHours,
    this.windSpeed10MMax,
    this.windGusts10MMax,
    this.windDirection10MDominant,
    this.sunrise,
    this.sunset,
    this.daylightDurationSeconds,
    this.uvIndexMax,
  });

  final String date;
  final int? weatherCode;
  final double? temperature2MMax;
  final double? temperature2MMin;
  final double? apparentTemperatureMax;
  final double? apparentTemperatureMin;
  final double? precipitationSum;
  final double? precipitationProbabilityMax;
  final double? precipitationHours;
  final double? windSpeed10MMax;
  final double? windGusts10MMax;
  final double? windDirection10MDominant;
  final DateTime? sunrise;
  final DateTime? sunset;
  final double? daylightDurationSeconds;
  final double? uvIndexMax;

  factory DailyForecast.fromJson(Map<String, dynamic> json) {
    return DailyForecast(
      date: _requiredString(json, 'date'),
      weatherCode: _optionalInt(json, 'weather_code'),
      temperature2MMax: _optionalDouble(json, 'temperature_2m_max'),
      temperature2MMin: _optionalDouble(json, 'temperature_2m_min'),
      apparentTemperatureMax: _optionalDouble(json, 'apparent_temperature_max'),
      apparentTemperatureMin: _optionalDouble(json, 'apparent_temperature_min'),
      precipitationSum: _optionalDouble(json, 'precipitation_sum'),
      precipitationProbabilityMax: _optionalDouble(
        json,
        'precipitation_probability_max',
      ),
      precipitationHours: _optionalDouble(json, 'precipitation_hours'),
      windSpeed10MMax: _optionalDouble(json, 'wind_speed_10m_max'),
      windGusts10MMax: _optionalDouble(json, 'wind_gusts_10m_max'),
      windDirection10MDominant: _optionalDouble(
        json,
        'wind_direction_10m_dominant',
      ),
      sunrise: _optionalDateTime(json, 'sunrise'),
      sunset: _optionalDateTime(json, 'sunset'),
      daylightDurationSeconds: _optionalDouble(
        json,
        'daylight_duration_seconds',
      ),
      uvIndexMax: _optionalDouble(json, 'uv_index_max'),
    );
  }
}

Map<String, dynamic> _requiredMap(Map<String, dynamic> json, String key) {
  return _asMap(json[key], key);
}

Map<String, dynamic> _asMap(Object? value, String field) {
  if (value is Map<String, dynamic>) {
    return value;
  }

  throw FormatException('$field must be a JSON object');
}

List<dynamic> _requiredList(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is List<dynamic>) {
    return value;
  }

  throw FormatException('$key must be a JSON array');
}

String _requiredString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is String && value.isNotEmpty) {
    return value;
  }

  throw FormatException('$key must be a non-empty string');
}

String? _optionalString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null || value is String) {
    return value as String?;
  }

  throw FormatException('$key must be a string or null');
}

int _requiredInt(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is int) {
    return value;
  }

  throw FormatException('$key must be an integer');
}

int? _optionalInt(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null || value is int) {
    return value as int?;
  }

  throw FormatException('$key must be an integer or null');
}

double _requiredDouble(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is num) {
    return value.toDouble();
  }

  throw FormatException('$key must be a number');
}

double? _optionalDouble(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null) {
    return null;
  }
  if (value is num) {
    return value.toDouble();
  }

  throw FormatException('$key must be a number or null');
}

bool? _optionalBool(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null || value is bool) {
    return value as bool?;
  }

  throw FormatException('$key must be a boolean or null');
}

DateTime _requiredDateTime(Map<String, dynamic> json, String key) {
  return DateTime.parse(_requiredString(json, key));
}

DateTime? _optionalDateTime(Map<String, dynamic> json, String key) {
  final value = _optionalString(json, key);
  return value == null ? null : DateTime.parse(value);
}
