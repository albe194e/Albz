import 'package:flutter/material.dart';

import 'app/app_controller.dart';
import 'app/app_scope.dart';
import 'ui/components/navigation/app_shell.dart';
import 'ui/router/app_router.dart';
import 'ui/theme/app_theme.dart';

void main() {
  const profileName = String.fromEnvironment('ALBZ_PROFILE');
  const configuredDataDir = String.fromEnvironment('ALBZ_DATA_DIR');
  const serverUrl = String.fromEnvironment(
    'ALBZ_SERVER_WS_URL',
    defaultValue: 'ws://localhost:8080/ws',
  );

  runApp(
    AlbzApp(
      profileName: profileName,
      dataDir: configuredDataDir.isEmpty ? null : configuredDataDir,
      serverUrl: serverUrl,
    ),
  );
}

class AlbzApp extends StatefulWidget {
  const AlbzApp({
    super.key,
    this.profileName = '',
    this.dataDir,
    this.serverUrl = 'ws://localhost:8080/ws',
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
              body: SafeArea(
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
                      child: AppShell(
                        child: AppRouter(controller: _controller),
                      ),
                    ),
                  ],
                ),
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
