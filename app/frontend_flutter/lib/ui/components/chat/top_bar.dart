import 'package:flutter/material.dart';

import '../../../app/app_scope.dart';

class ChatTopBar extends StatelessWidget {
  const ChatTopBar({super.key, required this.mobile});

  final bool mobile;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 16),
      decoration: BoxDecoration(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(mobile ? 18 : 24),
      ),
      child: Row(
        children: [
          if (mobile)
            IconButton(
              onPressed: controller.toggleSidebar,
              icon: const Icon(Icons.menu_rounded),
            ),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  controller.activeConversationName,
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 4),
                Text(
                  controller.currentUser == null
                      ? 'No active session'
                      : 'Signed in as @${controller.currentUser!.username}',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ],
            ),
          ),
          IconButton(
            onPressed: controller.refreshSnapshot,
            icon: const Icon(Icons.refresh_rounded),
          ),
        ],
      ),
    );
  }
}
