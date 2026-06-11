import 'package:flutter/material.dart';

import '../../app/app_scope.dart';
import '../components/base/app_button.dart';
import '../components/base/mobile_page_header.dart';
import '../components/base/section.dart';
import '../theme/app_theme.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final currentUser = controller.currentUser;
    final mobile = MediaQuery.sizeOf(context).width < 760;

    return Padding(
      padding: EdgeInsets.all(mobile ? 16 : 24),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 820),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              mobile
                  ? const MobilePageHeader(title: 'Profile')
                  : Text(
                      'Profile',
                      style: Theme.of(context).textTheme.headlineMedium,
                    ),
              const SizedBox(height: 20),
              Expanded(
                child: SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Section(
                        title: 'Account',
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              currentUser?.name.isNotEmpty == true
                                  ? currentUser!.name
                                  : 'Profile details unavailable',
                              style: Theme.of(context).textTheme.headlineSmall,
                            ),
                            if (currentUser?.username.isNotEmpty == true) ...[
                              const SizedBox(height: 8),
                              Text(
                                '@${currentUser!.username}',
                                style: const TextStyle(
                                  color: AppColors.textSecondary,
                                ),
                              ),
                            ],
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 18),
              Align(
                alignment: Alignment.centerRight,
                child: SizedBox(
                  width: mobile ? double.infinity : 140,
                  child: AppButton.secondary(
                    label: 'Logout',
                    onPressed: controller.isAuthInFlight
                        ? null
                        : () => controller.logout(),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
