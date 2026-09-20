import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'models/latest_forecast.dart';

class ForecastApiClient {
  const ForecastApiClient({
    required http.Client httpClient,
    required Uri baseUri,
    Duration timeout = const Duration(seconds: 10),
  }) : _httpClient = httpClient,
       _baseUri = baseUri,
       _timeout = timeout;

  final http.Client _httpClient;
  final Uri _baseUri;
  final Duration _timeout;

  Future<LatestForecast> getLatest(int locationId) async {
    if (locationId < 1) {
      throw ArgumentError.value(
        locationId,
        'locationId',
        'must be a positive integer',
      );
    }

    final uri = _baseUri.resolve('/api/v1/locations/$locationId/forecast');

    try {
      final response = await _httpClient
          .get(uri, headers: const {'Accept': 'application/json'})
          .timeout(_timeout);

      if (response.statusCode == 404) {
        throw ForecastNotFoundException(_readErrorMessage(response));
      }
      if (response.statusCode != 200) {
        throw ForecastHttpException(
          _readErrorMessage(response),
          statusCode: response.statusCode,
        );
      }

      final decoded = jsonDecode(utf8.decode(response.bodyBytes));
      if (decoded is! Map<String, dynamic>) {
        throw const FormatException('response must be a JSON object');
      }

      return LatestForecast.fromEnvelope(decoded);
    } on ForecastApiException {
      rethrow;
    } on TimeoutException catch (error) {
      throw ForecastNetworkException(
        'The forecast request timed out.',
        cause: error,
      );
    } on http.ClientException catch (error) {
      throw ForecastNetworkException(
        'The forecast service could not be reached.',
        cause: error,
      );
    } on FormatException catch (error) {
      throw InvalidForecastResponseException(
        'The forecast service returned an invalid response.',
        cause: error,
      );
    }
  }

  String _readErrorMessage(http.Response response) {
    try {
      final decoded = jsonDecode(utf8.decode(response.bodyBytes));
      if (decoded is Map<String, dynamic>) {
        final error = decoded['error'];
        if (error is Map<String, dynamic>) {
          final message = error['message'];
          if (message is String && message.isNotEmpty) {
            return message;
          }
        }
      }
    } on FormatException {
      // A non-success response is still represented by its status code when
      // the server does not provide the expected JSON error envelope.
    }

    return 'The forecast request failed with status ${response.statusCode}.';
  }
}

sealed class ForecastApiException implements Exception {
  const ForecastApiException(this.message, {this.statusCode, this.cause});

  final String message;
  final int? statusCode;
  final Object? cause;

  @override
  String toString() => 'ForecastApiException: $message';
}

final class ForecastNotFoundException extends ForecastApiException {
  const ForecastNotFoundException(super.message) : super(statusCode: 404);
}

final class ForecastHttpException extends ForecastApiException {
  const ForecastHttpException(super.message, {required super.statusCode});
}

final class ForecastNetworkException extends ForecastApiException {
  const ForecastNetworkException(super.message, {required super.cause});
}

final class InvalidForecastResponseException extends ForecastApiException {
  const InvalidForecastResponseException(super.message, {required super.cause});
}
