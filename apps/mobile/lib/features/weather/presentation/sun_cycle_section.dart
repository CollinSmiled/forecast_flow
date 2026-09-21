import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/time/app_time.dart';
import '../data/models/latest_forecast.dart';

class SunCycleSection extends StatelessWidget {
  const SunCycleSection({
    required this.forecast,
    required this.timezone,
    super.key,
  });

  final DailyForecast forecast;
  final String timezone;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
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
            'Sun and daylight',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 18),
          Row(
            children: [
              Expanded(
                child: _SunTime(
                  asset: 'assets/icons/weather/sunrise.png',
                  label: 'Sunrise',
                  value: _formatTime(forecast.sunrise, timezone),
                ),
              ),
              Expanded(
                child: _SunTime(
                  asset: 'assets/icons/weather/sunset.png',
                  label: 'Sunset',
                  value: _formatTime(forecast.sunset, timezone),
                ),
              ),
            ],
          ),
          const Divider(height: 32),
          Row(
            children: [
              Icon(
                Icons.light_mode_outlined,
                color: Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  'Daylight',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ),
              const SizedBox(width: 8),
              Flexible(
                child: Text(
                  _formatDuration(forecast.daylightDurationSeconds),
                  textAlign: TextAlign.end,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  String _formatTime(DateTime? instant, String locationName) {
    if (instant == null) {
      return '--';
    }

    final local = AppTime.tryAtLocation(instant, locationName);
    if (local == null) {
      return '--';
    }

    final period = local.hour < 12 ? 'AM' : 'PM';
    final hour = local.hour % 12 == 0 ? 12 : local.hour % 12;
    final minute = local.minute.toString().padLeft(2, '0');
    return '$hour:$minute $period';
  }

  String _formatDuration(double? seconds) {
    if (seconds == null) {
      return '--';
    }

    final minutes = (seconds / 60).round();
    final hoursPart = minutes ~/ 60;
    final minutesPart = minutes % 60;
    return '${hoursPart}h ${minutesPart}m';
  }
}

class _SunTime extends StatelessWidget {
  const _SunTime({
    required this.asset,
    required this.label,
    required this.value,
  });

  final String asset;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Image.asset(asset, width: 42, height: 42),
        const SizedBox(width: 10),
        Flexible(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: Theme.of(context).textTheme.bodyMedium),
              const SizedBox(height: 2),
              Text(value, style: Theme.of(context).textTheme.titleMedium),
            ],
          ),
        ),
      ],
    );
  }
}
