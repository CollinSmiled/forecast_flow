import 'package:flutter/widgets.dart';

import 'app/app.dart';
import 'core/time/app_time.dart';

void main() {
  AppTime.initialize();
  runApp(const ForecastFlowApp());
}
