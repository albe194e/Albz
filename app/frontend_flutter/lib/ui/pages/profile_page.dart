import 'package:flutter/material.dart';

import '../../app/app_page.dart';
import '../../app/app_scope.dart';
import '../../core/core_api.dart';
import '../components/base/app_button.dart';
import '../components/base/app_card.dart';
import '../components/base/app_input.dart';
import '../theme/app_theme.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({super.key});

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  final TextEditingController _friendCodeController = TextEditingController(
    text: 'ALBZ-',
  );

  @override
  void dispose() {
    _friendCodeController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Padding(
      padding: const EdgeInsets.all(24),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 820),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              AppCard(
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 18,
                ),
                child: Row(
                  children: [
                    SizedBox(
                      width: 140,
                      child: AppButton.secondary(
                        label: 'Back',
                        onPressed: () => controller.navigateTo(AppPage.chat),
                      ),
                    ),
                    const SizedBox(width: 16),
                    Text(
                      'Profile',
                      style: Theme.of(context).textTheme.headlineMedium,
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 18),
              Expanded(
                child: SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      AppCard(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              'Share this code so another person can add you.',
                              style: TextStyle(color: AppColors.textSecondary),
                            ),
                            const SizedBox(height: 12),
                            Text(
                              controller.currentUser?.friendCode.isNotEmpty ==
                                      true
                                  ? 'Your friend code: ${controller.currentUser!.friendCode}'
                                  : 'Friend code unavailable',
                              style: Theme.of(context).textTheme.titleLarge,
                            ),
                            const SizedBox(height: 16),
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.end,
                              children: [
                                Expanded(
                                  child: AppInput(
                                    label: 'Send friend request',
                                    hint: 'Enter a friend\'s code',
                                    controller: _friendCodeController,
                                  ),
                                ),
                                const SizedBox(width: 12),
                                SizedBox(
                                  width: 160,
                                  child: AppButton.primary(
                                    label: 'Send Request',
                                    onPressed: () async {
                                      await controller.sendFriendRequest(
                                        _friendCodeController.text,
                                      );
                                      if (mounted &&
                                          controller.errorMessage.isEmpty) {
                                        _friendCodeController.text = 'ALBZ-';
                                      }
                                    },
                                  ),
                                ),
                              ],
                            ),
                            if (controller.errorMessage.isNotEmpty) ...[
                              const SizedBox(height: 12),
                              Text(
                                controller.errorMessage,
                                style: const TextStyle(color: AppColors.error),
                              ),
                            ],
                            if (controller.infoMessage.isNotEmpty) ...[
                              const SizedBox(height: 12),
                              Text(
                                controller.infoMessage,
                                style: const TextStyle(
                                  color: AppColors.textMuted,
                                ),
                              ),
                            ],
                          ],
                        ),
                      ),
                      const SizedBox(height: 18),
                      _SectionCard(
                        title: 'Pending Requests',
                        child: controller.friendRequests.isEmpty
                            ? const Text(
                                'No pending friend requests',
                                style: TextStyle(color: AppColors.textMuted),
                              )
                            : Column(
                                children: controller.friendRequests
                                    .map(
                                      (request) =>
                                          _FriendRequestRow(request: request),
                                    )
                                    .toList(growable: false),
                              ),
                      ),
                      const SizedBox(height: 18),
                      _SectionCard(
                        title: 'Friends',
                        child: controller.friends.isEmpty
                            ? const Text(
                                'No friends added yet',
                                style: TextStyle(color: AppColors.textMuted),
                              )
                            : Column(
                                children: controller.friends
                                    .map((friend) => _FriendRow(friend: friend))
                                    .toList(growable: false),
                              ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 18),
              AppCard(
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 18,
                ),
                child: Row(
                  children: [
                    const Spacer(),
                    SizedBox(
                      width: 140,
                      child: AppButton.secondary(
                        label: 'Logout',
                        onPressed: controller.isAuthInFlight
                            ? null
                            : () => controller.logout(),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _SectionCard extends StatelessWidget {
  const _SectionCard({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 12),
          child,
        ],
      ),
    );
  }
}

class _FriendRequestRow extends StatelessWidget {
  const _FriendRequestRow({required this.request});

  final CoreFriendRequest request;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  request.name,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                if (request.username.isNotEmpty)
                  Text(
                    '@${request.username}',
                    style: const TextStyle(color: AppColors.textSecondary),
                  ),
              ],
            ),
          ),
          SizedBox(
            width: 110,
            child: AppButton.primary(
              label: 'Accept',
              onPressed: () =>
                  controller.acceptFriendRequest(request.fromUserId),
            ),
          ),
          const SizedBox(width: 10),
          SizedBox(
            width: 110,
            child: AppButton.secondary(
              label: 'Reject',
              onPressed: () =>
                  controller.rejectFriendRequest(request.fromUserId),
            ),
          ),
        ],
      ),
    );
  }
}

class _FriendRow extends StatelessWidget {
  const _FriendRow({required this.friend});

  final CoreFriend friend;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  friend.name,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                if (friend.username.isNotEmpty)
                  Text(
                    '@${friend.username}',
                    style: const TextStyle(color: AppColors.textSecondary),
                  ),
              ],
            ),
          ),
          SizedBox(
            width: 140,
            child: AppButton.primary(
              label: 'Start Chat',
              onPressed: () =>
                  controller.startDirectConversation(friend.userId),
            ),
          ),
        ],
      ),
    );
  }
}
