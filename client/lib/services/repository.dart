import '../models/category.dart';
import '../models/transaction.dart';
import '../models/transaction_type.dart';

/// Абстракция хранилища данных.
///
/// Текущая реализация — [InMemoryRepository] для UI-прототипа.
/// В дальнейшем появится [ApiRepository], который будет ходить в REST API.
abstract class Repository {
  /// Все категории.
  Future<List<Category>> getCategories();

  Future<void> addCategory(Category category);

  Future<void> removeCategory(String id);

  /// Все операции.
  Future<List<Transaction>> getTransactions();

  Future<void> addTransaction(Transaction transaction);

  Future<void> removeTransaction(String id);
}

/// In-memory хранилище с демонстрационными данными.
class InMemoryRepository implements Repository {
  final List<Category> _categories = [];
  final List<Transaction> _transactions = [];

  InMemoryRepository() {
    _seed();
  }

  void _seed() {
    _categories.addAll([
      Category(
        id: 'c1',
        name: 'Продукты',
        type: TransactionType.expense,
        icon: 'shopping_cart',
        color: 0xFF66BB6A,
      ),
      Category(
        id: 'c2',
        name: 'Транспорт',
        type: TransactionType.expense,
        icon: 'directions_bus',
        color: 0xFF42A5F5,
      ),
      Category(
        id: 'c3',
        name: 'Развлечения',
        type: TransactionType.expense,
        icon: 'sports_esports',
        color: 0xFFAB47BC,
      ),
      Category(
        id: 'c4',
        name: 'Зарплата',
        type: TransactionType.income,
        icon: 'payments',
        color: 0xFF26A69A,
      ),
      Category(
        id: 'c5',
        name: 'Подработка',
        type: TransactionType.income,
        icon: 'work',
        color: 0xFFEF5350,
      ),
    ]);

    final now = DateTime.now();
    _transactions.addAll([
      Transaction(
        id: 't1',
        categoryId: 'c1',
        type: TransactionType.expense,
        amount: 1250.50,
        date: now,
        currency: 'RUB',
        description: 'Супермаркет',
      ),
      Transaction(
        id: 't2',
        categoryId: 'c2',
        type: TransactionType.expense,
        amount: 60,
        date: now.subtract(const Duration(hours: 3)),
        currency: 'RUB',
      ),
      Transaction(
        id: 't3',
        categoryId: 'c4',
        type: TransactionType.income,
        amount: 85000,
        date: now.subtract(const Duration(days: 1)),
        currency: 'RUB',
      ),
      Transaction(
        id: 't4',
        categoryId: 'c3',
        type: TransactionType.expense,
        amount: 1490,
        date: now.subtract(const Duration(days: 2)),
        currency: 'RUB',
        description: 'Билеты в кино',
      ),
    ]);
  }

  @override
  Future<List<Category>> getCategories() async =>
      List.unmodifiable(_categories);

  @override
  Future<void> addCategory(Category category) async {
    _categories.add(category);
  }

  @override
  Future<void> removeCategory(String id) async {
    _categories.removeWhere((c) => c.id == id);
    _transactions.removeWhere((t) => t.categoryId == id);
  }

  @override
  Future<List<Transaction>> getTransactions() async =>
      List.unmodifiable(_transactions);

  @override
  Future<void> addTransaction(Transaction transaction) async {
    _transactions.add(transaction);
  }

  @override
  Future<void> removeTransaction(String id) async {
    _transactions.removeWhere((t) => t.id == id);
  }
}