import 'package:flutter/material.dart';

import '../../app/app_page.dart';
import '../../app/app_scope.dart';
import '../components/base/app_button.dart';
import '../components/base/app_card.dart';
import '../theme/app_theme.dart';

class LandingPage extends StatelessWidget {
  const LandingPage({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return _CenteredPage(
      child: AppCard(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Albz', style: Theme.of(context).textTheme.headlineLarge),
            const SizedBox(height: 12),
            const Text(
              'A local-first chat that keeps your history on your own device.',
              style: TextStyle(color: AppColors.textSecondary, fontSize: 16),
            ),
            const SizedBox(height: 24),
            AppButton.primary(
              label: 'Login',
              onPressed: () => controller.navigateTo(AppPage.login),
            ),
            const SizedBox(height: 12),
            AppButton.secondary(
              label: 'Register',
              onPressed: () => controller.navigateTo(AppPage.register),
            ),
            const SizedBox(height: 18),
            Text(
              controller.statusMessage,
              style: const TextStyle(color: AppColors.textMuted),
            ),
          ],
        ),
      ),
    );
  }
}

class _CenteredPage extends StatelessWidget {
  const _CenteredPage({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520),
        child: Padding(padding: const EdgeInsets.all(24), child: child),
      ),
    );
  }
}
