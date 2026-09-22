enum SupportedCountry {
  indonesia(code: 'ID', displayName: 'Indonesia'),
  singapore(code: 'SG', displayName: 'Singapore'),
  malaysia(code: 'MY', displayName: 'Malaysia'),
  thailand(code: 'TH', displayName: 'Thailand'),
  vietnam(code: 'VN', displayName: 'Vietnam'),
  philippines(code: 'PH', displayName: 'Philippines'),
  japan(code: 'JP', displayName: 'Japan'),
  southKorea(code: 'KR', displayName: 'South Korea'),
  china(code: 'CN', displayName: 'China');

  const SupportedCountry({required this.code, required this.displayName});

  final String code;
  final String displayName;
}
