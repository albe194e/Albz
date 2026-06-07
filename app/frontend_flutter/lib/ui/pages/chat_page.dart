import 'package:flutter/material.dart';

import '../../app/app_scope.dart';
import '../components/chat/conversation_list_panel.dart';
import '../components/chat/conversation_panel.dart';
import '../components/chat/message_input_panel.dart';
import '../components/chat/top_bar.dart';
import '../theme/app_theme.dart';

class ChatPage extends StatelessWidget {
  const ChatPage({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final width = MediaQuery.sizeOf(context).width;
    final mobile = width < 760;

    final chatBody = Column(
      children: const [
        Expanded(child: ConversationPanel()),
        SizedBox(height: 14),
        MessageInputPanel(),
      ],
    );

    if (mobile) {
      return Padding(
        padding: const EdgeInsets.all(10),
        child: Stack(
          children: [
            Column(
              children: [
                const ChatTopBar(mobile: true),
                const SizedBox(height: 12),
                Expanded(child: chatBody),
              ],
            ),
            if (controller.sidebarOpen)
              Row(
                children: [
                  Expanded(
                    child: GestureDetector(
                      onTap: controller.closeSidebar,
                      child: Container(color: Colors.black54),
                    ),
                  ),
                  const SizedBox(
                    width: 320,
                    child: ConversationListPanel(compact: true),
                  ),
                ],
              ),
          ],
        ),
      );
    }

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          const ConversationListPanel(),
          const SizedBox(width: 12),
          Expanded(
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: AppColors.appShell,
                borderRadius: BorderRadius.circular(30),
              ),
              child: Padding(
                padding: const EdgeInsets.all(18),
                child: Column(
                  children: [
                    const ChatTopBar(mobile: false),
                    const SizedBox(height: 16),
                    Expanded(child: chatBody),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
