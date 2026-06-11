import 'package:flutter/material.dart';

import '../../app/app_scope.dart';
import '../components/chat/conversation_list_panel.dart';
import '../components/chat/conversation_panel.dart';
import '../components/chat/message_input_panel.dart';
import '../components/chat/top_bar.dart';
import '../theme/app_theme.dart';
import '../theme/responsive.dart';

class ChatPage extends StatelessWidget {
  const ChatPage({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final width = MediaQuery.sizeOf(context).width;
    final mobile = AppBreakpoints.isMobile(width);

    final chatBody = Column(
      children: const [
        Expanded(child: ConversationPanel()),
        SizedBox(height: 8),
        MessageInputPanel(),
      ],
    );

    if (mobile) {
      return Column(
        children: [
          const ChatTopBar(mobile: true),
          const SizedBox(height: 8),
          Expanded(
            child: controller.sidebarOpen
                ? const ConversationListPanel(compact: true)
                : chatBody,
          ),
        ],
      );
    }

    return Row(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const ConversationListPanel(),
        const SizedBox(width: 12),
        Expanded(
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: AppColors.chatSurface,
              borderRadius: BorderRadius.circular(30),
            ),
            child: Padding(
              padding: const EdgeInsets.fromLTRB(18, 14, 18, 8),
              child: Column(
                children: [
                  const ChatTopBar(mobile: false),
                  const SizedBox(height: 12),
                  Expanded(child: chatBody),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}
