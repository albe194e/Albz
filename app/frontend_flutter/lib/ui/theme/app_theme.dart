import 'package:flutter/material.dart';

class AppColors {
  static const appBackground = Color(0xFF24282D);
  static const appShell = Color(0xE614171B);
  static const panel = Color(0xCC1D2127);
  static const panelMuted = Color(0xB3171A1F);
  static const card = Color(0xFF2A2F36);
  static const cardMuted = Color(0xFF20242A);
  static const accent = Color(0xFF6080F3);
  static const accentSoft = Color(0xFF3C4460);
  static const borderSoft = Color(0xFF363B43);
  static const textPrimary = Color(0xFFF2F4F8);
  static const textSecondary = Color(0xFFB5BDC9);
  static const textMuted = Color(0xFF8992A0);
  static const error = Color(0xFFE78D8D);
}

ThemeData buildAppTheme() {
  const colorScheme = ColorScheme.dark(
    primary: AppColors.accent,
    secondary: AppColors.accent,
    surface: AppColors.card,
    error: AppColors.error,
  );

  return ThemeData(
    brightness: Brightness.dark,
    colorScheme: colorScheme,
    scaffoldBackgroundColor: AppColors.appBackground,
    cardColor: AppColors.card,
    useMaterial3: true,
    textTheme: const TextTheme(
      headlineLarge: TextStyle(
        color: AppColors.textPrimary,
        fontSize: 36,
        fontWeight: FontWeight.w700,
      ),
      headlineMedium: TextStyle(
        color: AppColors.textPrimary,
        fontSize: 28,
        fontWeight: FontWeight.w700,
      ),
      titleLarge: TextStyle(
        color: AppColors.textPrimary,
        fontSize: 20,
        fontWeight: FontWeight.w600,
      ),
      titleMedium: TextStyle(
        color: AppColors.textPrimary,
        fontSize: 16,
        fontWeight: FontWeight.w600,
      ),
      bodyLarge: TextStyle(color: AppColors.textPrimary, fontSize: 16),
      bodyMedium: TextStyle(color: AppColors.textSecondary, fontSize: 14),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: AppColors.cardMuted,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(18),
        borderSide: const BorderSide(color: AppColors.borderSoft),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(18),
        borderSide: const BorderSide(color: AppColors.borderSoft),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(18),
        borderSide: const BorderSide(color: AppColors.accent),
      ),
      hintStyle: const TextStyle(color: AppColors.textMuted),
      labelStyle: const TextStyle(color: AppColors.textSecondary),
    ),
  );
}
