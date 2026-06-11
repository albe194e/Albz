import 'package:flutter/material.dart';

import '../../../app/app_scope.dart';
import '../../theme/responsive.dart';
import 'app_nav_rail.dart';
import 'mobile_app_drawer.dart';

class AppShell extends StatelessWidget {
  const AppShell({super.key, required this.child, this.navRailOverlapTop = 0});

  final Widget child;
  final double navRailOverlapTop;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final width = MediaQuery.sizeOf(context).width;

    if (!AppBreakpoints.showNavRail(width)) {
      return Stack(
        children: [
          child,
          if (controller.mobileNavOpen) const MobileAppDrawer(),
        ],
      );
    }

    return Stack(
      clipBehavior: Clip.none,
      children: [
        Positioned(
          left: 0,
          top: -navRailOverlapTop,
          bottom: 0,
          child: const AppNavRail(),
        ),
        Positioned.fill(left: 90, child: child),
      ],
    );
  }
}
