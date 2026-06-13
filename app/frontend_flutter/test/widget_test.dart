import 'package:flutter_test/flutter_test.dart';

import 'package:frontend_flutter/main.dart';

void main() {
  testWidgets('app renders landing screen', (WidgetTester tester) async {
    await tester.pumpWidget(const HaddleApp());

    await tester.pump();

    expect(find.text('Haddle'), findsWidgets);
    expect(find.text('Login'), findsOneWidget);
  });
}
