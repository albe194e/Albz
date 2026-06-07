import 'package:flutter/material.dart';

import '../../../app/app_page.dart';
import '../../../app/app_scope.dart';
import '../../../core/core_api.dart';
import '../../theme/app_theme.dart';
import '../base/app_button.dart';
import '../base/app_input.dart';
import '../base/app_panel.dart';
import '../base/profile_picture.dart';

class ConversationListPanel extends StatelessWidget {
  const ConversationListPanel({super.key, this.compact = false});

  final bool compact;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return AppPanel(
      width: compact ? null : 310,
      color: AppColors.panelMuted,
      header: Row(
        children: [
          Text('Albz', style: Theme.of(context).textTheme.headlineMedium),
          const Spacer(),
          IconButton(
            onPressed: () async {
              final request = await showDialog<_CreateConversationRequest>(
                context: context,
                builder: (_) =>
                    _CreateConversationDialog(friends: controller.friends),
              );
              if (request == null || !context.mounted) {
                return;
              }

              await controller.createConversation(
                name: request.name,
                participantUserIds: request.participantUserIds,
              );
            },
            icon: const Icon(Icons.add),
          ),
        ],
      ),
      footer: Row(
        children: [
          Expanded(
            child: AppButton.secondary(
              label: 'Profile',
              onPressed: () => controller.navigateTo(AppPage.profile),
            ),
          ),
          const SizedBox(width: 12),
          ProfilePicture(path: controller.currentUser?.profilePictureUrl ?? ''),
        ],
      ),
      child: controller.conversations.isEmpty
          ? const Center(
              child: Text(
                'No conversations yet',
                style: TextStyle(color: AppColors.textMuted),
              ),
            )
          : ListView.separated(
              itemCount: controller.conversations.length,
              separatorBuilder: (_, _) => const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final conversation = controller.conversations[index];
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

class _CreateConversationRequest {
  const _CreateConversationRequest({
    required this.name,
    required this.participantUserIds,
  });

  final String name;
  final List<String> participantUserIds;
}

class _CreateConversationDialog extends StatefulWidget {
  const _CreateConversationDialog({required this.friends});

  final List<CoreFriend> friends;

  @override
  State<_CreateConversationDialog> createState() =>
      _CreateConversationDialogState();
}

class _CreateConversationDialogState extends State<_CreateConversationDialog> {
  final TextEditingController _nameController = TextEditingController();
  final Set<String> _selectedFriendIds = <String>{};
  String _errorMessage = '';

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final hasFriends = widget.friends.isNotEmpty;
    final listHeight = widget.friends.length > 3
        ? 240.0
        : (widget.friends.length * 72.0).clamp(72.0, 240.0).toDouble();

    return AlertDialog(
      backgroundColor: AppColors.panel,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
      title: Text(
        'New Conversation',
        style: Theme.of(context).textTheme.titleLarge,
      ),
      content: SizedBox(
        width: 420,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Choose one or more friends to start a conversation.',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 16),
            AppInput(
              label: 'Conversation name',
              hint: 'Optional',
              controller: _nameController,
            ),
            const SizedBox(height: 16),
            if (hasFriends)
              SizedBox(
                height: listHeight,
                child: ListView.separated(
                  shrinkWrap: true,
                  itemCount: widget.friends.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 10),
                  itemBuilder: (context, index) {
                    final friend = widget.friends[index];
                    final selected = _selectedFriendIds.contains(friend.userId);

                    return Material(
                      color: selected
                          ? AppColors.accentSoft
                          : AppColors.cardMuted,
                      borderRadius: BorderRadius.circular(20),
                      child: InkWell(
                        onTap: () => _toggleFriend(friend.userId),
                        borderRadius: BorderRadius.circular(20),
                        child: Padding(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 14,
                            vertical: 10,
                          ),
                          child: Row(
                            children: [
                              Checkbox(
                                value: selected,
                                onChanged: (_) => _toggleFriend(friend.userId),
                              ),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Text(
                                      friend.name.isEmpty
                                          ? friend.userId
                                          : friend.name,
                                      style: Theme.of(
                                        context,
                                      ).textTheme.titleMedium,
                                    ),
                                    if (friend.username.isNotEmpty)
                                      Text(
                                        '@${friend.username}',
                                        style: Theme.of(
                                          context,
                                        ).textTheme.bodyMedium,
                                      ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    );
                  },
                ),
              )
            else
              const Text(
                'Add a friend before creating a conversation.',
                style: TextStyle(color: AppColors.textMuted),
              ),
            if (_errorMessage.isNotEmpty) ...[
              const SizedBox(height: 12),
              Text(
                _errorMessage,
                style: const TextStyle(color: AppColors.error),
              ),
            ],
          ],
        ),
      ),
      actions: [
        SizedBox(
          width: 132,
          child: AppButton.secondary(
            label: 'Cancel',
            onPressed: () => Navigator.of(context).pop(),
          ),
        ),
        SizedBox(
          width: 132,
          child: AppButton.primary(
            label: 'Create',
            onPressed: hasFriends ? _submit : null,
          ),
        ),
      ],
    );
  }

  void _toggleFriend(String userId) {
    setState(() {
      if (_selectedFriendIds.contains(userId)) {
        _selectedFriendIds.remove(userId);
      } else {
        _selectedFriendIds.add(userId);
      }
    });
  }

  void _submit() {
    if (_selectedFriendIds.isEmpty) {
      setState(() {
        _errorMessage = 'Select at least one friend.';
      });
      return;
    }

    Navigator.of(context).pop(
      _CreateConversationRequest(
        name: _nameController.text,
        participantUserIds: _selectedFriendIds.toList(growable: false),
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
      color: active ? AppColors.accent : AppColors.card,
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
