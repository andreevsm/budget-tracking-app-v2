import 'package:flutter/material.dart';

import '../models/category.dart';
import '../models/transaction.dart';
import '../models/transaction_type.dart';
import '../services/repository.dart';
import '../utils/format.dart';
import '../widgets/add_category_dialog.dart';
import '../widgets/add_transaction_dialog.dart';

/// Период отображения операций.
enum PeriodFilter { day, month, year }

/// Главная страница: учёт расходов и доходов.
class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key, required this.repository, this.onLogout});

  final Repository repository;
  final VoidCallback? onLogout;

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  late Future<void> _loadFuture;
  List<Category> _categories = [];
  List<Transaction> _transactions = [];

  TransactionType _type = TransactionType.expense;
  PeriodFilter _period = PeriodFilter.day;

  @override
  void initState() {
    super.initState();
    _loadFuture = _load();
  }

  Future<void> _load() async {
    final categories = await widget.repository.getCategories();
    final transactions = await widget.repository.getTransactions();
    if (!mounted) return;
    setState(() {
      _categories = categories;
      _transactions = transactions;
    });
  }

  void _reload() {
    _loadFuture = _load();
    setState(() {});
  }

  /// Операции выбранного типа за выбранный период.
  List<Transaction> get _filtered {
    final now = DateTime.now();
    return _transactions.where((t) {
      if (t.type != _type) return false;
      return switch (_period) {
        PeriodFilter.day =>
          t.date.year == now.year &&
              t.date.month == now.month &&
              t.date.day == now.day,
        PeriodFilter.month =>
          t.date.year == now.year && t.date.month == now.month,
        PeriodFilter.year => t.date.year == now.year,
      };
    }).toList()
      ..sort((a, b) => b.date.compareTo(a.date));
  }

  /// Итог по операциям, сгруппированный по валютам.
  Map<String, double> get _totalsByCurrency {
    final totals = <String, double>{};
    for (final t in _filtered) {
      totals[t.currency] = (totals[t.currency] ?? 0) + t.amount;
    }
    return totals;
  }

  String get _periodTitle => switch (_period) {
        PeriodFilter.day => 'за сегодня',
        PeriodFilter.month => 'за месяц',
        PeriodFilter.year => 'за год',
      };

  Future<void> _createCategory() async {
    final category = await showAddCategoryDialog(context, _type);
    if (category == null) return;
    await widget.repository.addCategory(category);
    _reload();
  }

  Future<void> _addTransaction(Category category) async {
    final transaction = await showAddTransactionDialog(context, category: category);
    if (transaction == null) return;
    await widget.repository.addTransaction(transaction);
    _reload();
  }

  Future<void> _removeTransaction(Transaction transaction) async {
    await widget.repository.removeTransaction(transaction.id);
    _reload();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Учёт бюджета'),
        actions: [
          PopupMenuButton<String>(
            tooltip: 'Профиль',
            onSelected: (value) {
              if (value == 'logout') widget.onLogout?.call();
            },
            itemBuilder: (context) => const [
              PopupMenuItem(
                value: 'logout',
                child: ListTile(
                  leading: Icon(Icons.logout),
                  title: Text('Выйти'),
                  contentPadding: EdgeInsets.zero,
                ),
              ),
            ],
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _createCategory,
        icon: const Icon(Icons.add),
        label: const Text('Категория'),
      ),
      body: FutureBuilder<void>(
        future: _loadFuture,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          return RefreshIndicator(
            onRefresh: _load,
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                _buildTypeToggle(),
                const SizedBox(height: 12),
                _buildPeriodToggle(),
                const SizedBox(height: 16),
                _buildSummaryCard(),
                const SizedBox(height: 16),
                Text(
                  'Категории',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                _buildCategoryGrid(),
                const SizedBox(height: 24),
                Text(
                  '${_type.label == 'Расход' ? 'Траты' : 'Доходы'} $_periodTitle',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                ..._buildTransactionList(),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildTypeToggle() {
    return SegmentedButton<TransactionType>(
      segments: const [
        ButtonSegment(
          value: TransactionType.expense,
          label: Text('Траты'),
          icon: Icon(Icons.trending_down),
        ),
        ButtonSegment(
          value: TransactionType.income,
          label: Text('Доходы'),
          icon: Icon(Icons.trending_up),
        ),
      ],
      selected: {_type},
      onSelectionChanged: (selection) => setState(() => _type = selection.first),
    );
  }

  Widget _buildPeriodToggle() {
    return SegmentedButton<PeriodFilter>(
      segments: const [
        ButtonSegment(value: PeriodFilter.day, label: Text('День')),
        ButtonSegment(value: PeriodFilter.month, label: Text('Месяц')),
        ButtonSegment(value: PeriodFilter.year, label: Text('Год')),
      ],
      selected: {_period},
      onSelectionChanged: (selection) => setState(() => _period = selection.first),
    );
  }

  Widget _buildSummaryCard() {
    final color = _type == TransactionType.expense
        ? Theme.of(context).colorScheme.error
        : Theme.of(context).colorScheme.primary;
    return Card(
      elevation: 0,
      color: color.withValues(alpha: 0.1),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '${_type.label} $_periodTitle',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 4),
            if (_totalsByCurrency.isEmpty)
              Text(
                '0',
                style: Theme.of(
                  context,
                ).textTheme.headlineMedium?.copyWith(color: color),
              )
            else
              ..._totalsByCurrency.entries.map(
                (e) => Text(
                  formatAmount(e.value, e.key),
                  style: Theme.of(
                    context,
                  ).textTheme.headlineSmall?.copyWith(color: color),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildCategoryGrid() {
    if (_categories.isEmpty) {
      return const Text('Категорий пока нет. Создайте первую.');
    }
    final items = _categories.where((c) => c.type == _type).toList();
    if (items.isEmpty) {
      return Text('Нет категорий для «${_type.label}». Создайте новую.');
    }
    return GridView.count(
      crossAxisCount: 3,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 8,
      crossAxisSpacing: 8,
      childAspectRatio: 1.15,
      children: items.map((category) => _buildCategoryTile(category)).toList(),
    );
  }

  Widget _buildCategoryTile(Category category) {
    final color = Color(category.color ?? 0xFF90A4AE);
    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () => _addTransaction(category),
      child: Ink(
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.12),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            CircleAvatar(
              radius: 18,
              backgroundColor: color,
              child: const Icon(Icons.category, color: Colors.white, size: 20),
            ),
            const SizedBox(height: 6),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4),
              child: Text(
                category.name,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ),
          ],
        ),
      ),
    );
  }

  List<Widget> _buildTransactionList() {
    if (_filtered.isEmpty) {
      return [
        Card(
          elevation: 0,
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              children: [
                Icon(
                  _type == TransactionType.expense
                      ? Icons.receipt_long
                      : Icons.account_balance_wallet,
                  size: 40,
                  color: Theme.of(context).colorScheme.outline,
                ),
                const SizedBox(height: 8),
                Text(
                  'Операций за этот период нет',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ],
            ),
          ),
        ),
      ];
    }

    return _filtered.map((t) {
      final category = _categories
          .where((c) => c.id == t.categoryId)
          .firstOrNull;
      final color = category != null && category.color != null
          ? Color(category.color!)
          : Theme.of(context).colorScheme.outline;
      return Dismissible(
        key: ValueKey(t.id),
        direction: DismissDirection.endToStart,
        background: Container(
          alignment: Alignment.centerRight,
          padding: const EdgeInsets.symmetric(horizontal: 20),
          color: Theme.of(context).colorScheme.error,
          child: const Icon(Icons.delete, color: Colors.white),
        ),
        onDismissed: (_) => _removeTransaction(t),
        child: ListTile(
          contentPadding: const EdgeInsets.symmetric(horizontal: 8),
          leading: CircleAvatar(
            backgroundColor: color.withValues(alpha: 0.15),
            child: Icon(Icons.category, color: color),
          ),
          title: Text(category?.name ?? 'Без категории'),
          subtitle: Text(t.description ?? formatDay(t.date)),
          trailing: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                '${_type == TransactionType.expense ? '−' : '+'}'
                '${formatAmount(t.amount, t.currency)}',
                style: TextStyle(
                  color: _type == TransactionType.expense
                      ? Theme.of(context).colorScheme.error
                      : Theme.of(context).colorScheme.primary,
                  fontWeight: FontWeight.w600,
                ),
              ),
              Text(
                formatTime(t.date),
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ),
        ),
      );
    }).toList();
  }
}