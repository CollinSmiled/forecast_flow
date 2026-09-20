import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/features/location/domain/supported_country.dart';

void main() {
  test('contains only the selected Asian country codes', () {
    expect(SupportedCountry.values.map((country) => country.code), [
      'ID',
      'SG',
      'MY',
      'TH',
      'VN',
      'PH',
      'JP',
      'KR',
      'CN',
    ]);
    expect(
      SupportedCountry.values.any((country) => country.code == 'IN'),
      isFalse,
    );
  });
}
