import 'package:flutter_test/flutter_test.dart';

import 'package:frontend_flutter/main.dart';

void main() {
  testWidgets('app renders landing screen', (WidgetTester tester) async {
    await tester.pumpWidget(const AlbzApp());

    await tester.pump();

    expect(find.text('Albz'), findsWidgets);
    expect(find.text('Login'), findsOneWidget);
  });
}
