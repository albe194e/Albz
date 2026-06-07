import 'package:flutter/material.dart';

import '../../../app/app_scope.dart';
import '../base/app_button.dart';
import '../base/app_input.dart';

class MessageInputPanel extends StatefulWidget {
  const MessageInputPanel({super.key});

  @override
  State<MessageInputPanel> createState() => _MessageInputPanelState();
}

class _MessageInputPanelState extends State<MessageInputPanel> {
  final TextEditingController _messageController = TextEditingController();

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(24),
      ),
      child: Row(
        children: [
          Expanded(
            child: AppInput(
              label: 'Message',
              hint: 'Write your message',
              controller: _messageController,
            ),
          ),
          const SizedBox(width: 12),
          SizedBox(
            width: 120,
            child: AppButton.primary(
              label: 'Send',
              onPressed: () async {
                final success = await controller.sendMessage(
                  _messageController.text,
                );
                if (success) {
                  _messageController.clear();
                }
              },
            ),
          ),
        ],
      ),
    );
  }
}
