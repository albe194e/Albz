import 'package:flutter/material.dart';

import '../../../app/app_page.dart';
import '../../../app/app_scope.dart';
import '../../../core/core_api.dart';
import '../../theme/app_theme.dart';
import '../base/app_button.dart';
import '../base/app_panel.dart';
import '../base/profile_picture.dart';

class ConversationListPanel extends StatefulWidget {
  const ConversationListPanel({super.key, this.compact = false});

  final bool compact;

  @override
  State<ConversationListPanel> createState() => _ConversationListPanelState();
}

class _ConversationListPanelState extends State<ConversationListPanel> {
  final TextEditingController _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final query = _searchController.text.trim().toLowerCase();
    final conversations = controller.conversations
        .where((conversation) {
          final name = conversation.name.isEmpty
              ? 'Unnamed conversation'
              : conversation.name;
          return name.toLowerCase().contains(query);
        })
        .toList(growable: false);

    return AppPanel(
      width: widget.compact ? null : 252,
      color: AppColors.chatSurface,
      showHeaderDivider: false,
      header: SizedBox(
        height: 68,
        child: Align(
          alignment: Alignment.centerLeft,
          child: TextField(
            controller: _searchController,
            onChanged: (_) => setState(() {}),
            decoration: const InputDecoration(
              hintText: 'Search',
              prefixIcon: Icon(Icons.search_rounded),
            ),
          ),
        ),
      ),
      footer: widget.compact
          ? Row(
              children: [
                Expanded(
                  child: AppButton.secondary(
                    label: 'Profile',
                    onPressed: () => controller.navigateTo(AppPage.profile),
                  ),
                ),
                const SizedBox(width: 12),
                ProfilePicture(
                  path: controller.currentUser?.profilePictureUrl ?? '',
                ),
              ],
            )
          : null,
      child: conversations.isEmpty
          ? const Center(
              child: Text(
                'No matching conversations',
                style: TextStyle(color: AppColors.textMuted),
              ),
            )
          : ListView.separated(
              itemCount: conversations.length,
              separatorBuilder: (_, _) => const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final conversation = conversations[index];
                return _ConversationCard(
                  conversation: conversation,
                  active: controller.activeConversationId == conversation.id,
                  onTap: () => controller.openConversation(conversation.id),
                );
              },
            ),
    );
  }
}

class _ConversationCard extends StatelessWidget {
  const _ConversationCard({
    required this.conversation,
    required this.active,
    required this.onTap,
  });

  final CoreConversation conversation;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: active ? AppColors.accent : AppColors.cardMuted,
      borderRadius: BorderRadius.circular(24),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(24),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 22, vertical: 18),
          child: Align(
            alignment: Alignment.centerLeft,
            child: Text(
              conversation.name.isEmpty
                  ? 'Unnamed conversation'
                  : conversation.name,
              style: Theme.of(context).textTheme.titleMedium,
            ),
          ),
        ),
      ),
    );
  }
}
