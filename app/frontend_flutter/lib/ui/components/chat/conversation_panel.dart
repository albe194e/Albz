import 'package:flutter/material.dart';

import '../../../app/app_scope.dart';
import '../../../core/core_api.dart';
import '../../theme/app_theme.dart';
import '../../theme/responsive.dart';
import '../base/app_panel.dart';

class ConversationPanel extends StatelessWidget {
  const ConversationPanel({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final messages = controller.messages;
    final width = MediaQuery.sizeOf(context).width;
    final mobile = AppBreakpoints.isMobile(width);

    return AppPanel(
      color: Colors.transparent,
      radius: 0,
      padding: EdgeInsets.zero,
      child: messages.isEmpty
          ? Center(
              child: Text(
                controller.activeConversationId.isEmpty
                    ? 'Choose a conversation'
                    : 'No messages in this conversation yet',
                style: const TextStyle(color: AppColors.textMuted),
              ),
            )
          : ListView.separated(
              padding: EdgeInsets.symmetric(
                horizontal: mobile ? 12 : 0,
                vertical: 4,
              ),
              itemCount: messages.length,
              separatorBuilder: (_, _) => const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final message = messages[index];
                final isOwnMessage =
                    message.senderId == controller.currentUser?.id;
                return _MessageBubble(message: message, alignEnd: isOwnMessage);
              },
            ),
    );
  }
}

class _MessageBubble extends StatelessWidget {
  const _MessageBubble({required this.message, required this.alignEnd});

  final CoreMessage message;
  final bool alignEnd;

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    final mobile = AppBreakpoints.isMobile(width);
    final maxBubbleWidth = mobile ? width - 48 : 520.0;

    return Align(
      alignment: alignEnd ? Alignment.centerRight : Alignment.centerLeft,
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: maxBubbleWidth),
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: alignEnd ? AppColors.accent : AppColors.card,
            borderRadius: BorderRadius.circular(24),
          ),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 14),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (!alignEnd)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: Text(
                      message.senderId,
                      style: const TextStyle(
                        color: AppColors.textSecondary,
                        fontSize: 12,
                      ),
                    ),
                  ),
                Text(
                  message.body.isEmpty ? '(empty message)' : message.body,
                  style: const TextStyle(color: AppColors.textPrimary),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
