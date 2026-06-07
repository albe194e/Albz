import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';

class AppPanel extends StatelessWidget {
  const AppPanel({
    super.key,
    required this.child,
    this.header,
    this.footer,
    this.showHeaderDivider = true,
    this.showFooterDivider = true,
    this.padding = const EdgeInsets.all(18),
    this.color = AppColors.panel,
    this.radius = 24,
    this.width,
  });

  final Widget child;
  final Widget? header;
  final Widget? footer;
  final bool showHeaderDivider;
  final bool showFooterDivider;
  final EdgeInsetsGeometry padding;
  final Color color;
  final double radius;
  final double? width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(radius),
      ),
      child: Column(
        children: [
          if (header != null) ...[
            Padding(
              padding: const EdgeInsets.fromLTRB(18, 18, 18, 0),
              child: header!,
            ),
            if (showHeaderDivider) ...[
              const SizedBox(height: 14),
              const Divider(height: 1),
            ],
          ],
          Expanded(
            child: Padding(padding: padding, child: child),
          ),
          if (footer != null) ...[
            if (showFooterDivider) const Divider(height: 1),
            Padding(padding: const EdgeInsets.all(18), child: footer!),
          ],
        ],
      ),
    );
  }
}
