import 'transaction_type.dart';

/// Одна операция (расход или доход).
class Transaction {
  Transaction({
    required this.id,
    required this.categoryId,
    required this.type,
    required this.amount,
    required this.date,
    required this.currency,
    this.description,
  });

  final String id;
  final String categoryId;
  final TransactionType type;

  /// Сумма всегда хранится в положительном виде.
  final double amount;
  final DateTime date;
  final String currency;
  final String? description;

  factory Transaction.fromJson(Map<String, dynamic> json) {
    return Transaction(
      id: json['id'] as String,
      categoryId: json['categoryId'] as String,
      type: TransactionType.values.byName(json['type'] as String),
      amount: (json['amount'] as num).toDouble(),
      date: DateTime.parse(json['date'] as String),
      currency: json['currency'] as String,
      description: json['description'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'categoryId': categoryId,
        'type': type.name,
        'amount': amount,
        'date': date.toIso8601String(),
        'currency': currency,
        if (description != null) 'description': description,
      };
}