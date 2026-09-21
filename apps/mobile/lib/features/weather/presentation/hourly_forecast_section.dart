import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/time/app_time.dart';
import '../data/models/latest_forecast.dart';
import '../domain/weather_condition.dart';
import 'weather_asset_resolver.dart';

class HourlyForecastSection extends StatelessWidget {
  const HourlyForecastSection({
    required this.forecasts,
    required this.currentValidAt,
    required this.timezone,
    super.key,
  });

  final List<HourlyForecast> forecasts;
  final DateTime currentValidAt;
  final String timezone;

  @override
  Widget build(BuildContext context) {
    final upcoming = forecasts
        .where((forecast) => !forecast.validAt.isBefore(currentValidAt))
        .take(24)
        .toList(growable: false);
    final visible = upcoming.isEmpty
        ? forecasts.take(24).toList(growable: false)
        : upcoming;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(20, 20, 0, 20),
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
            'Hourly forecast',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 16),
          SizedBox(
            height: 152,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.only(right: 20),
              itemCount: visible.length,
              separatorBuilder: (_, _) => const SizedBox(width: 10),
              itemBuilder: (context, index) {
                return _HourlyForecastCard(
                  forecast: visible[index],
                  timezone: timezone,
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _HourlyForecastCard extends StatelessWidget {
  const _HourlyForecastCard({required this.forecast, required this.timezone});

  final HourlyForecast forecast;
  final String timezone;

  @override
  Widget build(BuildContext context) {
    final weather = forecast.weather;
    final condition = weatherConditionFromWmoCode(weather.weatherCode ?? -1);

    return Container(
      width: 78,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.56),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.cardBorder),
      ),
      child: Column(
        children: [
          Text(
            _formatHour(forecast.validAt, timezone),
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: 5),
          Image.asset(
            WeatherAssetResolver.resolve(
              condition,
              isDay: weather.isDay ?? true,
            ),
            width: 40,
            height: 40,
            semanticLabel: condition.label,
          ),
          const SizedBox(height: 4),
          Text(
            _temperature(weather.temperature2M),
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const Spacer(),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(
                Icons.water_drop_outlined,
                size: 13,
                color: AppColors.primary,
              ),
              const SizedBox(width: 2),
              Text(
                _percentage(weather.precipitationProbability),
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ),
        ],
      ),
    );
  }

  String _formatHour(DateTime instant, String locationName) {
    final local = AppTime.tryAtLocation(instant, locationName);
    if (local == null) {
      return '--';
    }

    final period = local.hour < 12 ? 'AM' : 'PM';
    final hour = local.hour % 12 == 0 ? 12 : local.hour % 12;
    return '$hour $period';
  }

  String _temperature(double? value) {
    return value == null ? '--°' : '${value.round()}°';
  }

  String _percentage(double? value) {
    return value == null ? '--' : '${value.round()}%';
  }
}
