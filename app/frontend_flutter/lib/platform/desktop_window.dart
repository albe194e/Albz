import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

class DesktopWindow {
  DesktopWindow._();

  static bool get isSupported =>
      Platform.isWindows || Platform.isMacOS || Platform.isLinux;

  static Future<void> configure() async {
    if (!isSupported) {
      return;
    }

    await windowManager.ensureInitialized();
    const options = WindowOptions(
      size: Size(1280, 720),
      center: true,
      backgroundColor: Colors.transparent,
      skipTaskbar: false,
      titleBarStyle: TitleBarStyle.hidden,
    );

    windowManager.waitUntilReadyToShow(options, () async {
      await windowManager.show();
      await windowManager.focus();
    });
  }

  static Future<void> startDrag() async {
    if (!isSupported) {
      return;
    }
    await windowManager.startDragging();
  }

  static Future<void> minimize() async {
    if (!isSupported) {
      return;
    }
    await windowManager.minimize();
  }

  static Future<void> toggleMaximize() async {
    if (!isSupported) {
      return;
    }
    if (await windowManager.isMaximized()) {
      await windowManager.unmaximize();
      return;
    }
    await windowManager.maximize();
  }

  static Future<bool> isMaximized() async {
    if (!isSupported) {
      return false;
    }
    return windowManager.isMaximized();
  }

  static Future<void> close() async {
    if (!isSupported) {
      return;
    }
    await windowManager.close();
  }
}
