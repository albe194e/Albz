import 'package:flutter/material.dart';

import '../../../app/app_scope.dart';

class MobilePageHeader extends StatelessWidget {
  const MobilePageHeader({super.key, required this.title});

  final String title;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Row(
      children: [
        IconButton(
          onPressed: controller.toggleMobileNav,
          icon: const Icon(Icons.menu_rounded),
        ),
        Expanded(
          child: Text(title, style: Theme.of(context).textTheme.headlineMedium),
        ),
      ],
    );
  }
}
