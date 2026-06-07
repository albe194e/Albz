import 'package:flutter/material.dart';

import '../../../app/app_scope.dart';

class ChatTopBar extends StatelessWidget {
  const ChatTopBar({super.key, required this.mobile});

  final bool mobile;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return SizedBox(
      height: 68,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 18),
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
                mainAxisAlignment: MainAxisAlignment.center,
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
      ),
    );
  }
}
