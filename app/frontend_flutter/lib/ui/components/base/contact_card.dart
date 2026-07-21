import 'dart:io';

import 'package:flutter/material.dart';
import 'package:frontend_flutter/core/core_api.dart';

import '../../theme/app_theme.dart';

class ContactCard extends StatelessWidget {
  const ContactCard({super.key, this.onTap, required this.contact});

  final VoidCallback? onTap;
  final CoreContact contact;

  @override
  Widget build(BuildContext context) {
    final path = contact.profilePicturePath;
    final file = path.isNotEmpty ? File(path) : null;
    final double size = 42;
    return Material(
      color: AppColors.cardMuted,
      borderRadius: BorderRadius.circular(24),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(24),
        hoverColor: AppColors.navRailHover.withValues(alpha: 0.45),
        child: AspectRatio(
          aspectRatio: 1,
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                file != null && file.existsSync()
                    ? Image.file(file, fit: BoxFit.cover)
                    : Icon(
                        Icons.person_rounded,
                        size: size,
                        color: AppColors.textPrimary,
                      ),
                const SizedBox(height: 18),
                Text(
                  contact.displayName,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
