import 'package:flutter_test/flutter_test.dart';

import 'package:budget_tracker/services/auth_service.dart';

void main() {
  group('InMemoryAuthService', () {
    test('регистрация создаёт пользователя и выполняет вход', () async {
      final service = InMemoryAuthService();

      final user = await service.register('alice', '1234');

      expect(user.login, 'alice');
      expect(service.currentUser?.login, 'alice');
    });

    test('вход с верным паролем успешен', () async {
      final service = InMemoryAuthService();
      await service.register('bob', 'pass');

      final user = await service.login('bob', 'pass');

      expect(user.login, 'bob');
    });

    test('вход с неверным паролем бросает AuthException', () async {
      final service = InMemoryAuthService();
      await service.register('bob', 'pass');

      expect(
        () => service.login('bob', 'wrong'),
        throwsA(isA<AuthException>()),
      );
    });

    test('регистрация с дублирующимся логином бросает AuthException', () async {
      final service = InMemoryAuthService();
      await service.register('alice', '1234');

      expect(
        () => service.register('alice', '4321'),
        throwsA(isA<AuthException>()),
      );
    });

    test('logout очищает текущего пользователя', () async {
      final service = InMemoryAuthService();
      await service.register('alice', '1234');

      await service.logout();

      expect(service.currentUser, isNull);
    });
  });
}