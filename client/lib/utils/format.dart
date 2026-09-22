import 'package:intl/intl.dart';

/// Символ валюты по коду (сокращённое отображение).
String currencySymbol(String code) {
  return switch (code.toUpperCase()) {
    'RUB' => '₽',
    'USD' => '\$',
    'EUR' => '€',
    _ => code.toUpperCase(),
  };
}

/// Форматирование суммы с валютой, например «1 250,50 ₽».
String formatAmount(double amount, String currencyCode) {
  final formatter = NumberFormat.currency(
    locale: 'ru_RU',
    symbol: currencySymbol(currencyCode),
    decimalDigits: amount % 1 == 0 ? 0 : 2,
  );
  return formatter.format(amount);
}

/// Форматирование даты «22 сен 2026».
String formatDay(DateTime date) => DateFormat('d MMM yyyy', 'ru_RU').format(date);

/// Форматирование времени «14:05».
String formatTime(DateTime date) => DateFormat('HH:mm').format(date);