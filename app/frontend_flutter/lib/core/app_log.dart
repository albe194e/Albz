import 'dart:io';

import 'package:flutter/foundation.dart';

class AppLog {
  AppLog._();

  static final bool devMode = _resolveDevMode();

  static void debug(String message) {
    if (!devMode) {
      return;
    }
    debugPrint('[haddle][debug] $message');
  }

  static void info(String message) {
    if (!devMode) {
      return;
    }
    debugPrint('[haddle][info] $message');
  }

  static void error({
    required String context,
    required String message,
    StackTrace? stackTrace,
    bool logStackTrace = false,
    bool logInNonDev = false,
  }) {
    if (devMode) {
      debugPrint('[haddle][error][$context] $message');
      if (logStackTrace && stackTrace != null) {
        debugPrintStack(stackTrace: stackTrace, label: '[haddle][$context]');
      }
      return;
    }

    if (!logInNonDev) {
      return;
    }

    debugPrint('[haddle][error][$context] operation failed');
  }

  static bool _parseDevMode(String raw) {
    switch (raw.trim().toLowerCase()) {
      case '':
      case '1':
      case 'true':
      case 'yes':
      case 'on':
        return true;
      case '0':
      case 'false':
      case 'no':
      case 'off':
        return false;
      default:
        return true;
    }
  }

  static bool _resolveDevMode() {
    const compileTimeValue = String.fromEnvironment('HADDLE_DEV_MODE');
    if (compileTimeValue.isNotEmpty) {
      return _parseDevMode(compileTimeValue);
    }

    if (_supportsRuntimeEnvironment()) {
      final runtimeValue = Platform.environment['HADDLE_DEV_MODE'];
      if (runtimeValue != null && runtimeValue.isNotEmpty) {
        return _parseDevMode(runtimeValue);
      }
    }

    return true;
  }

  static bool _supportsRuntimeEnvironment() {
    return Platform.isWindows || Platform.isLinux || Platform.isMacOS;
  }
}
