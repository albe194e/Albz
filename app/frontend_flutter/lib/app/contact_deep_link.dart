final class ContactDeepLink {
  const ContactDeepLink._();

  static String? parseContactCode(Uri uri) {
    if (!_matchesSupportedRoute(uri)) {
      return null;
    }

    final rawCode = _extractRawCode(uri);
    if (rawCode == null) {
      return null;
    }

    return normalizeContactCode(rawCode);
  }

  static String? normalizeContactCode(String value) {
    final trimmed = value.trim().toUpperCase();
    if (!RegExp(r'^HADDLE-[A-Z0-9]{4,}$').hasMatch(trimmed)) {
      return null;
    }

    return trimmed;
  }

  static bool _matchesSupportedRoute(Uri uri) {
    final scheme = uri.scheme.toLowerCase();
    if (scheme == 'haddle') {
      final host = uri.host.toLowerCase();
      if (host == 'add-contact') {
        return true;
      }

      final segments = uri.pathSegments
          .where((segment) => segment.isNotEmpty)
          .map((segment) => segment.toLowerCase())
          .toList(growable: false);
      return segments.isNotEmpty && segments.first == 'add-contact';
    }

    if (scheme == 'http' || scheme == 'https') {
      final segments = uri.pathSegments
          .where((segment) => segment.isNotEmpty)
          .map((segment) => segment.toLowerCase())
          .toList(growable: false);
      return segments.isNotEmpty && segments.first == 'add-contact';
    }

    return false;
  }

  static String? _extractRawCode(Uri uri) {
    final queryCode = uri.queryParameters['code'];
    if (queryCode != null && queryCode.trim().isNotEmpty) {
      return queryCode;
    }

    final segments = uri.pathSegments
        .where((segment) => segment.isNotEmpty)
        .toList(growable: false);
    if (segments.length >= 2) {
      return segments.last;
    }

    return null;
  }
}
