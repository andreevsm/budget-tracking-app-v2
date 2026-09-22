/// Пользователь приложения.
class User {
  User({
    required this.id,
    required this.login,
  });

  final String id;
  final String login;

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      login: json['login'] as String,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'login': login,
      };
}