import '../../../core/time/app_time.dart';

abstract final class CityClock {
  static String? label(String timezone, {DateTime? now}) {
    final instant = (now ?? DateTime.now()).toUtc();
    final local = AppTime.tryAtLocation(instant, timezone);
    if (local == null) {
      return null;
    }

    final period = local.hour < 12 ? 'AM' : 'PM';
    final hour = local.hour % 12 == 0 ? 12 : local.hour % 12;
    final minute = local.minute.toString().padLeft(2, '0');
    return 'Local time $hour:$minute $period';
  }
}
