import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';

class Section extends StatelessWidget {
  const Section({
    super.key,
    required this.title,
    required this.child,
    this.padding = const EdgeInsets.all(20),
    this.backgroundColor = Colors.transparent,
    this.showBorder = false,
  });

  final String title;
  final Widget child;
  final EdgeInsetsGeometry padding;
  final Color backgroundColor;
  final bool showBorder;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(24),
        border: showBorder
            ? Border.all(color: AppColors.borderSoft.withValues(alpha: 0.65))
            : null,
      ),
      child: Padding(
        padding: padding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                color: AppColors.textSecondary,
                letterSpacing: 0.2,
              ),
            ),
            const SizedBox(height: 14),
            child,
          ],
        ),
      ),
    );
  }
}
