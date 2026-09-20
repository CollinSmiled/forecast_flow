import 'package:flutter/material.dart';

import '../core/theme/app_theme.dart';
import '../features/weather/presentation/weather_home_screen.dart';

class ForecastFlowApp extends StatelessWidget {
  const ForecastFlowApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Forecast Flow',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      home: const WeatherHomeScreen(),
    );
  }
}
