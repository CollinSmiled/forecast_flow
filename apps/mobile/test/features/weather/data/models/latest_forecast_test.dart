import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/weather/data/models/latest_forecast.dart';

void main() {
  test('parses the latest forecast API envelope', () {
    final forecast = LatestForecast.fromEnvelope(_forecastEnvelope());

    expect(forecast.forecastId, 'forecast-123');
    expect(forecast.location.city, 'Jakarta');
    expect(forecast.location.population, 10560000);
    expect(forecast.retrievedAt, DateTime.utc(2026, 9, 20, 0));
    expect(forecast.current.weather.temperature2M, 31.1);
    expect(forecast.current.weather.isDay, isTrue);
    expect(forecast.hourly, hasLength(1));
    expect(forecast.hourly.single.weather.precipitationProbability, 70);
    expect(forecast.daily, hasLength(1));
    expect(forecast.daily.single.date, '2026-09-20');
    expect(forecast.daily.single.sunrise, DateTime.utc(2026, 9, 19, 22, 42));
  });

  test('preserves unavailable weather values as null', () {
    final envelope = _forecastEnvelope();
    final data = envelope['data']! as Map<String, dynamic>;
    final current = data['current']! as Map<String, dynamic>;
    current['visibility'] = null;
    current['uv_index'] = null;

    final forecast = LatestForecast.fromEnvelope(envelope);

    expect(forecast.current.weather.visibility, isNull);
    expect(forecast.current.weather.uvIndex, isNull);
  });

  test('rejects a malformed required field', () {
    final envelope = _forecastEnvelope();
    final data = envelope['data']! as Map<String, dynamic>;
    data['hourly'] = 'not-an-array';

    expect(
      () => LatestForecast.fromEnvelope(envelope),
      throwsA(isA<FormatException>()),
    );
  });
}

Map<String, dynamic> _forecastEnvelope() {
  final weather = <String, dynamic>{
    'temperature_2m': 31.1,
    'apparent_temperature': 35,
    'relative_humidity_2m': 78,
    'precipitation': 1.2,
    'precipitation_probability': 70,
    'rain': 1.2,
    'showers': 0,
    'snowfall': 0,
    'weather_code': 61,
    'cloud_cover': 82,
    'pressure_msl': 1008,
    'visibility': 12000,
    'wind_speed_10m': 14.2,
    'wind_direction_10m': 220,
    'wind_gusts_10m': 28,
    'uv_index': 7.2,
    'is_day': true,
  };

  return {
    'data': {
      'forecast_id': 'forecast-123',
      'location': {
        'location_id': 3,
        'open_meteo_location_id': 1642911,
        'city': 'Jakarta',
        'country': 'Indonesia',
        'country_code': 'ID',
        'latitude': -6.2146,
        'longitude': 106.8451,
        'timezone': 'Asia/Jakarta',
        'elevation': 8,
        'population': 10560000,
        'administrative_area': 'Jakarta',
      },
      'source': 'open_meteo_best_match',
      'retrieved_at': '2026-09-20T00:00:00Z',
      'timezone': 'Asia/Jakarta',
      'units': {
        'temperature': 'celsius',
        'precipitation': 'mm',
        'wind_speed': 'km/h',
        'pressure': 'hPa',
        'visibility': 'm',
      },
      'current': {
        'valid_at': '2026-09-20T01:00:00Z',
        'interval_seconds': 900,
        ...weather,
      },
      'hourly': [
        {'valid_at': '2026-09-20T02:00:00Z', ...weather},
      ],
      'daily': [
        {
          'date': '2026-09-20',
          'weather_code': 61,
          'temperature_2m_max': 32,
          'temperature_2m_min': 25,
          'apparent_temperature_max': 36,
          'apparent_temperature_min': 27,
          'precipitation_sum': 8.4,
          'precipitation_probability_max': 80,
          'precipitation_hours': 4,
          'wind_speed_10m_max': 18,
          'wind_gusts_10m_max': 31,
          'wind_direction_10m_dominant': 220,
          'sunrise': '2026-09-19T22:42:00Z',
          'sunset': '2026-09-20T10:48:00Z',
          'daylight_duration_seconds': 43560,
          'uv_index_max': 9.1,
        },
      ],
    },
  };
}
