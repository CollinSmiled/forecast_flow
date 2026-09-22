import 'package:flutter/foundation.dart';

import '../data/forecast_api_client.dart';
import '../data/models/latest_forecast.dart';

typedef ForecastLoader = Future<LatestForecast> Function(int locationId);
typedef ForecastRetryWaiter = Future<void> Function(Duration duration);

Future<void> _waitForRetry(Duration duration) => Future<void>.delayed(duration);

class WeatherController extends ChangeNotifier {
  WeatherController({
    required ForecastLoader loadForecast,
    int pendingForecastRetries = 0,
    Duration pendingForecastRetryDelay = const Duration(seconds: 10),
    ForecastRetryWaiter waitForPendingForecast = _waitForRetry,
  }) : assert(pendingForecastRetries >= 0),
       assert(!pendingForecastRetryDelay.isNegative),
       _loadForecast = loadForecast,
       _pendingForecastRetries = pendingForecastRetries,
       _pendingForecastRetryDelay = pendingForecastRetryDelay,
       _waitForPendingForecast = waitForPendingForecast;

  final ForecastLoader _loadForecast;
  final int _pendingForecastRetries;
  final Duration _pendingForecastRetryDelay;
  final ForecastRetryWaiter _waitForPendingForecast;

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
    await _request(locationId, pendingForecastRetries: _pendingForecastRetries);
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
    int pendingForecastRetries = 0,
  }) async {
    final requestVersion = ++_requestVersion;
    if (retainedForecast == null) {
      _emit(const WeatherLoading());
    }

    var retriesRemaining = pendingForecastRetries;
    while (_isCurrent(requestVersion)) {
      try {
        final forecast = await _loadForecast(locationId);
        if (!_isCurrent(requestVersion)) {
          return;
        }

        _emit(WeatherLoaded(forecast));
        return;
      } on ForecastNotFoundException catch (error) {
        if (!_isCurrent(requestVersion)) {
          return;
        }

        if (retainedForecast == null && retriesRemaining > 0) {
          _emit(const WeatherAwaitingForecast());
          retriesRemaining--;
          await _waitForPendingForecast(_pendingForecastRetryDelay);
          continue;
        }

        _emitProblem(
          requestVersion,
          WeatherNotFound(error.message),
          retainedForecast,
        );
        return;
      } on ForecastNetworkException catch (error) {
        _emitProblem(
          requestVersion,
          WeatherFailure(message: error.message, canRetry: true),
          retainedForecast,
        );
        return;
      } on InvalidForecastResponseException catch (error) {
        _emitProblem(
          requestVersion,
          WeatherFailure(message: error.message, canRetry: true),
          retainedForecast,
        );
        return;
      } on ForecastHttpException catch (error) {
        _emitProblem(
          requestVersion,
          WeatherFailure(
            message: error.message,
            canRetry: error.statusCode == null || error.statusCode! >= 500,
          ),
          retainedForecast,
        );
        return;
      } on Object {
        _emitProblem(
          requestVersion,
          const WeatherFailure(
            message: 'The forecast could not be loaded.',
            canRetry: true,
          ),
          retainedForecast,
        );
        return;
      }
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

final class WeatherAwaitingForecast extends WeatherViewState {
  const WeatherAwaitingForecast();
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
