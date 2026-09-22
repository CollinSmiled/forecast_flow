import 'package:flutter/material.dart';

import '../../../core/time/app_time.dart';
import '../data/models/latest_forecast.dart';

class ForecastFreshness {
  const ForecastFreshness({
    required this.localRetrievedAt,
    required this.localNow,
    required this.isStale,
  });

  static const staleAfter = Duration(hours: 2);

  final DateTime localRetrievedAt;
  final DateTime localNow;
  final bool isStale;

  factory ForecastFreshness.fromForecast(
    LatestForecast forecast, {
    DateTime? now,
  }) {
    final instant = (now ?? DateTime.now()).toUtc();
    final localRetrievedAt =
        AppTime.tryAtLocation(forecast.retrievedAt, forecast.timezone) ??
        forecast.retrievedAt.toLocal();
    final localNow =
        AppTime.tryAtLocation(instant, forecast.timezone) ?? instant.toLocal();
    final age = instant.difference(forecast.retrievedAt.toUtc());

    return ForecastFreshness(
      localRetrievedAt: localRetrievedAt,
      localNow: localNow,
      isStale: age > staleAfter,
    );
  }

  String updatedLabel(MaterialLocalizations localizations) {
    final time = localizations.formatTimeOfDay(
      TimeOfDay.fromDateTime(localRetrievedAt),
    );
    final retrievedDate = DateUtils.dateOnly(localRetrievedAt);
    final currentDate = DateUtils.dateOnly(localNow);

    if (retrievedDate == currentDate) {
      return 'Updated $time';
    }
    if (retrievedDate == currentDate.subtract(const Duration(days: 1))) {
      return 'Updated yesterday at $time';
    }

    final date = localizations.formatMediumDate(localRetrievedAt);
    return 'Updated $date at $time';
  }
}
