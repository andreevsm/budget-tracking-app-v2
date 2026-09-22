/// Тип операции: расход или доход.
enum TransactionType {
  expense,
  income;

  String get label => switch (this) {
        TransactionType.expense => 'Расход',
        TransactionType.income => 'Доход',
      };
}