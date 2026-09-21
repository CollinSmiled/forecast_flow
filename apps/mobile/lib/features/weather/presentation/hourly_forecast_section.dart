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
    this.now,
    super.key,
  });

  final List<HourlyForecast> forecasts;
  final DateTime currentValidAt;
  final String timezone;
  final DateTime? now;

  @override
  Widget build(BuildContext context) {
    final instant = (now ?? DateTime.now()).toUtc();
    final forecastCurrent = currentValidAt.toUtc();
    final cutoff = instant.isAfter(forecastCurrent) ? instant : forecastCurrent;
    final upcoming = forecasts
        .where((forecast) => !forecast.validAt.toUtc().isBefore(cutoff))
        .take(24)
        .toList(growable: false);
    final textScale = MediaQuery.textScalerOf(context).scale(1);
    final accessibilityGrowth = (textScale - 1).clamp(0.0, 1.0);
    final cardHeight = 152 + (80 * accessibilityGrowth);

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
          if (upcoming.isEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20, bottom: 4),
              child: Text(
                'No upcoming hourly forecast is available yet.',
                key: const ValueKey('hourly-forecast-empty'),
                style: Theme.of(context).textTheme.bodyMedium,
              ),
            )
          else
            SizedBox(
              height: cardHeight,
              child: ListView.separated(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.only(right: 20),
                itemCount: upcoming.length,
                separatorBuilder: (_, _) => const SizedBox(width: 10),
                itemBuilder: (context, index) {
                  return _HourlyForecastCard(
                    forecast: upcoming[index],
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
    final textScale = MediaQuery.textScalerOf(context).scale(1);
    final accessibilityGrowth = (textScale - 1).clamp(0.0, 1.0);
    final cardWidth = 78 + (50 * accessibilityGrowth);

    return Container(
      width: cardWidth,
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
