import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import '../domain/supported_country.dart';
import 'models/location_result.dart';

class LocationApiClient {
  const LocationApiClient({
    required http.Client httpClient,
    required Uri baseUri,
    Duration timeout = const Duration(seconds: 10),
  }) : _httpClient = httpClient,
       _baseUri = baseUri,
       _timeout = timeout;

  final http.Client _httpClient;
  final Uri _baseUri;
  final Duration _timeout;

  Future<List<LocationResult>> searchAvailable({
    required String query,
    required SupportedCountry country,
    int limit = 10,
  }) {
    return _search(
      path: '/api/v1/locations/search',
      query: query,
      country: country,
      limit: limit,
    );
  }

  Future<List<LocationResult>> searchSaved({
    required String query,
    required SupportedCountry country,
    int limit = 10,
  }) {
    return _search(
      path: '/api/v1/locations',
      query: query,
      country: country,
      limit: limit,
    );
  }

  Future<LocationResult> add(int openMeteoLocationId) async {
    if (openMeteoLocationId < 1) {
      throw ArgumentError.value(
        openMeteoLocationId,
        'openMeteoLocationId',
        'must be a positive integer',
      );
    }

    final uri = _baseUri.resolve('/api/v1/locations');
    final response = await _send(
      () => _httpClient.post(
        uri,
        headers: const {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body: jsonEncode({'open_meteo_location_id': openMeteoLocationId}),
      ),
    );
    _requireStatus(response, 201);

    try {
      final envelope = _decodeObject(response);
      return LocationResult.fromJson(_requiredMap(envelope, 'data'));
    } on FormatException catch (error) {
      throw InvalidLocationResponseException(
        'The location service returned an invalid response.',
        cause: error,
      );
    }
  }

  Future<List<LocationResult>> _search({
    required String path,
    required String query,
    required SupportedCountry country,
    required int limit,
  }) async {
    final trimmedQuery = query.trim();
    if (trimmedQuery.isEmpty) {
      throw ArgumentError.value(query, 'query', 'must not be empty');
    }
    if (limit < 1 || limit > 100) {
      throw RangeError.range(limit, 1, 100, 'limit');
    }

    final uri = _baseUri
        .resolve(path)
        .replace(
          queryParameters: {
            'q': trimmedQuery,
            'country_code': country.code,
            'limit': '$limit',
          },
        );
    final response = await _send(
      () => _httpClient.get(uri, headers: const {'Accept': 'application/json'}),
    );
    _requireStatus(response, 200);

    try {
      final envelope = _decodeObject(response);
      final items = envelope['data'];
      if (items is! List<dynamic>) {
        throw const FormatException('data must be a JSON array');
      }

      return items
          .map((item) => LocationResult.fromJson(_asMap(item, 'location item')))
          .toList(growable: false);
    } on FormatException catch (error) {
      throw InvalidLocationResponseException(
        'The location service returned an invalid response.',
        cause: error,
      );
    }
  }

  Future<http.Response> _send(Future<http.Response> Function() request) async {
    try {
      return await request().timeout(_timeout);
    } on TimeoutException catch (error) {
      throw LocationNetworkException(
        'The location request timed out.',
        cause: error,
      );
    } on http.ClientException catch (error) {
      throw LocationNetworkException(
        'The location service could not be reached.',
        cause: error,
      );
    }
  }

  void _requireStatus(http.Response response, int expectedStatus) {
    if (response.statusCode != expectedStatus) {
      throw LocationHttpException(
        _readErrorMessage(response),
        statusCode: response.statusCode,
      );
    }
  }

  Map<String, dynamic> _decodeObject(http.Response response) {
    final decoded = jsonDecode(utf8.decode(response.bodyBytes));
    return _asMap(decoded, 'response');
  }

  String _readErrorMessage(http.Response response) {
    try {
      final decoded = _decodeObject(response);
      final error = decoded['error'];
      if (error is Map<String, dynamic>) {
        final message = error['message'];
        if (message is String && message.isNotEmpty) {
          return message;
        }
      }
    } on FormatException {
      // The HTTP status remains useful when no structured error is available.
    }

    return 'The location request failed with status ${response.statusCode}.';
  }
}

Map<String, dynamic> _requiredMap(Map<String, dynamic> json, String key) {
  return _asMap(json[key], key);
}

Map<String, dynamic> _asMap(Object? value, String field) {
  if (value is Map<String, dynamic>) {
    return value;
  }

  throw FormatException('$field must be a JSON object');
}

sealed class LocationApiException implements Exception {
  const LocationApiException(this.message, {this.statusCode, this.cause});

  final String message;
  final int? statusCode;
  final Object? cause;

  @override
  String toString() => 'LocationApiException: $message';
}

final class LocationHttpException extends LocationApiException {
  const LocationHttpException(super.message, {required super.statusCode});
}

final class LocationNetworkException extends LocationApiException {
  const LocationNetworkException(super.message, {required super.cause});
}

final class InvalidLocationResponseException extends LocationApiException {
  const InvalidLocationResponseException(super.message, {required super.cause});
}
