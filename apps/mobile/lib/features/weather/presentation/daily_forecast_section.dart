import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/time/app_time.dart';
import '../data/models/latest_forecast.dart';
import '../domain/weather_condition.dart';
import 'weather_asset_resolver.dart';

class DailyForecastSection extends StatelessWidget {
  const DailyForecastSection({
    required this.forecasts,
    required this.timezone,
    this.now,
    super.key,
  });

  final List<DailyForecast> forecasts;
  final String timezone;
  final DateTime? now;

  @override
  Widget build(BuildContext context) {
    final instant = (now ?? DateTime.now()).toUtc();
    final cityNow = AppTime.tryAtLocation(instant, timezone);
    final cityDate = cityNow == null
        ? null
        : DateTime.utc(cityNow.year, cityNow.month, cityNow.day);
    final upcoming = cityDate == null
        ? forecasts
        : forecasts
              .where((forecast) {
                final date = _parseDate(forecast.date);
                return date == null || !date.isBefore(cityDate);
              })
              .toList(growable: false);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 8),
      decoration: BoxDecoration(
        color: AppColors.card,
        borderRadius: BorderRadius.circular(28),
        border: Border.all(color: AppColors.cardBorder),
        boxShadow: const [
          BoxShadow(
            color: AppColors.cardShadow,
            blurRadius: 18,
            offset: Offset(0, 8),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            upcoming.isEmpty
                ? 'Daily forecast'
                : '${upcoming.length}-day forecast',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 12),
          if (upcoming.isEmpty)
            Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: Text(
                'No upcoming daily forecast is available yet.',
                key: const ValueKey('daily-forecast-empty'),
                style: Theme.of(context).textTheme.bodyMedium,
              ),
            )
          else
            for (final (index, forecast) in upcoming.indexed) ...[
              _DailyForecastRow(
                forecast: forecast,
                dayLabel: _dayLabel(forecast.date, cityDate),
              ),
              if (index < upcoming.length - 1) const Divider(height: 1),
            ],
        ],
      ),
    );
  }
}

class _DailyForecastRow extends StatelessWidget {
  const _DailyForecastRow({required this.forecast, required this.dayLabel});

  final DailyForecast forecast;
  final String dayLabel;

  @override
  Widget build(BuildContext context) {
    final condition = weatherConditionFromWmoCode(forecast.weatherCode ?? -1);

    return SizedBox(
      height: 72,
      child: Row(
        children: [
          SizedBox(
            width: 72,
            child: Text(
              dayLabel,
              style: Theme.of(context).textTheme.titleMedium,
            ),
          ),
          Image.asset(
            WeatherAssetResolver.resolve(condition, isDay: true),
            width: 42,
            height: 42,
            semanticLabel: condition.label,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Row(
              children: [
                const Icon(
                  Icons.water_drop_outlined,
                  size: 15,
                  color: AppColors.primary,
                ),
                const SizedBox(width: 3),
                Text(
                  _percentage(forecast.precipitationProbabilityMax),
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ],
            ),
          ),
          Text(
            _temperature(forecast.temperature2MMax),
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(width: 10),
          Text(
            _temperature(forecast.temperature2MMin),
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ],
      ),
    );
  }

  String _temperature(double? value) {
    return value == null ? '--°' : '${value.round()}°';
  }

  String _percentage(double? value) {
    return value == null ? '--' : '${value.round()}%';
  }
}

String _weekday(String date) {
  final parsed = _parseDate(date);
  if (parsed == null) {
    return date;
  }

  final weekday = parsed.weekday;
  return const ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'][weekday - 1];
}

String _dayLabel(String date, DateTime? cityDate) {
  final parsed = _parseDate(date);
  if (parsed == null || cityDate == null) {
    return _weekday(date);
  }
  if (parsed == cityDate) {
    return 'Today';
  }
  if (parsed == cityDate.add(const Duration(days: 1))) {
    return 'Tomorrow';
  }

  return _weekday(date);
}

DateTime? _parseDate(String date) {
  final parts = date.split('-');
  if (parts.length != 3) {
    return null;
  }

  final year = int.tryParse(parts[0]);
  final month = int.tryParse(parts[1]);
  final day = int.tryParse(parts[2]);
  if (year == null || month == null || day == null) {
    return null;
  }

  final parsed = DateTime.utc(year, month, day);
  if (parsed.year != year || parsed.month != month || parsed.day != day) {
    return null;
  }

  return parsed;
}
