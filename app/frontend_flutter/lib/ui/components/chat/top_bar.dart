import 'package:flutter/material.dart';

import '../../../app/app_scope.dart';
import '../../theme/responsive.dart';

class ChatTopBar extends StatelessWidget {
  const ChatTopBar({super.key, required this.mobile});

  final bool mobile;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final width = MediaQuery.sizeOf(context).width;
    final narrow = AppBreakpoints.isNarrow(width);
    final showingConversationList = mobile && controller.sidebarOpen;

    return SizedBox(
      height: narrow ? 64 : 68,
      child: Padding(
        padding: EdgeInsets.symmetric(horizontal: narrow ? 14 : 18),
        child: Row(
          children: [
            if (mobile)
              IconButton(
                onPressed: showingConversationList
                    ? controller.toggleMobileNav
                    : controller.showConversationList,
                icon: Icon(
                  showingConversationList
                      ? Icons.menu_rounded
                      : Icons.arrow_back_rounded,
                ),
              ),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    showingConversationList
                        ? 'Chats'
                        : controller.activeConversationName,
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      fontSize: narrow ? 18 : null,
                    ),
                  ),
                  if (controller.currentUser == null) ...[
                    const SizedBox(height: 4),
                    Text(
                      'No active session',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ],
                ],
              ),
            ),
            IconButton(
              onPressed: controller.refreshSnapshot,
              icon: const Icon(Icons.refresh_rounded),
            ),
          ],
        ),
      ),
    );
  }
}
