import 'package:flutter_test/flutter_test.dart';
import 'package:forecast_flow_mobile/app/app.dart';

import 'helpers/memory_selected_location_store.dart';

void main() {
  testWidgets('starts with city selection', (tester) async {
    await tester.pumpWidget(
      ForecastFlowApp(selectedLocationStore: MemorySelectedLocationStore()),
    );
    await tester.pumpAndSettle();

    expect(find.text('Choose a city'), findsOneWidget);
    expect(find.text('Find your city'), findsOneWidget);
  });
}
