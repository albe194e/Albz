import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:path_provider/path_provider.dart';

import 'app/app_controller.dart';
import 'app/app_page.dart';
import 'app/app_scope.dart';
import 'platform/desktop_window.dart';
import 'ui/components/navigation/app_shell.dart';
import 'ui/components/navigation/desktop_title_bar.dart';
import 'ui/router/app_router.dart';
import 'ui/theme/app_theme.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await DesktopWindow.configure();

  const profileName = String.fromEnvironment('ALBZ_PROFILE');
  const configuredDataDir = String.fromEnvironment('ALBZ_DATA_DIR');
  const serverUrl = String.fromEnvironment(
    'ALBZ_SERVER_WS_URL',
    defaultValue: 'wss://oncological-paroxysmal-karolyn.ngrok-free.dev/ws',
  );
  final resolvedDataDir = configuredDataDir.isEmpty
      ? await _resolveDefaultDataDir()
      : configuredDataDir;

  if (kDebugMode) {
    debugPrint(
      '[albz] Flutter startup config: '
      'profile="${profileName.isEmpty ? '(default)' : profileName}", '
      'dataDir="${resolvedDataDir ?? '(core-go default)'}", '
      'serverUrl="$serverUrl"',
    );
  }

  runApp(
    AlbzApp(
      profileName: profileName,
      dataDir: resolvedDataDir,
      serverUrl: serverUrl,
    ),
  );
}

Future<String?> _resolveDefaultDataDir() async {
  if (!Platform.isAndroid) {
    return null;
  }

  final supportDir = await getApplicationSupportDirectory();
  return [
    supportDir.path,
    'dev-local-db',
    'local_storage',
  ].join(Platform.pathSeparator);
}

class AlbzApp extends StatefulWidget {
  const AlbzApp({
    super.key,
    this.profileName = '',
    this.dataDir,
    this.serverUrl = 'wss://oncological-paroxysmal-karolyn.ngrok-free.dev/ws',
  });

  final String profileName;
  final String? dataDir;
  final String serverUrl;

  @override
  State<AlbzApp> createState() => _AlbzAppState();
}

class _AlbzAppState extends State<AlbzApp> {
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
            ? 'Albz Flutter'
            : 'Albz Flutter (${widget.profileName})',
        theme: buildAppTheme(),
        home: AnimatedBuilder(
          animation: _controller,
          builder: (context, _) {
            return Scaffold(
              backgroundColor: AppColors.appBackground,
              body: Column(
                children: [
                  if (useDesktopWindowChrome) const DesktopTitleBar(),
                  Expanded(
                    child: SafeArea(
                      top: !useDesktopWindowChrome,
                      child: Column(
                        children: [
                          if (_controller.errorMessage.isNotEmpty)
                            _AppBanner(
                              message: _controller.errorMessage,
                              color: AppColors.error,
                            ),
                          if (_controller.infoMessage.isNotEmpty)
                            _AppBanner(
                              message: _controller.infoMessage,
                              color: AppColors.textMuted,
                            ),
                          Expanded(
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
                        ],
                      ),
                    ),
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
