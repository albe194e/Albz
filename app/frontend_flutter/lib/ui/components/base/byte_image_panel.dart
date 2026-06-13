import 'dart:typed_data';

import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';

class ByteImagePanel extends StatelessWidget {
  const ByteImagePanel({
    super.key,
    required this.bytes,
    required this.loading,
    this.emptyLabel = 'Image unavailable',
    this.size = 180,
  });

  final Uint8List? bytes;
  final bool loading;
  final String emptyLabel;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: AppColors.cardMuted,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.borderSoft),
      ),
      padding: const EdgeInsets.all(18),
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(18),
        ),
        child: Center(
          child: loading
              ? const SizedBox(
                  width: 28,
                  height: 28,
                  child: CircularProgressIndicator(strokeWidth: 2.4),
                )
              : bytes != null
              ? ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: Image.memory(
                    bytes!,
                    fit: BoxFit.contain,
                    filterQuality: FilterQuality.none,
                    gaplessPlayback: true,
                  ),
                )
              : Padding(
                  padding: const EdgeInsets.all(12),
                  child: Text(
                    emptyLabel,
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: AppColors.textSecondary,
                    ),
                  ),
                ),
        ),
      ),
    );
  }
}
