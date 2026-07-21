import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

import '../../../app/app_page.dart';
import '../../../app/app_scope.dart';
import '../../theme/app_theme.dart';
import '../../theme/responsive.dart';
import '../base/profile_picture.dart';

class MobileAppDrawer extends StatelessWidget {
  const MobileAppDrawer({super.key});

  static const _chatIconAsset = 'assets/icons/chat.svg';
  static const _contactsIconAsset = 'assets/icons/contacts.svg';
  static const _logoAsset = 'assets/icons/icon.svg';

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final page = controller.page;
    final width = MediaQuery.sizeOf(context).width;
    final drawerWidth = AppBreakpoints.mobileDrawerWidth(width);

    return Stack(
      children: [
        Positioned.fill(
          child: GestureDetector(
            behavior: HitTestBehavior.opaque,
            onTap: controller.closeMobileNav,
            child: ColoredBox(color: Colors.black.withValues(alpha: 0.3)),
          ),
        ),
        Align(
          alignment: Alignment.centerLeft,
          child: Container(
            width: drawerWidth,
            height: double.infinity,
            color: AppColors.navRail,
            child: SafeArea(
              right: false,
              bottom: false,
              child: Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 18,
                  vertical: 20,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SizedBox(
                      width: 54,
                      height: 54,
                      child: Padding(
                        padding: const EdgeInsets.all(6),
                        child: SvgPicture.asset(
                          _logoAsset,
                          fit: BoxFit.contain,
                          errorBuilder: (context, error, stackTrace) {
                            return Center(
                              child: Text(
                                'AL',
                                style: Theme.of(context).textTheme.titleMedium
                                    ?.copyWith(
                                      fontSize: 18,
                                      fontWeight: FontWeight.w700,
                                      letterSpacing: 0.8,
                                    ),
                              ),
                            );
                          },
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                    _DrawerItem(
                      selected: page == AppPage.chat,
                      label: 'Chat',
                      icon: SvgPicture.asset(
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
                      onTap: () => controller.navigateTo(AppPage.chat),
                    ),
                    const SizedBox(height: 10),
                    _DrawerItem(
                      selected: page == AppPage.contacts,
                      label: 'Contacts',
                      icon: SvgPicture.asset(
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
                      onTap: () => controller.navigateTo(AppPage.contacts),
                    ),
                    const SizedBox(height: 16),
                    FractionallySizedBox(
                      widthFactor: 0.7,
                      child: Container(height: 1, color: AppColors.borderSoft),
                    ),
                    const Spacer(),
                    _DrawerItem(
                      selected: page == AppPage.profile,
                      label: 'Profile',
                      icon: ProfilePicture(
                        path: controller.currentUser?.profilePicturePath ?? '',
                        size: 28,
                      ),
                      onTap: () => controller.navigateTo(AppPage.profile),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _DrawerItem extends StatelessWidget {
  const _DrawerItem({
    required this.selected,
    required this.label,
    required this.icon,
    required this.onTap,
  });

  final bool selected;
  final String label;
  final Widget icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: selected ? AppColors.accentSoft : Colors.transparent,
      borderRadius: BorderRadius.circular(18),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
          child: Row(
            children: [
              SizedBox(width: 28, child: Center(child: icon)),
              const SizedBox(width: 14),
              Text(
                label,
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: selected
                      ? AppColors.textPrimary
                      : AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
