abstract final class AppConfig {
  static const _apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8080',
  );

  static final Uri apiBaseUri = _parseApiBaseUri(_apiBaseUrl);

  static Uri _parseApiBaseUri(String value) {
    final uri = Uri.tryParse(value);
    if (uri == null || !uri.hasScheme || uri.host.isEmpty) {
      throw StateError('API_BASE_URL must be an absolute URL');
    }

    return uri;
  }
}
