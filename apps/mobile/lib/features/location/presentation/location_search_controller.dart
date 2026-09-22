import 'package:flutter/foundation.dart';

import '../data/location_api_client.dart';
import '../data/models/location_result.dart';
import '../domain/supported_country.dart';

typedef AvailableLocationSearcher =
    Future<List<LocationResult>> Function({
      required String query,
      required SupportedCountry country,
      int limit,
    });

typedef LocationSaver =
    Future<LocationResult> Function(int openMeteoLocationId);

class LocationSearchController extends ChangeNotifier {
  LocationSearchController({
    required AvailableLocationSearcher searchAvailable,
    required LocationSaver saveLocation,
    SupportedCountry initialCountry = SupportedCountry.indonesia,
  }) : _searchAvailable = searchAvailable,
       _saveLocation = saveLocation,
       _state = LocationSearchState.initial(initialCountry);

  final AvailableLocationSearcher _searchAvailable;
  final LocationSaver _saveLocation;

  LocationSearchState _state;
  int _requestVersion = 0;
  bool _isDisposed = false;

  LocationSearchState get state => _state;

  void reset({SupportedCountry? country}) {
    _requestVersion++;
    _emit(LocationSearchState.initial(country ?? _state.country));
  }

  Future<void> search({
    required String query,
    required SupportedCountry country,
  }) async {
    final trimmedQuery = query.trim();
    if (trimmedQuery.isEmpty) {
      throw ArgumentError.value(query, 'query', 'must not be empty');
    }

    final requestVersion = ++_requestVersion;
    _emit(
      LocationSearchState(
        status: LocationSearchStatus.searching,
        query: trimmedQuery,
        country: country,
      ),
    );

    try {
      final results = await _searchAvailable(
        query: trimmedQuery,
        country: country,
        limit: 10,
      );
      if (!_isCurrent(requestVersion)) {
        return;
      }

      _emit(
        LocationSearchState(
          status: results.isEmpty
              ? LocationSearchStatus.empty
              : LocationSearchStatus.results,
          query: trimmedQuery,
          country: country,
          results: results,
        ),
      );
    } on LocationApiException catch (error) {
      if (_isCurrent(requestVersion)) {
        _emit(
          LocationSearchState(
            status: LocationSearchStatus.failure,
            query: trimmedQuery,
            country: country,
            errorMessage: error.message,
          ),
        );
      }
    } on Object {
      if (_isCurrent(requestVersion)) {
        _emit(
          LocationSearchState(
            status: LocationSearchStatus.failure,
            query: trimmedQuery,
            country: country,
            errorMessage: 'Locations could not be loaded.',
          ),
        );
      }
    }
  }

  Future<void> retry() async {
    if (_state.query.isNotEmpty) {
      await search(query: _state.query, country: _state.country);
    }
  }

  Future<LocationResult?> save(LocationResult location) async {
    if (location.isSaved) {
      return location;
    }
    if (_state.savingOpenMeteoLocationId != null) {
      return null;
    }

    final requestVersion = _requestVersion;
    _emit(
      LocationSearchState(
        status: LocationSearchStatus.results,
        query: _state.query,
        country: _state.country,
        results: _state.results,
        savingOpenMeteoLocationId: location.openMeteoLocationId,
      ),
    );

    try {
      final saved = await _saveLocation(location.openMeteoLocationId);
      if (!_isCurrent(requestVersion)) {
        return null;
      }

      final updatedResults = _state.results
          .map(
            (result) => result.openMeteoLocationId == saved.openMeteoLocationId
                ? saved
                : result,
          )
          .toList(growable: false);
      _emit(
        LocationSearchState(
          status: LocationSearchStatus.results,
          query: _state.query,
          country: _state.country,
          results: updatedResults,
        ),
      );

      return saved;
    } on LocationApiException catch (error) {
      if (_isCurrent(requestVersion)) {
        _emitSaveFailure(error.message);
      }
    } on Object {
      if (_isCurrent(requestVersion)) {
        _emitSaveFailure('The city could not be saved.');
      }
    }

    return null;
  }

  void _emitSaveFailure(String message) {
    _emit(
      LocationSearchState(
        status: LocationSearchStatus.results,
        query: _state.query,
        country: _state.country,
        results: _state.results,
        errorMessage: message,
      ),
    );
  }

  bool _isCurrent(int requestVersion) {
    return !_isDisposed && requestVersion == _requestVersion;
  }

  void _emit(LocationSearchState nextState) {
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

enum LocationSearchStatus { initial, searching, results, empty, failure }

@immutable
class LocationSearchState {
  LocationSearchState({
    required this.status,
    required this.query,
    required this.country,
    List<LocationResult> results = const [],
    this.savingOpenMeteoLocationId,
    this.errorMessage,
  }) : results = List.unmodifiable(results);

  factory LocationSearchState.initial(SupportedCountry country) {
    return LocationSearchState(
      status: LocationSearchStatus.initial,
      query: '',
      country: country,
    );
  }

  final LocationSearchStatus status;
  final String query;
  final SupportedCountry country;
  final List<LocationResult> results;
  final int? savingOpenMeteoLocationId;
  final String? errorMessage;

  bool get isSaving => savingOpenMeteoLocationId != null;
}
