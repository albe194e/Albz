import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

import '../../../app/app_scope.dart';
import '../../theme/app_theme.dart';
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

    Future<void> sendCurrentMessage() async {
      final success = await controller.sendMessage(_messageController.text);
      if (success) {
        _messageController.clear();
      }
    }

    return Container(
      padding: const EdgeInsets.fromLTRB(0, 6, 0, 0),
      decoration: BoxDecoration(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(24),
      ),
      child: AppInput(
        label: 'Message',
        hint: 'Write your message',
        controller: _messageController,
        suffixIcon: Padding(
          padding: const EdgeInsets.only(right: 6),
          child: IconButton(
            onPressed: sendCurrentMessage,
            icon: SvgPicture.asset(
              'assets/icons/send.svg',
              width: 18,
              height: 18,
              colorFilter: const ColorFilter.mode(
                AppColors.textPrimary,
                BlendMode.srcIn,
              ),
            ),
            splashRadius: 18,
            tooltip: 'Send',
          ),
        ),
      ),
    );
  }
}
