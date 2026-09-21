import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../data/models/latest_forecast.dart';
import '../domain/weather_condition.dart';
import 'weather_asset_resolver.dart';

class DailyForecastSection extends StatelessWidget {
  const DailyForecastSection({required this.forecasts, super.key});

  final List<DailyForecast> forecasts;

  @override
  Widget build(BuildContext context) {
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
            '${forecasts.length}-day forecast',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 12),
          for (final (index, forecast) in forecasts.indexed) ...[
            _DailyForecastRow(
              forecast: forecast,
              dayLabel: index == 0 ? 'Today' : _weekday(forecast.date),
            ),
            if (index < forecasts.length - 1) const Divider(height: 1),
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
  final parts = date.split('-');
  if (parts.length != 3) {
    return date;
  }

  final year = int.tryParse(parts[0]);
  final month = int.tryParse(parts[1]);
  final day = int.tryParse(parts[2]);
  if (year == null || month == null || day == null) {
    return date;
  }

  final weekday = DateTime.utc(year, month, day).weekday;
  return const ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'][weekday - 1];
}
