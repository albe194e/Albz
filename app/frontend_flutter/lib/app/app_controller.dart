import 'dart:async';

import 'package:flutter/foundation.dart';

import '../core/core_api.dart';
import 'app_page.dart';

class AppController extends ChangeNotifier {
  AppController({
    this.profileName = '',
    this.dataDir,
    this.serverUrl = 'ws://localhost:8080/ws',
  }) {
    bootstrap();
  }

  final String profileName;
  final String? dataDir;
  final String serverUrl;

  CoreApi? _coreApi;
  CoreConfig? _config;
  CoreSnapshot? _snapshot;

  AppPage _page = AppPage.landing;
  String _activeConversationId = '';
  String _activeConversationName = 'Choose a conversation';
  bool _sidebarOpen = false;
  bool _mobileNavOpen = false;
  bool _bootstrapping = false;
  bool _authInFlight = false;
  bool _initialized = false;
  String _statusMessage = 'Initializing core-go...';
  String _errorMessage = '';
  String _infoMessage = '';
  Timer? _eventPollTimer;
  bool _pollInFlight = false;

  AppPage get page => _page;
  String get activeConversationId => _activeConversationId;
  String get activeConversationName => _activeConversationName;
  bool get sidebarOpen => _sidebarOpen;
  bool get mobileNavOpen => _mobileNavOpen;
  bool get isBootstrapping => _bootstrapping;
  bool get isAuthInFlight => _authInFlight;
  bool get isInitialized => _initialized;
  String get statusMessage => _statusMessage;
  String get errorMessage => _errorMessage;
  String get infoMessage => _infoMessage;
  CoreConfig? get config => _config;
  CoreSnapshot? get snapshot => _snapshot;

  CoreUser? get currentUser => _snapshot?.currentUser;
  List<CoreConversation> get conversations =>
      _snapshot?.conversations ?? const [];
  List<CoreMessage> get messages {
    final snapshot = _snapshot;
    if (snapshot == null) {
      return const [];
    }
    if (_activeConversationId.isEmpty) {
      return snapshot.messages;
    }
    return snapshot.messages
        .where((message) => message.conversationId == _activeConversationId)
        .toList(growable: false);
  }

  List<CoreContact> get contacts => _snapshot?.contacts ?? const [];
  List<CoreContactRequest> get contactRequests =>
      _snapshot?.contactRequests ?? const [];

  Future<void> bootstrap() async {
    if (_bootstrapping || _initialized) {
      return;
    }

    _bootstrapping = true;
    _errorMessage = '';
    _statusMessage = 'Initializing core-go...';
    _debugLog(
      'Bootstrapping core-go with '
      'profile="${profileName.isEmpty ? '(default)' : profileName}", '
      'dataDir="${dataDir ?? '(core-go default)'}", '
      'serverUrl="$serverUrl"',
    );
    notifyListeners();

    try {
      final coreApi = CoreApi.load();
      await coreApi.initialize(
        profileName: profileName,
        dataDir: dataDir,
        serverUrl: serverUrl,
      );
      _coreApi = coreApi;
      _config = coreApi.config();
      _debugLog(
        'core-go config loaded: '
        'dataDir="${_config?.dataDir ?? ''}", '
        'dbPath="${_config?.dbPath ?? ''}", '
        'serverUrl="${_config?.serverUrl ?? ''}"',
      );
      final loadedSession = coreApi.tryLoadSession();
      _applySnapshot(coreApi.snapshot());
      _initialized = true;
      _statusMessage = loadedSession
          ? 'Loaded existing local session'
          : 'core-go initialized';
      _page = loadedSession ? AppPage.chat : AppPage.landing;
      _sidebarOpen = loadedSession;
      _mobileNavOpen = false;
      _startEventPolling();
    } catch (error, stackTrace) {
      _setErrorMessage(
        error.toString(),
        context: 'bootstrap',
        stackTrace: stackTrace,
        logStackTrace: true,
      );
      _statusMessage = 'Failed to initialize core-go';
    } finally {
      _bootstrapping = false;
      notifyListeners();
    }
  }

  Future<void> refreshSnapshot() async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage(
        'core-go is not initialized',
        context: 'refreshSnapshot',
      );
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      _config = coreApi.config();
      _applySnapshot(coreApi.snapshot());
      _statusMessage = 'Snapshot refreshed';
    } catch (error, stackTrace) {
      _setErrorMessage(
        error.toString(),
        context: 'refreshSnapshot',
        stackTrace: stackTrace,
      );
    }

    notifyListeners();
  }

  Future<void> openConversation(String conversationId) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage(
        'core-go is not initialized',
        context: 'openConversation',
      );
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.openConversation(conversationId);
      _applySnapshot(coreApi.snapshot());
      _sidebarOpen = false;
      _mobileNavOpen = false;
    } catch (error, stackTrace) {
      _setErrorMessage(
        error.toString(),
        context: 'openConversation',
        stackTrace: stackTrace,
      );
    }

    notifyListeners();
  }

  Future<void> login({
    required String username,
    required String password,
  }) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage('core-go is not initialized', context: 'login');
      notifyListeners();
      return;
    }

    _authInFlight = true;
    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.login(username: username, password: password);
      _config = coreApi.config();
      _applySnapshot(coreApi.snapshot());
      _statusMessage = 'Signed in';
      _page = AppPage.chat;
      _sidebarOpen = true;
      _mobileNavOpen = false;
    } catch (error) {
      _setErrorMessage(error.toString(), context: 'login');
    } finally {
      _authInFlight = false;
      notifyListeners();
    }
  }

  Future<void> register({
    required String name,
    required String username,
    required String password,
    String profilePicturePath = '',
  }) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage('core-go is not initialized', context: 'register');
      notifyListeners();
      return;
    }

    _authInFlight = true;
    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.register(
        name: name,
        username: username,
        password: password,
        profilePicturePath: profilePicturePath,
      );
      _config = coreApi.config();
      _applySnapshot(coreApi.snapshot());
      _statusMessage = 'Account created';
      _page = AppPage.chat;
      _sidebarOpen = true;
      _mobileNavOpen = false;
    } catch (error) {
      _setErrorMessage(error.toString(), context: 'register');
    } finally {
      _authInFlight = false;
      notifyListeners();
    }
  }

  Future<void> logout() async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage('core-go is not initialized', context: 'logout');
      notifyListeners();
      return;
    }

    _authInFlight = true;
    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.logout();
      _applySnapshot(coreApi.snapshot());
      _activeConversationId = '';
      _activeConversationName = 'Choose a conversation';
      _sidebarOpen = false;
      _mobileNavOpen = false;
      _statusMessage = 'Signed out';
      _page = AppPage.landing;
    } catch (error, stackTrace) {
      _setErrorMessage(
        error.toString(),
        context: 'logout',
        stackTrace: stackTrace,
      );
    } finally {
      _authInFlight = false;
      notifyListeners();
    }
  }

  Future<bool> sendMessage(String body) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage('core-go is not initialized', context: 'sendMessage');
      notifyListeners();
      return false;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.sendMessage(
        conversationId: _activeConversationId,
        body: body,
      );
      _applySnapshot(coreApi.snapshot());
      return true;
    } catch (error) {
      _setErrorMessage(error.toString(), context: 'sendMessage');
      notifyListeners();
      return false;
    } finally {
      notifyListeners();
    }
  }

  Future<void> startDirectConversation(String contactUserId) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage(
        'core-go is not initialized',
        context: 'startDirectConversation',
      );
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      final conversationId = await coreApi.startDirectConversation(
        contactUserId,
      );
      _applySnapshot(coreApi.snapshot());
      _page = AppPage.chat;
      await openConversation(conversationId);
      return;
    } catch (error) {
      _setErrorMessage(error.toString(), context: 'startDirectConversation');
    }

    notifyListeners();
  }

  Future<void> createConversation({
    required String name,
    required List<String> participantUserIds,
  }) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage(
        'core-go is not initialized',
        context: 'createConversation',
      );
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      final conversationId = await coreApi.createConversation(
        name: name,
        participantUserIds: participantUserIds,
      );
      _applySnapshot(coreApi.snapshot());
      _page = AppPage.chat;
      await openConversation(conversationId);
      _infoMessage = 'Conversation created';
    } catch (error) {
      _setErrorMessage(error.toString(), context: 'createConversation');
    }

    notifyListeners();
  }

  Future<void> sendContactRequest(String contactCode) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage(
        'core-go is not initialized',
        context: 'sendContactRequest',
      );
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.sendContactRequest(contactCode);
      _infoMessage = 'Contact request sent';
    } catch (error) {
      _setErrorMessage(error.toString(), context: 'sendContactRequest');
    }

    notifyListeners();
  }

  Future<void> acceptContactRequest(String fromUserId) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage(
        'core-go is not initialized',
        context: 'acceptContactRequest',
      );
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.acceptContactRequest(fromUserId);
      _infoMessage = 'Contact request accepted';
    } catch (error) {
      _setErrorMessage(error.toString(), context: 'acceptContactRequest');
    }

    notifyListeners();
  }

  Future<void> rejectContactRequest(String fromUserId) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _setErrorMessage(
        'core-go is not initialized',
        context: 'rejectContactRequest',
      );
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.rejectContactRequest(fromUserId);
      _infoMessage = 'Contact request rejected';
    } catch (error) {
      _setErrorMessage(error.toString(), context: 'rejectContactRequest');
    }

    notifyListeners();
  }

  Future<String> saveProfileImage({
    required Uint8List data,
    required String filename,
  }) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      throw CoreApiError('core-go is not initialized');
    }

    final savedPath = await coreApi.saveProfileImage(
      data: data,
      filename: filename,
    );
    _infoMessage = 'Profile picture selected';
    _errorMessage = '';
    notifyListeners();
    return savedPath;
  }

  void navigateTo(AppPage page) {
    _page = page;
    _mobileNavOpen = false;
    _sidebarOpen = page == AppPage.chat;
    _infoMessage = '';
    notifyListeners();
  }

  void selectConversation(String conversationId, String conversationName) {
    _activeConversationId = conversationId;
    _activeConversationName = conversationName;
    _sidebarOpen = false;
    _mobileNavOpen = false;
    notifyListeners();
  }

  void toggleSidebar() {
    _sidebarOpen = !_sidebarOpen;
    notifyListeners();
  }

  void showConversationList() {
    _page = AppPage.chat;
    _sidebarOpen = true;
    _mobileNavOpen = false;
    notifyListeners();
  }

  void openMobileNav() {
    if (_mobileNavOpen) {
      return;
    }
    _mobileNavOpen = true;
    notifyListeners();
  }

  void toggleMobileNav() {
    _mobileNavOpen = !_mobileNavOpen;
    notifyListeners();
  }

  void closeMobileNav() {
    if (!_mobileNavOpen) {
      return;
    }
    _mobileNavOpen = false;
    notifyListeners();
  }

  void closeSidebar() {
    if (!_sidebarOpen) {
      return;
    }
    _sidebarOpen = false;
    notifyListeners();
  }

  void showInfo(String message) {
    _infoMessage = message;
    _errorMessage = '';
    notifyListeners();
  }

  void showUnsupportedAction(String action) {
    showInfo('$action is the next wiring step for the Flutter migration.');
  }

  void _applySnapshot(CoreSnapshot snapshot) {
    final previousNetworkError = _snapshot?.lastNetworkError ?? '';
    _snapshot = snapshot;

    if (snapshot.lastNetworkError.isNotEmpty &&
        snapshot.lastNetworkError != previousNetworkError) {
      _debugLog('Network error from core-go: ${snapshot.lastNetworkError}');
    }

    if (_activeConversationId.isNotEmpty &&
        !snapshot.conversations.any(
          (conversation) => conversation.id == _activeConversationId,
        )) {
      _activeConversationId = '';
      _activeConversationName = 'Choose a conversation';
    }

    if (snapshot.loadedConversationId.isNotEmpty) {
      final loadedConversation = snapshot.conversations
          .where(
            (conversation) => conversation.id == snapshot.loadedConversationId,
          )
          .firstOrNull;
      if (loadedConversation != null) {
        _activeConversationId = loadedConversation.id;
        _activeConversationName = loadedConversation.name;
        return;
      }
    }

    if (_activeConversationId.isEmpty && snapshot.conversations.isNotEmpty) {
      final firstConversation = snapshot.conversations.first;
      _activeConversationId = firstConversation.id;
      _activeConversationName = firstConversation.name;
    }
  }

  void _startEventPolling() {
    _eventPollTimer?.cancel();
    _eventPollTimer = Timer.periodic(const Duration(milliseconds: 350), (_) {
      unawaited(_pollEvents());
    });
  }

  Future<void> _pollEvents() async {
    final coreApi = _coreApi;
    if (_pollInFlight || coreApi == null || !_initialized) {
      return;
    }

    _pollInFlight = true;
    var changed = false;
    try {
      while (true) {
        final event = coreApi.pollEvent();
        if (event == null) {
          break;
        }
        if (event.snapshot != null) {
          _applySnapshot(event.snapshot!);
          changed = true;
        }
      }
    } catch (error, stackTrace) {
      _setErrorMessage(
        error.toString(),
        context: 'pollEvents',
        stackTrace: stackTrace,
        logStackTrace: true,
      );
      changed = true;
    } finally {
      _pollInFlight = false;
    }

    if (changed) {
      notifyListeners();
    }
  }

  @override
  void dispose() {
    _eventPollTimer?.cancel();
    _coreApi?.dispose();
    super.dispose();
  }

  void _setErrorMessage(
    String message, {
    required String context,
    StackTrace? stackTrace,
    bool logStackTrace = false,
  }) {
    _errorMessage = message;
    if (message.isEmpty) {
      return;
    }
    _debugLog('[$context] ERROR: $message');
    if (logStackTrace && stackTrace != null && kDebugMode) {
      debugPrintStack(stackTrace: stackTrace, label: '[albz][$context]');
    }
  }

  void _debugLog(String message) {
    if (!kDebugMode) {
      return;
    }
    debugPrint('[albz] $message');
  }
}
