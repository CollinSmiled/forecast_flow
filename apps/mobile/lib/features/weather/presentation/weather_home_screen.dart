import 'dart:async';

import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../data/models/latest_forecast.dart';
import '../domain/weather_condition.dart';
import 'current_conditions_view_data.dart';
import 'daily_forecast_section.dart';
import 'forecast_freshness.dart';
import 'hourly_forecast_section.dart';
import 'sun_cycle_section.dart';
import 'weather_asset_resolver.dart';
import 'weather_controller.dart';
import 'weather_daylight_resolver.dart';
import 'weather_details_section.dart';
import 'weather_scene_resolver.dart';

class WeatherHomeScreen extends StatefulWidget {
  const WeatherHomeScreen({
    required this.controller,
    required this.onChooseLocation,
    super.key,
  });

  final WeatherController controller;
  final VoidCallback onChooseLocation;

  @override
  State<WeatherHomeScreen> createState() => _WeatherHomeScreenState();
}

class _WeatherHomeScreenState extends State<WeatherHomeScreen>
    with WidgetsBindingObserver {
  Timer? _sceneClock;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _sceneClock = Timer.periodic(const Duration(minutes: 1), (_) {
      if (mounted) {
        setState(() {});
      }
    });
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && mounted) {
      setState(() {});
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _sceneClock?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: ListenableBuilder(
        listenable: widget.controller,
        builder: (context, _) {
          final state = widget.controller.state;
          final sceneAsset = switch (state) {
            WeatherLoaded(:final forecast) =>
              WeatherSceneResolver.resolveForForecast(forecast),
            _ => null,
          };

          return DecoratedBox(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [AppColors.skyTop, AppColors.skyBottom],
              ),
            ),
            child: Stack(
              fit: StackFit.expand,
              children: [
                if (sceneAsset != null)
                  Image.asset(
                    sceneAsset,
                    key: const ValueKey('weather-scene-background'),
                    fit: BoxFit.cover,
                  ),
                if (sceneAsset != null)
                  const DecoratedBox(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topCenter,
                        end: Alignment.bottomCenter,
                        colors: [Color(0x2605182B), Color(0x5205182B)],
                      ),
                    ),
                  ),
                SafeArea(
                  child: switch (state) {
                    WeatherInitial() || WeatherLoading() => _WeatherLoading(
                      onChooseLocation: widget.onChooseLocation,
                    ),
                    WeatherAwaitingForecast() => _WeatherPreparing(
                      onRetry: widget.controller.retry,
                      onChooseLocation: widget.onChooseLocation,
                    ),
                    WeatherLoaded(
                      :final forecast,
                      :final refreshErrorMessage,
                    ) =>
                      _WeatherContent(
                        forecast: forecast,
                        refreshErrorMessage: refreshErrorMessage,
                        onRefresh: widget.controller.refresh,
                        onChooseLocation: widget.onChooseLocation,
                      ),
                    WeatherNotFound(:final message) => _WeatherProblem(
                      title: 'Forecast unavailable',
                      message: message,
                      onRetry: widget.controller.retry,
                      onChooseLocation: widget.onChooseLocation,
                    ),
                    WeatherFailure(:final message, :final canRetry) =>
                      _WeatherProblem(
                        title: 'Could not load weather',
                        message: message,
                        onRetry: canRetry ? widget.controller.retry : null,
                        onChooseLocation: widget.onChooseLocation,
                      ),
                  },
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}

class _WeatherContent extends StatelessWidget {
  const _WeatherContent({
    required this.forecast,
    required this.refreshErrorMessage,
    required this.onRefresh,
    required this.onChooseLocation,
  });

  final LatestForecast forecast;
  final String? refreshErrorMessage;
  final Future<void> Function() onRefresh;
  final VoidCallback onChooseLocation;

  @override
  Widget build(BuildContext context) {
    final weather = forecast.current.weather;
    final conditions = CurrentConditionsViewData.fromForecast(forecast);
    final condition = weatherConditionFromWmoCode(weather.weatherCode ?? -1);
    final isDay = WeatherDaylightResolver.resolve(forecast);
    final freshness = ForecastFreshness.fromForecast(forecast);
    final updatedLabel = freshness.updatedLabel(
      MaterialLocalizations.of(context),
    );
    final today = forecast.daily.isEmpty ? null : forecast.daily.first;
    final hasSunData =
        today != null &&
        (today.sunrise != null ||
            today.sunset != null ||
            today.daylightDurationSeconds != null);

    final scrollView = SingleChildScrollView(
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(24, 20, 24, 28),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _LocationHeader(
            city: forecast.location.city,
            country: forecast.location.country,
            onChooseLocation: onChooseLocation,
          ),
          const SizedBox(height: 4),
          Padding(
            padding: const EdgeInsets.only(left: 28),
            child: Text(
              updatedLabel,
              key: const ValueKey('forecast-updated-at'),
              style: const TextStyle(color: Colors.white70, fontSize: 13),
            ),
          ),
          if (freshness.isStale) ...[
            const SizedBox(height: 12),
            const _StaleForecastNotice(),
          ],
          if (refreshErrorMessage != null) ...[
            const SizedBox(height: 12),
            _RefreshError(message: refreshErrorMessage!),
          ],
          const SizedBox(height: 28),
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
              key: const ValueKey('current-temperature'),
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
                          conditions.precipitationProbability,
                          suffix: '%',
                        ),
                      ),
                    ),
                    Expanded(
                      child: _Metric(
                        icon: Icons.wb_sunny_outlined,
                        label: 'UV index',
                        value: _decimal(conditions.uvIndex),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          if (forecast.hourly.isNotEmpty) ...[
            const SizedBox(height: 20),
            HourlyForecastSection(
              forecasts: forecast.hourly,
              currentValidAt: forecast.current.validAt,
              timezone: forecast.timezone,
            ),
          ],
          const SizedBox(height: 20),
          WeatherDetailsSection(conditions: conditions, units: forecast.units),
          if (hasSunData) ...[
            const SizedBox(height: 20),
            SunCycleSection(forecast: today, timezone: forecast.timezone),
          ],
          if (forecast.daily.isNotEmpty) ...[
            const SizedBox(height: 20),
            DailyForecastSection(forecasts: forecast.daily),
          ],
        ],
      ),
    );

    return RefreshIndicator(
      onRefresh: onRefresh,
      color: Theme.of(context).colorScheme.primary,
      child: scrollView,
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

class _StaleForecastNotice extends StatelessWidget {
  const _StaleForecastNotice();

  @override
  Widget build(BuildContext context) {
    return Container(
      key: const ValueKey('stale-forecast-notice'),
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 11),
      decoration: BoxDecoration(
        color: const Color(0xEFFFF8E7),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0x52E2A53A)),
      ),
      child: const Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: EdgeInsets.only(top: 1),
            child: Icon(
              Icons.schedule_rounded,
              size: 19,
              color: Color(0xFF9A6200),
            ),
          ),
          SizedBox(width: 9),
          Expanded(
            child: Text(
              'Weather data may be out of date. Pull down to check for a newer update.',
              style: TextStyle(
                color: Color(0xFF714A08),
                fontSize: 13,
                height: 1.35,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _RefreshError extends StatelessWidget {
  const _RefreshError({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      key: const ValueKey('forecast-refresh-error'),
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.9),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        children: [
          Icon(
            Icons.info_outline_rounded,
            size: 18,
            color: Theme.of(context).colorScheme.error,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              'Showing previous forecast. $message',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
        ],
      ),
    );
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

class _WeatherPreparing extends StatelessWidget {
  const _WeatherPreparing({
    required this.onRetry,
    required this.onChooseLocation,
  });

  final VoidCallback onRetry;
  final VoidCallback onChooseLocation;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Container(
          key: const ValueKey('preparing-forecast'),
          padding: const EdgeInsets.all(24),
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
            mainAxisSize: MainAxisSize.min,
            children: [
              const SizedBox(
                width: 44,
                height: 44,
                child: CircularProgressIndicator(strokeWidth: 3),
              ),
              const SizedBox(height: 18),
              Text(
                'Preparing forecast',
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 8),
              Text(
                'This city is new. Its latest weather should be ready within '
                'about a minute, and this screen will update automatically.',
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 20),
              FilledButton.icon(
                onPressed: onRetry,
                icon: const Icon(Icons.refresh_rounded),
                label: const Text('Check now'),
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
