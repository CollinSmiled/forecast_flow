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
    await _request(locationId);
  }

  Future<void> refresh() async {
    final locationId = _locationId;
    final retainedForecast = switch (_state) {
      WeatherLoaded(:final forecast) => forecast,
      _ => null,
    };
    if (locationId == null) {
      return;
    }

    await _request(locationId, retainedForecast: retainedForecast);
  }

  Future<void> _request(
    int locationId, {
    LatestForecast? retainedForecast,
  }) async {
    final requestVersion = ++_requestVersion;
    if (retainedForecast == null) {
      _emit(const WeatherLoading());
    }

    try {
      final forecast = await _loadForecast(locationId);
      if (!_isCurrent(requestVersion)) {
        return;
      }

      _emit(WeatherLoaded(forecast));
    } on ForecastNotFoundException catch (error) {
      _emitProblem(
        requestVersion,
        WeatherNotFound(error.message),
        retainedForecast,
      );
    } on ForecastNetworkException catch (error) {
      _emitProblem(
        requestVersion,
        WeatherFailure(message: error.message, canRetry: true),
        retainedForecast,
      );
    } on InvalidForecastResponseException catch (error) {
      _emitProblem(
        requestVersion,
        WeatherFailure(message: error.message, canRetry: true),
        retainedForecast,
      );
    } on ForecastHttpException catch (error) {
      _emitProblem(
        requestVersion,
        WeatherFailure(
          message: error.message,
          canRetry: error.statusCode == null || error.statusCode! >= 500,
        ),
        retainedForecast,
      );
    } on Object {
      _emitProblem(
        requestVersion,
        const WeatherFailure(
          message: 'The forecast could not be loaded.',
          canRetry: true,
        ),
        retainedForecast,
      );
    }
  }

  void _emitProblem(
    int requestVersion,
    WeatherViewState problem,
    LatestForecast? retainedForecast,
  ) {
    if (!_isCurrent(requestVersion)) {
      return;
    }

    if (retainedForecast == null) {
      _emit(problem);
      return;
    }

    final message = switch (problem) {
      WeatherNotFound(:final message) => message,
      WeatherFailure(:final message) => message,
      _ => 'The forecast could not be refreshed.',
    };
    _emit(WeatherLoaded(retainedForecast, refreshErrorMessage: message));
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
  const WeatherLoaded(this.forecast, {this.refreshErrorMessage});

  final LatestForecast forecast;
  final String? refreshErrorMessage;
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
