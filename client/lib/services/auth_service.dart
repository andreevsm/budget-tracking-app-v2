import '../models/user.dart';

/// Ошибка авторизации/регистрации с понятным сообщением.
class AuthException implements Exception {
  AuthException(this.message);

  final String message;

  @override
  String toString() => message;
}

/// Абстракция сервиса аутентификации.
///
/// Текущая реализация — [InMemoryAuthService] для UI-прототипа.
/// В дальнейшем появится реализация, работающая через REST API.
abstract class AuthService {
  /// Регистрация нового пользователя.
  Future<User> register(String login, String password);

  /// Вход по логину и паролю.
  Future<User> login(String login, String password);

  /// Текущий пользователь, если выполнен вход.
  User? get currentUser;

  /// Выход из аккаунта.
  Future<void> logout();
}

/// In-memory реализация авторизации с хранением пользователей в памяти.
class InMemoryAuthService implements AuthService {
  final Map<String, String> _users = {};
  User? _currentUser;

  @override
  User? get currentUser => _currentUser;

  @override
  Future<User> register(String login, String password) async {
    await Future<void>.delayed(const Duration(milliseconds: 300));
    final normalized = login.trim();
    if (normalized.length < 3) {
      throw AuthException('Логин должен содержать минимум 3 символа');
    }
    if (password.length < 4) {
      throw AuthException('Пароль должен содержать минимум 4 символа');
    }
    if (_users.containsKey(normalized)) {
      throw AuthException('Пользователь с таким логином уже существует');
    }
    _users[normalized] = password;
    _currentUser = User(id: normalized, login: normalized);
    return _currentUser!;
  }

  @override
  Future<User> login(String login, String password) async {
    await Future<void>.delayed(const Duration(milliseconds: 300));
    final normalized = login.trim();
    if (_users[normalized] != password) {
      throw AuthException('Неверный логин или пароль');
    }
    _currentUser = User(id: normalized, login: normalized);
    return _currentUser!;
  }

  @override
  Future<void> logout() async {
    _currentUser = null;
  }
}