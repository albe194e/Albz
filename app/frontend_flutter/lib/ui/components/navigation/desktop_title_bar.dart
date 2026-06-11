import 'package:flutter/material.dart';

import '../../../platform/desktop_window.dart';
import '../../theme/app_theme.dart';

class DesktopTitleBar extends StatefulWidget {
  const DesktopTitleBar({super.key});

  @override
  State<DesktopTitleBar> createState() => _DesktopTitleBarState();
}

class _DesktopTitleBarState extends State<DesktopTitleBar> {
  bool _maximized = false;

  @override
  void initState() {
    super.initState();
    _refreshWindowState();
  }

  Future<void> _refreshWindowState() async {
    final maximized = await DesktopWindow.isMaximized();
    if (!mounted) {
      return;
    }
    setState(() => _maximized = maximized);
  }

  Future<void> _toggleMaximize() async {
    await DesktopWindow.toggleMaximize();
    await _refreshWindowState();
  }

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.navRail,
      child: SizedBox(
        height: 42,
        child: Row(
          children: [
            Expanded(
              child: GestureDetector(
                behavior: HitTestBehavior.opaque,
                onPanStart: (_) => DesktopWindow.startDrag(),
                onDoubleTap: _toggleMaximize,
                child: const SizedBox.expand(),
              ),
            ),
            _TitleBarButton(
              icon: Icons.remove_rounded,
              onPressed: DesktopWindow.minimize,
            ),
            _TitleBarButton(
              icon: _maximized
                  ? Icons.filter_none_rounded
                  : Icons.crop_square_rounded,
              onPressed: _toggleMaximize,
            ),
            _TitleBarButton(
              icon: Icons.close_rounded,
              hoverColor: const Color(0xFFD35A5A),
              onPressed: DesktopWindow.close,
            ),
          ],
        ),
      ),
    );
  }
}

class _TitleBarButton extends StatefulWidget {
  const _TitleBarButton({
    required this.icon,
    required this.onPressed,
    this.hoverColor,
  });

  final IconData icon;
  final Future<void> Function() onPressed;
  final Color? hoverColor;

  @override
  State<_TitleBarButton> createState() => _TitleBarButtonState();
}

class _TitleBarButtonState extends State<_TitleBarButton> {
  bool _hovering = false;

  @override
  Widget build(BuildContext context) {
    final hoverColor = widget.hoverColor ?? AppColors.navRailHover;

    return MouseRegion(
      onEnter: (_) => setState(() => _hovering = true),
      onExit: (_) => setState(() => _hovering = false),
      child: GestureDetector(
        behavior: HitTestBehavior.opaque,
        onTap: () => widget.onPressed(),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 120),
          width: 48,
          height: double.infinity,
          color: _hovering ? hoverColor : Colors.transparent,
          child: Icon(widget.icon, size: 18, color: AppColors.textPrimary),
        ),
      ),
    );
  }
}
