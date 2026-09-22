import 'package:flutter/material.dart';

import '../models/category.dart';
import '../models/transaction.dart';

/// Диалог добавления операции в выбранную категорию.
Future<Transaction?> showAddTransactionDialog(
  BuildContext context, {
  required Category category,
}) {
  return showDialog<Transaction>(
    context: context,
    builder: (_) => _AddTransactionDialog(category: category),
  );
}

class _AddTransactionDialog extends StatefulWidget {
  const _AddTransactionDialog({required this.category});

  final Category category;

  @override
  State<_AddTransactionDialog> createState() => _AddTransactionDialogState();
}

class _AddTransactionDialogState extends State<_AddTransactionDialog> {
  final _amountController = TextEditingController();
  final _descriptionController = TextEditingController();
  late DateTime _date;
  late String _currency;

  @override
  void initState() {
    super.initState();
    _date = DateTime.now();
    _currency = widget.category.currency;
  }

  @override
  void dispose() {
    _amountController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _date,
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );
    if (picked != null) setState(() => _date = picked);
  }

  void _submit() {
    final amount = double.tryParse(
      _amountController.text.replaceAll(',', '.'),
    );
    if (amount == null || amount <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Укажите корректную сумму')),
      );
      return;
    }
    Navigator.of(context).pop(
      Transaction(
        id: DateTime.now().microsecondsSinceEpoch.toString(),
        categoryId: widget.category.id,
        type: widget.category.type,
        amount: amount,
        date: _date,
        currency: _currency,
        description:
            _descriptionController.text.trim().isEmpty
                ? null
                : _descriptionController.text.trim(),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('${widget.category.type.label}: ${widget.category.name}'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: _amountController,
              autofocus: true,
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              decoration: const InputDecoration(
                labelText: 'Сумма',
                prefixText: ' ',
              ),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: _currency,
              decoration: const InputDecoration(labelText: 'Валюта'),
              items: const [
                DropdownMenuItem(value: 'RUB', child: Text('RUB — ₽')),
                DropdownMenuItem(value: 'USD', child: Text('USD — \$')),
                DropdownMenuItem(value: 'EUR', child: Text('EUR — €')),
                DropdownMenuItem(value: 'BYN', child: Text('BYN')),
                DropdownMenuItem(value: 'KZT', child: Text('KZT')),
              ],
              onChanged: (value) {
                if (value != null) setState(() => _currency = value);
              },
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _descriptionController,
              decoration: const InputDecoration(
                labelText: 'Описание (необязательно)',
              ),
            ),
            const SizedBox(height: 12),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.calendar_today),
              title: Text(
                '${_date.day}.${_date.month}.${_date.year}',
              ),
              onTap: _pickDate,
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
          child: const Text('Сохранить'),
        ),
      ],
    );
  }
}