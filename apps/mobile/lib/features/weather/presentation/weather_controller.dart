import 'package:flutter/foundation.dart';

import '../data/forecast_api_client.dart';
import '../data/models/latest_forecast.dart';

typedef ForecastLoader = Future<LatestForecast> Function(int locationId);

class WeatherController extends ChangeNotifier {
  WeatherController({required ForecastLoader loadForecast})
    : _loadForecast = loadForecast;

  final ForecastLoader _loadForecast;

  WeatherViewState _state = const WeatherInitial();
  int? _locationId;
  int _requestVersion = 0;
  bool _isDisposed = false;

  WeatherViewState get state => _state;

  Future<void> load(int locationId) async {
    if (locationId < 1) {
      throw ArgumentError.value(
        locationId,
        'locationId',
        'must be a positive integer',
      );
    }

    _locationId = locationId;
    final requestVersion = ++_requestVersion;
    _emit(const WeatherLoading());

    try {
      final forecast = await _loadForecast(locationId);
      if (!_isCurrent(requestVersion)) {
        return;
      }

      _emit(WeatherLoaded(forecast));
    } on ForecastNotFoundException catch (error) {
      if (_isCurrent(requestVersion)) {
        _emit(WeatherNotFound(error.message));
      }
    } on ForecastNetworkException catch (error) {
      if (_isCurrent(requestVersion)) {
        _emit(WeatherFailure(message: error.message, canRetry: true));
      }
    } on InvalidForecastResponseException catch (error) {
      if (_isCurrent(requestVersion)) {
        _emit(WeatherFailure(message: error.message, canRetry: true));
      }
    } on ForecastHttpException catch (error) {
      if (_isCurrent(requestVersion)) {
        _emit(
          WeatherFailure(
            message: error.message,
            canRetry: error.statusCode == null || error.statusCode! >= 500,
          ),
        );
      }
    } on Object {
      if (_isCurrent(requestVersion)) {
        _emit(
          const WeatherFailure(
            message: 'The forecast could not be loaded.',
            canRetry: true,
          ),
        );
      }
    }
  }

  Future<void> retry() async {
    final locationId = _locationId;
    if (locationId != null) {
      await load(locationId);
    }
  }

  bool _isCurrent(int requestVersion) {
    return !_isDisposed && requestVersion == _requestVersion;
  }

  void _emit(WeatherViewState nextState) {
    if (_isDisposed) {
      return;
    }

    _state = nextState;
    notifyListeners();
  }

  @override
  void dispose() {
    _isDisposed = true;
    _requestVersion++;
    super.dispose();
  }
}

sealed class WeatherViewState {
  const WeatherViewState();
}

final class WeatherInitial extends WeatherViewState {
  const WeatherInitial();
}

final class WeatherLoading extends WeatherViewState {
  const WeatherLoading();
}

final class WeatherLoaded extends WeatherViewState {
  const WeatherLoaded(this.forecast);

  final LatestForecast forecast;
}

final class WeatherNotFound extends WeatherViewState {
  const WeatherNotFound(this.message);

  final String message;
}

final class WeatherFailure extends WeatherViewState {
  const WeatherFailure({required this.message, required this.canRetry});

  final String message;
  final bool canRetry;
}
