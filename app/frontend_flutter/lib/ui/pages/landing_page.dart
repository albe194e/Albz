import 'package:flutter/material.dart';

import '../../app/app_page.dart';
import '../../app/app_scope.dart';
import '../components/base/app_button.dart';
import '../components/base/app_card.dart';
import '../theme/app_theme.dart';
import '../theme/responsive.dart';

class LandingPage extends StatelessWidget {
  const LandingPage({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return LayoutBuilder(
      builder: (context, constraints) {
        final width = constraints.maxWidth;
        final mobile = AppBreakpoints.isMobile(width);
        final horizontalPadding = mobile ? 16.0 : 24.0;
        final verticalPadding = mobile ? 16.0 : 24.0;
        final bottomInset = MediaQuery.viewInsetsOf(context).bottom;
        final minHeight =
            constraints.maxHeight - (verticalPadding * 2) - bottomInset;

        return SingleChildScrollView(
          padding: EdgeInsets.fromLTRB(
            horizontalPadding,
            verticalPadding,
            horizontalPadding,
            verticalPadding + bottomInset,
          ),
          child: ConstrainedBox(
            constraints: BoxConstraints(
              minHeight: minHeight > 0 ? minHeight : 0,
            ),
            child: Center(
              child: ConstrainedBox(
                constraints: BoxConstraints(maxWidth: mobile ? 420 : 520),
                child: AppCard(
                  padding: EdgeInsets.all(mobile ? 18 : 20),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Albz',
                        style: Theme.of(context).textTheme.headlineLarge,
                      ),
                      const SizedBox(height: 24),
                      AppButton.primary(
                        label: 'Login',
                        onPressed: () => controller.navigateTo(AppPage.login),
                      ),
                      const SizedBox(height: 12),
                      AppButton.secondary(
                        label: 'Register',
                        onPressed: () =>
                            controller.navigateTo(AppPage.register),
                      ),
                      const SizedBox(height: 18),
                      Text(
                        controller.statusMessage,
                        style: const TextStyle(color: AppColors.textMuted),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}
