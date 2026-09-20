class LocationResult {
  const LocationResult({
    required this.openMeteoLocationId,
    required this.city,
    required this.country,
    required this.countryCode,
    required this.latitude,
    required this.longitude,
    required this.timezone,
    this.locationId,
    this.elevation,
    this.population,
    this.administrativeArea,
  });

  final int? locationId;
  final int openMeteoLocationId;
  final String city;
  final String country;
  final String countryCode;
  final double latitude;
  final double longitude;
  final String timezone;
  final double? elevation;
  final int? population;
  final String? administrativeArea;

  bool get isSaved => locationId != null;

  factory LocationResult.fromJson(Map<String, dynamic> json) {
    return LocationResult(
      locationId: _optionalInt(json, 'location_id'),
      openMeteoLocationId: _requiredInt(json, 'open_meteo_location_id'),
      city: _requiredString(json, 'city'),
      country: _requiredString(json, 'country'),
      countryCode: _requiredString(json, 'country_code'),
      latitude: _requiredDouble(json, 'latitude'),
      longitude: _requiredDouble(json, 'longitude'),
      timezone: _requiredString(json, 'timezone'),
      elevation: _optionalDouble(json, 'elevation'),
      population: _optionalInt(json, 'population'),
      administrativeArea: _optionalString(json, 'administrative_area'),
    );
  }
}

String _requiredString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is String && value.isNotEmpty) {
    return value;
  }

  throw FormatException('$key must be a non-empty string');
}

String? _optionalString(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null || value is String) {
    return value as String?;
  }

  throw FormatException('$key must be a string or null');
}

int _requiredInt(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is int) {
    return value;
  }

  throw FormatException('$key must be an integer');
}

int? _optionalInt(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null || value is int) {
    return value as int?;
  }

  throw FormatException('$key must be an integer or null');
}

double _requiredDouble(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value is num) {
    return value.toDouble();
  }

  throw FormatException('$key must be a number');
}

double? _optionalDouble(Map<String, dynamic> json, String key) {
  final value = json[key];
  if (value == null) {
    return null;
  }
  if (value is num) {
    return value.toDouble();
  }

  throw FormatException('$key must be a number or null');
}
