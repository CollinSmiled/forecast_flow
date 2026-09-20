import 'package:timezone/data/latest_10y.dart' as timezone_data;
import 'package:timezone/timezone.dart' as timezone;

abstract final class AppTime {
  static bool _initialized = false;

  static void initialize() {
    if (_initialized) {
      return;
    }

    timezone_data.initializeTimeZones();
    _initialized = true;
  }

  static timezone.TZDateTime? tryAtLocation(
    DateTime instant,
    String locationName,
  ) {
    initialize();

    try {
      final location = timezone.getLocation(locationName);
      return timezone.TZDateTime.from(instant.toUtc(), location);
    } on timezone.LocationNotFoundException {
      return null;
    }
  }
}
