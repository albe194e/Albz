import 'package:flutter/material.dart';

import 'app_nav_rail.dart';

class AppShell extends StatelessWidget {
  const AppShell({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;

    if (width < 900) {
      return child;
    }

    return Row(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const AppNavRail(),
        const SizedBox(width: 12),
        Expanded(child: child),
      ],
    );
  }
}
