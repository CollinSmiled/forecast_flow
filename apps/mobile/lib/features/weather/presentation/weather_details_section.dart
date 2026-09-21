import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../data/models/latest_forecast.dart';
import 'current_conditions_view_data.dart';

class WeatherDetailsSection extends StatelessWidget {
  const WeatherDetailsSection({
    required this.conditions,
    required this.units,
    super.key,
  });

  final CurrentConditionsViewData conditions;
  final ForecastUnits units;

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
            'Weather details',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 18),
          Row(
            children: [
              Expanded(
                child: _Detail(
                  icon: Icons.cloud_outlined,
                  label: 'Cloud cover',
                  value: _percentage(conditions.cloudCover),
                ),
              ),
              Expanded(
                child: _Detail(
                  icon: Icons.speed_rounded,
                  label: 'Pressure',
                  value: _wholeNumber(
                    conditions.pressureMsl,
                    suffix: ' ${units.pressure}',
                  ),
                ),
              ),
            ],
          ),
          const Divider(height: 32),
          Row(
            children: [
              Expanded(
                child: _Detail(
                  icon: Icons.visibility_outlined,
                  label: 'Visibility',
                  value: _visibility(conditions.visibility, units.visibility),
                ),
              ),
              Expanded(
                child: _Detail(
                  icon: Icons.air_rounded,
                  label: 'Wind gusts',
                  value: _decimal(
                    conditions.windGusts10M,
                    suffix: ' ${units.windSpeed}',
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  String _percentage(double? value) {
    return value == null ? '--' : '${value.round()}%';
  }

  String _wholeNumber(double? value, {String suffix = ''}) {
    return value == null ? '--' : '${value.round()}$suffix';
  }

  String _decimal(double? value, {String suffix = ''}) {
    return value == null ? '--' : '${value.toStringAsFixed(1)}$suffix';
  }

  String _visibility(double? value, String unit) {
    if (value == null) {
      return '--';
    }
    if (unit == 'm') {
      final kilometers = value / 1000;
      final formatted = kilometers == kilometers.roundToDouble()
          ? kilometers.toStringAsFixed(0)
          : kilometers.toStringAsFixed(1);
      return '$formatted km';
    }

    return '${value.toStringAsFixed(1)} $unit';
  }
}

class _Detail extends StatelessWidget {
  const _Detail({required this.icon, required this.label, required this.value});

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(icon, color: Theme.of(context).colorScheme.primary),
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
