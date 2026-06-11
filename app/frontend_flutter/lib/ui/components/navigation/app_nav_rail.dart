import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

import '../../../app/app_page.dart';
import '../../../app/app_scope.dart';
import '../../theme/app_theme.dart';
import '../base/profile_picture.dart';

class AppNavRail extends StatelessWidget {
  const AppNavRail({super.key});

  static const _chatIconAsset = 'assets/icons/chat.svg';
  static const _contactsIconAsset = 'assets/icons/contacts.svg';
  static const _logoAsset = 'assets/icons/icon.svg';

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final page = controller.page;

    return SizedBox(
      width: 78,
      child: DecoratedBox(
        decoration: const BoxDecoration(color: AppColors.navRail),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 18),
          child: Column(
            children: [
              const _RailLogo(),
              const SizedBox(height: 18),
              _RailButton(
                selected: page == AppPage.chat,
                onPressed: () => controller.navigateTo(AppPage.chat),
                child: SvgPicture.asset(
                  _chatIconAsset,
                  width: 20,
                  height: 20,
                  colorFilter: ColorFilter.mode(
                    page == AppPage.chat
                        ? AppColors.textPrimary
                        : AppColors.textSecondary,
                    BlendMode.srcIn,
                  ),
                ),
              ),
              const SizedBox(height: 10),
              _RailButton(
                selected: page == AppPage.contacts,
                onPressed: () => controller.navigateTo(AppPage.contacts),
                child: SvgPicture.asset(
                  _contactsIconAsset,
                  width: 20,
                  height: 20,
                  colorFilter: ColorFilter.mode(
                    page == AppPage.contacts
                        ? AppColors.textPrimary
                        : AppColors.textSecondary,
                    BlendMode.srcIn,
                  ),
                ),
              ),
              const SizedBox(height: 14),
              const _RailSeparator(),
              const Spacer(),
              _RailButton(
                selected: page == AppPage.profile,
                onPressed: () => controller.navigateTo(AppPage.profile),
                child: ProfilePicture(
                  path: controller.currentUser?.profilePictureUrl ?? '',
                  size: 28,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _RailSeparator extends StatelessWidget {
  const _RailSeparator();

  @override
  Widget build(BuildContext context) {
    return FractionallySizedBox(
      widthFactor: 0.7,
      child: Container(height: 1, color: AppColors.borderSoft),
    );
  }
}

class _RailLogo extends StatelessWidget {
  const _RailLogo();

  @override
  Widget build(BuildContext context) {
    return const SizedBox(
      width: 54,
      height: 54,
      child: Padding(padding: EdgeInsets.all(6), child: _RailLogoImage()),
    );
  }
}

class _RailLogoImage extends StatelessWidget {
  const _RailLogoImage();

  @override
  Widget build(BuildContext context) {
    return SvgPicture.asset(
      AppNavRail._logoAsset,
      fit: BoxFit.contain,
      width: 42,
      height: 42,
      errorBuilder: (context, error, stackTrace) {
        return Center(
          child: Text(
            'AL',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              fontSize: 18,
              fontWeight: FontWeight.w700,
              letterSpacing: 0.8,
            ),
          ),
        );
      },
    );
  }
}

class _RailButton extends StatelessWidget {
  const _RailButton({
    required this.selected,
    required this.onPressed,
    required this.child,
  });

  final bool selected;
  final VoidCallback onPressed;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return _RailHoverTile(
      onTap: onPressed,
      child: Center(
        child: IconTheme(
          data: IconThemeData(
            color: selected ? AppColors.textPrimary : AppColors.textSecondary,
          ),
          child: child,
        ),
      ),
    );
  }
}

class _RailHoverTile extends StatefulWidget {
  const _RailHoverTile({required this.child, this.onTap});

  final Widget child;
  final VoidCallback? onTap;

  @override
  State<_RailHoverTile> createState() => _RailHoverTileState();
}

class _RailHoverTileState extends State<_RailHoverTile> {
  bool _hovering = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _hovering = true),
      onExit: (_) => setState(() => _hovering = false),
      child: GestureDetector(
        behavior: HitTestBehavior.opaque,
        onTap: widget.onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 120),
          width: 46,
          height: 46,
          decoration: BoxDecoration(
            color: _hovering ? AppColors.navRailHover : Colors.transparent,
            borderRadius: BorderRadius.circular(18),
          ),
          child: widget.child,
        ),
      ),
    );
  }
}
