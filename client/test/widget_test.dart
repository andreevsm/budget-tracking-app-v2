import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';

import 'package:budget_tracker/screens/home_screen.dart';
import 'package:budget_tracker/services/repository.dart';

void main() {
  setUpAll(() async {
    await initializeDateFormatting('ru_RU');
  });

  testWidgets('Главная страница отображается', (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialAppWrapper(repository: InMemoryRepository()),
    );
    await tester.pumpAndSettle();

    expect(find.text('Учёт бюджета'), findsOneWidget);
  });
}

// Небольшой враппер, чтобы протестировать HomeScreen изолированно.
class MaterialAppWrapper extends StatelessWidget {
  const MaterialAppWrapper({super.key, required this.repository});

  final Repository repository;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: HomeScreen(repository: repository),
    );
  }
}