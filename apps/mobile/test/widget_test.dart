import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/app/app.dart';

void main() {
  testWidgets('starts with city selection', (tester) async {
    await tester.pumpWidget(const ForecastFlowApp());

    expect(find.text('Choose a city'), findsOneWidget);
    expect(find.text('Find your city'), findsOneWidget);
  });
}
