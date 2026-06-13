import 'dart:async';
import 'dart:io';

import 'package:flutter/services.dart';

final class DeepLinkBridge {
  const DeepLinkBridge._();

  static const MethodChannel _methodChannel = MethodChannel(
    'haddle/deep_links/methods',
  );
  static const EventChannel _eventChannel = EventChannel(
    'haddle/deep_links/events',
  );

  static Future<Uri?> getInitialUri() async {
    if (!_supportsPlatform) {
      return null;
    }

    final value = await _methodChannel.invokeMethod<String>('getInitialLink');
    return _parseValue(value);
  }

  static Stream<Uri> get uriStream {
    if (!_supportsPlatform) {
      return const Stream<Uri>.empty();
    }

    return _eventChannel
        .receiveBroadcastStream()
        .map(_parseValue)
        .where((uri) => uri != null)
        .cast<Uri>();
  }

  static bool get _supportsPlatform => Platform.isAndroid || Platform.isIOS;

  static Uri? _parseValue(dynamic value) {
    if (value is! String) {
      return null;
    }

    final trimmed = value.trim();
    if (trimmed.isEmpty) {
      return null;
    }

    return Uri.tryParse(trimmed);
  }
}
