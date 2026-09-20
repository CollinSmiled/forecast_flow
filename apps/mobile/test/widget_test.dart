import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/app/app.dart';

void main() {
  testWidgets('renders the weather home foundation', (tester) async {
    await tester.pumpWidget(const ForecastFlowApp());

    expect(find.text('Jakarta, Indonesia'), findsOneWidget);
    expect(find.text('Weather at a glance'), findsOneWidget);
  });
}
