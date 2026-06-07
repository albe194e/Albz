import 'dart:io';

import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';

class ProfilePicture extends StatelessWidget {
  const ProfilePicture({super.key, required this.path, this.size = 42});

  final String path;
  final double size;

  @override
  Widget build(BuildContext context) {
    final file = path.isNotEmpty ? File(path) : null;

    return ClipOval(
      child: Container(
        width: size,
        height: size,
        color: AppColors.accentSoft,
        child: file != null && file.existsSync()
            ? Image.file(file, fit: BoxFit.cover)
            : Icon(
                Icons.person_rounded,
                size: size * 0.58,
                color: AppColors.textPrimary,
              ),
      ),
    );
  }
}
