import 'dart:io';

import 'package:flutter/material.dart';
import 'package:path_provider/path_provider.dart';

import 'app/app_controller.dart';
import 'app/app_page.dart';
import 'app/app_scope.dart';
import 'core/app_log.dart';
import 'platform/desktop_window.dart';
import 'ui/components/navigation/app_shell.dart';
import 'ui/components/navigation/desktop_title_bar.dart';
import 'ui/router/app_router.dart';
import 'ui/theme/app_theme.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await DesktopWindow.configure();

  const compileTimeProfileName = String.fromEnvironment('HADDLE_PROFILE');
  const compileTimeDataDir = String.fromEnvironment('HADDLE_DATA_DIR');
  const compileTimeServerUrl = String.fromEnvironment(
    'HADDLE_SERVER_WS_URL',
    defaultValue: 'wss://oncological-paroxysmal-karolyn.ngrok-free.dev/ws',
  );
  final profileName = _resolveDesktopConfigValue(
    'HADDLE_PROFILE',
    compileTimeProfileName,
  );
  final configuredDataDir = _resolveDesktopConfigValue(
    'HADDLE_DATA_DIR',
    compileTimeDataDir,
  );
  final serverUrl = _resolveDesktopConfigValue(
    'HADDLE_SERVER_WS_URL',
    compileTimeServerUrl,
  );
  final resolvedDataDir = configuredDataDir.isEmpty
      ? await _resolveDefaultDataDir()
      : configuredDataDir;

  AppLog.info(
    'Flutter startup config: '
    'profile="${profileName.isEmpty ? '(default)' : profileName}", '
    'dataDir="${resolvedDataDir ?? '(core-go default)'}", '
    'serverUrl="$serverUrl"',
  );

  runApp(
    HaddleApp(
      profileName: profileName,
      dataDir: resolvedDataDir,
      serverUrl: serverUrl,
    ),
  );
}

String _resolveDesktopConfigValue(String key, String compileTimeValue) {
  if (compileTimeValue.isNotEmpty) {
    return compileTimeValue;
  }
  if (Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
    return Platform.environment[key] ?? '';
  }
  return '';
}

Future<String?> _resolveDefaultDataDir() async {
  if (!Platform.isAndroid && !Platform.isIOS) {
    return null;
  }

  final supportDir = await getApplicationSupportDirectory();
  return [
    supportDir.path,
    'dev-local-db',
    'local_storage',
  ].join(Platform.pathSeparator);
}

class HaddleApp extends StatefulWidget {
  const HaddleApp({
    super.key,
    this.profileName = '',
    this.dataDir,
    this.serverUrl = 'wss://oncological-paroxysmal-karolyn.ngrok-free.dev/ws',
  });

  final String profileName;
  final String? dataDir;
  final String serverUrl;

  @override
  State<HaddleApp> createState() => _HaddleAppState();
}

class _HaddleAppState extends State<HaddleApp> {
  late final AppController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AppController(
      profileName: widget.profileName,
      dataDir: widget.dataDir,
      serverUrl: widget.serverUrl,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final useDesktopWindowChrome = DesktopWindow.isSupported;

    return AppScope(
      controller: _controller,
      child: MaterialApp(
        title: widget.profileName.isEmpty
            ? 'Haddle'
            : 'Haddle (${widget.profileName})',
        theme: buildAppTheme(),
        home: AnimatedBuilder(
          animation: _controller,
          builder: (context, _) {
            return Scaffold(
              backgroundColor: AppColors.appBackground,
              body: Stack(
                children: [
                  Column(
                    children: [
                      if (useDesktopWindowChrome) const DesktopTitleBar(),
                      Expanded(
                        child: SafeArea(
                          top: !useDesktopWindowChrome,
                          child: switch (_controller.page) {
                            AppPage.chat ||
                            AppPage.profile ||
                            AppPage.contacts => AppShell(
                              navRailOverlapTop: useDesktopWindowChrome
                                  ? 42
                                  : 0,
                              child: AppRouter(controller: _controller),
                            ),
                            AppPage.landing ||
                            AppPage.login ||
                            AppPage.register => AppRouter(
                              controller: _controller,
                            ),
                          },
                        ),
                      ),
                    ],
                  ),
                  _AppBannerOverlay(
                    desktopOffset: useDesktopWindowChrome ? 42 : 0,
                    errorMessage: _controller.errorMessage,
                    infoMessage: _controller.infoMessage,
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
  }
}

class _AppBanner extends StatelessWidget {
  const _AppBanner({required this.message, required this.color});

  final String message;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: AppColors.card,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Text(message, style: TextStyle(color: color)),
      ),
    );
  }
}

class _AppBannerOverlay extends StatelessWidget {
  const _AppBannerOverlay({
    required this.desktopOffset,
    required this.errorMessage,
    required this.infoMessage,
  });

  final double desktopOffset;
  final String errorMessage;
  final String infoMessage;

  @override
  Widget build(BuildContext context) {
    final hasBanner = errorMessage.isNotEmpty || infoMessage.isNotEmpty;

    return IgnorePointer(
      child: SafeArea(
        bottom: false,
        child: Padding(
          padding: EdgeInsets.fromLTRB(16, desktopOffset + 16, 16, 0),
          child: Align(
            alignment: Alignment.topCenter,
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 520),
              child: AnimatedSwitcher(
                duration: const Duration(milliseconds: 180),
                child: !hasBanner
                    ? const SizedBox.shrink()
                    : Column(
                        key: ValueKey('$errorMessage|$infoMessage'),
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          if (errorMessage.isNotEmpty)
                            _AppBanner(
                              message: errorMessage,
                              color: AppColors.error,
                            ),
                          if (infoMessage.isNotEmpty)
                            _AppBanner(
                              message: infoMessage,
                              color: AppColors.textMuted,
                            ),
                        ],
                      ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
