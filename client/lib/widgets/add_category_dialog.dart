import 'package:flutter/material.dart';

import '../models/category.dart';
import '../models/transaction_type.dart';

/// Диалог создания новой категории.
Future<Category?> showAddCategoryDialog(
  BuildContext context,
  TransactionType initialType,
) {
  return showDialog<Category>(
    context: context,
    builder: (_) => _AddCategoryDialog(initialType: initialType),
  );
}

class _AddCategoryDialog extends StatefulWidget {
  const _AddCategoryDialog({required this.initialType});

  final TransactionType initialType;

  @override
  State<_AddCategoryDialog> createState() => _AddCategoryDialogState();
}

class _AddCategoryDialogState extends State<_AddCategoryDialog> {
  final _nameController = TextEditingController();
  late TransactionType _type;
  String _currency = 'RUB';
  int _color = 0xFF66BB6A;

  static const _currencies = [
    'RUB',
    'USD',
    'EUR',
    'BYN',
    'KZT',
    'UAH',
  ];

  static const _colors = [
    0xFF66BB6A,
    0xFF42A5F5,
    0xFFAB47BC,
    0xFFEF5350,
    0xFFFFA726,
    0xFF26A69A,
    0xFF5C6BC0,
    0xFF8D6E63,
  ];

  @override
  void initState() {
    super.initState();
    _type = widget.initialType;
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  void _submit() {
    final name = _nameController.text.trim();
    if (name.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Укажите название категории')),
      );
      return;
    }
    Navigator.of(context).pop(
      Category(
        id: DateTime.now().microsecondsSinceEpoch.toString(),
        name: name,
        type: _type,
        currency: _currency,
        color: _color,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Новая категория'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TextField(
              controller: _nameController,
              autofocus: true,
              decoration: const InputDecoration(
                labelText: 'Название',
                hintText: 'Например, Продукты',
              ),
            ),
            const SizedBox(height: 16),
            SegmentedButton<TransactionType>(
              segments: const [
                ButtonSegment(
                  value: TransactionType.expense,
                  label: Text('Расход'),
                ),
                ButtonSegment(
                  value: TransactionType.income,
                  label: Text('Доход'),
                ),
              ],
              selected: {_type},
              onSelectionChanged: (selection) {
                setState(() => _type = selection.first);
              },
            ),
            const SizedBox(height: 16),
            DropdownButtonFormField<String>(
              initialValue: _currency,
              decoration: const InputDecoration(labelText: 'Валюта'),
              items: _currencies
                  .map((c) => DropdownMenuItem(value: c, child: Text(c)))
                  .toList(),
              onChanged: (value) {
                if (value != null) setState(() => _currency = value);
              },
            ),
            const SizedBox(height: 16),
            const Text('Цвет'),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              children: _colors.map((color) {
                final selected = _color == color;
                return GestureDetector(
                  onTap: () => setState(() => _color = color),
                  child: Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      color: Color(color),
                      shape: BoxShape.circle,
                      border: selected
                          ? Border.all(color: Colors.black, width: 2)
                          : null,
                    ),
                  ),
                );
              }).toList(),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Отмена'),
        ),
        FilledButton(
          onPressed: _submit,
          child: const Text('Создать'),
        ),
      ],
    );
  }
}