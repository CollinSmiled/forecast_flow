import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../data/models/latest_forecast.dart';
import '../domain/weather_condition.dart';
import 'daily_forecast_section.dart';
import 'weather_asset_resolver.dart';
import 'weather_controller.dart';

class WeatherHomeScreen extends StatelessWidget {
  const WeatherHomeScreen({
    required this.controller,
    required this.onChooseLocation,
    super.key,
  });

  final WeatherController controller;
  final VoidCallback onChooseLocation;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: DecoratedBox(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [AppColors.skyTop, AppColors.skyBottom],
          ),
        ),
        child: SafeArea(
          child: ListenableBuilder(
            listenable: controller,
            builder: (context, _) {
              return switch (controller.state) {
                WeatherInitial() || WeatherLoading() => _WeatherLoading(
                  onChooseLocation: onChooseLocation,
                ),
                WeatherLoaded(:final forecast) => _WeatherContent(
                  forecast: forecast,
                  onChooseLocation: onChooseLocation,
                ),
                WeatherNotFound(:final message) => _WeatherProblem(
                  title: 'Forecast unavailable',
                  message: message,
                  onRetry: controller.retry,
                  onChooseLocation: onChooseLocation,
                ),
                WeatherFailure(:final message, :final canRetry) =>
                  _WeatherProblem(
                    title: 'Could not load weather',
                    message: message,
                    onRetry: canRetry ? controller.retry : null,
                    onChooseLocation: onChooseLocation,
                  ),
              };
            },
          ),
        ),
      ),
    );
  }
}

class _WeatherContent extends StatelessWidget {
  const _WeatherContent({
    required this.forecast,
    required this.onChooseLocation,
  });

  final LatestForecast forecast;
  final VoidCallback onChooseLocation;

  @override
  Widget build(BuildContext context) {
    final weather = forecast.current.weather;
    final condition = weatherConditionFromWmoCode(weather.weatherCode ?? -1);
    final isDay = weather.isDay ?? true;

    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(24, 20, 24, 28),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _LocationHeader(
            city: forecast.location.city,
            country: forecast.location.country,
            onChooseLocation: onChooseLocation,
          ),
          const SizedBox(height: 36),
          Center(
            child: Image.asset(
              WeatherAssetResolver.resolve(condition, isDay: isDay),
              width: 120,
              height: 120,
              semanticLabel: condition.label,
            ),
          ),
          const SizedBox(height: 18),
          Center(
            child: Text(
              _temperature(weather.temperature2M),
              style: Theme.of(context).textTheme.displayLarge,
            ),
          ),
          const SizedBox(height: 8),
          Center(
            child: Text(
              condition.label,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          const SizedBox(height: 4),
          Center(
            child: Text(
              'Feels like ${_temperature(weather.apparentTemperature)}',
              style: const TextStyle(color: Colors.white70, fontSize: 15),
            ),
          ),
          const SizedBox(height: 32),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: AppColors.card,
              borderRadius: BorderRadius.circular(28),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Current conditions',
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 20),
                Row(
                  children: [
                    Expanded(
                      child: _Metric(
                        icon: Icons.water_drop_outlined,
                        label: 'Humidity',
                        value: _wholeNumber(
                          weather.relativeHumidity2M,
                          suffix: '%',
                        ),
                      ),
                    ),
                    Expanded(
                      child: _Metric(
                        icon: Icons.air_rounded,
                        label: 'Wind',
                        value: _decimal(weather.windSpeed10M, suffix: ' km/h'),
                      ),
                    ),
                  ],
                ),
                const Divider(height: 32),
                Row(
                  children: [
                    Expanded(
                      child: _Metric(
                        icon: Icons.umbrella_outlined,
                        label: 'Rain chance',
                        value: _wholeNumber(
                          weather.precipitationProbability,
                          suffix: '%',
                        ),
                      ),
                    ),
                    Expanded(
                      child: _Metric(
                        icon: Icons.wb_sunny_outlined,
                        label: 'UV index',
                        value: _decimal(weather.uvIndex),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          if (forecast.daily.isNotEmpty) ...[
            const SizedBox(height: 20),
            DailyForecastSection(forecasts: forecast.daily),
          ],
        ],
      ),
    );
  }

  String _temperature(double? value) {
    return value == null ? '--°' : '${value.round()}°';
  }

  String _wholeNumber(double? value, {String suffix = ''}) {
    return value == null ? '--' : '${value.round()}$suffix';
  }

  String _decimal(double? value, {String suffix = ''}) {
    return value == null ? '--' : '${value.toStringAsFixed(1)}$suffix';
  }
}

class _LocationHeader extends StatelessWidget {
  const _LocationHeader({
    required this.city,
    required this.country,
    required this.onChooseLocation,
  });

  final String city;
  final String country;
  final VoidCallback onChooseLocation;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        const Icon(Icons.location_on_rounded, color: Colors.white),
        const SizedBox(width: 6),
        Expanded(
          child: Text(
            '$city, $country',
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 16,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
        IconButton(
          key: const ValueKey('choose-location'),
          onPressed: onChooseLocation,
          tooltip: 'Choose another city',
          color: Colors.white,
          icon: const Icon(Icons.search_rounded, size: 28),
        ),
      ],
    );
  }
}

class _Metric extends StatelessWidget {
  const _Metric({required this.icon, required this.label, required this.value});

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

class _WeatherLoading extends StatelessWidget {
  const _WeatherLoading({required this.onChooseLocation});

  final VoidCallback onChooseLocation;

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        const Center(child: CircularProgressIndicator(color: Colors.white)),
        Positioned(
          top: 12,
          right: 16,
          child: IconButton(
            onPressed: onChooseLocation,
            tooltip: 'Choose another city',
            color: Colors.white,
            icon: const Icon(Icons.search_rounded),
          ),
        ),
      ],
    );
  }
}

class _WeatherProblem extends StatelessWidget {
  const _WeatherProblem({
    required this.title,
    required this.message,
    required this.onRetry,
    required this.onChooseLocation,
  });

  final String title;
  final String message;
  final VoidCallback? onRetry;
  final VoidCallback onChooseLocation;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: AppColors.card,
            borderRadius: BorderRadius.circular(28),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.cloud_off_rounded,
                size: 52,
                color: Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(height: 16),
              Text(title, style: Theme.of(context).textTheme.headlineSmall),
              const SizedBox(height: 8),
              Text(
                message,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 20),
              if (onRetry != null)
                FilledButton.icon(
                  onPressed: onRetry,
                  icon: const Icon(Icons.refresh_rounded),
                  label: const Text('Try again'),
                ),
              TextButton(
                onPressed: onChooseLocation,
                child: const Text('Choose another city'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
