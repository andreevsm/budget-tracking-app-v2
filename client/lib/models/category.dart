import 'transaction_type.dart';

/// Категория операции (например «Продукты»).
class Category {
  Category({
    required this.id,
    required this.name,
    required this.type,
    this.currency = 'RUB',
    this.icon,
    this.color,
  });

  final String id;
  final String name;
  final TransactionType type;

  /// Валюта категории по умолчанию — рубли (RUB).
  String currency;

  final String? icon;
  final int? color;

  factory Category.fromJson(Map<String, dynamic> json) {
    return Category(
      id: json['id'] as String,
      name: json['name'] as String,
      type: TransactionType.values.byName(json['type'] as String),
      currency: (json['currency'] as String?) ?? 'RUB',
      icon: json['icon'] as String?,
      color: json['color'] as int?,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'type': type.name,
        'currency': currency,
        if (icon != null) 'icon': icon,
        if (color != null) 'color': color,
      };

  Category copyWith({
    String? name,
    TransactionType? type,
    String? currency,
    String? icon,
    int? color,
  }) {
    return Category(
      id: id,
      name: name ?? this.name,
      type: type ?? this.type,
      currency: currency ?? this.currency,
      icon: icon ?? this.icon,
      color: color ?? this.color,
    );
  }
}