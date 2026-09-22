import 'package:flutter/material.dart';
import 'package:intl/date_symbol_data_local.dart';

import 'models/user.dart';
import 'screens/auth_screen.dart';
import 'screens/home_screen.dart';
import 'services/auth_service.dart';
import 'services/repository.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await initializeDateFormatting('ru_RU');
  runApp(const BudgetTrackerApp());
}

class BudgetTrackerApp extends StatefulWidget {
  const BudgetTrackerApp({super.key});

  @override
  State<BudgetTrackerApp> createState() => _BudgetTrackerAppState();
}

class _BudgetTrackerAppState extends State<BudgetTrackerApp> {
  final AuthService _authService = InMemoryAuthService();
  final Repository _repository = InMemoryRepository();

  User? _user;

  void _handleSuccess(User user) {
    setState(() => _user = user);
  }

  Future<void> _handleLogout() async {
    await _authService.logout();
    if (!mounted) return;
    setState(() => _user = null);
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Учёт бюджета',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF26A69A)),
        useMaterial3: true,
      ),
      home: _user == null
          ? AuthScreen(
              authService: _authService,
              onSuccess: _handleSuccess,
            )
          : HomeScreen(repository: _repository, onLogout: _handleLogout),
    );
  }
}