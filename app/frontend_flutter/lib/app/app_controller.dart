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

  List<CoreFriend> get friends => _snapshot?.friends ?? const [];
  List<CoreFriendRequest> get friendRequests =>
      _snapshot?.friendRequests ?? const [];

  Future<void> bootstrap() async {
    if (_bootstrapping || _initialized) {
      return;
    }

    _bootstrapping = true;
    _errorMessage = '';
    _statusMessage = 'Initializing core-go...';
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
      final loadedSession = coreApi.tryLoadSession();
      _applySnapshot(coreApi.snapshot());
      _initialized = true;
      _statusMessage = loadedSession
          ? 'Loaded existing local session'
          : 'core-go initialized';
      _page = loadedSession ? AppPage.chat : AppPage.landing;
      _startEventPolling();
    } catch (error) {
      _errorMessage = error.toString();
      _statusMessage = 'Failed to initialize core-go';
    } finally {
      _bootstrapping = false;
      notifyListeners();
    }
  }

  Future<void> refreshSnapshot() async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
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
    } catch (error) {
      _errorMessage = error.toString();
    }

    notifyListeners();
  }

  Future<void> openConversation(String conversationId) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
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
    } catch (error) {
      _errorMessage = error.toString();
    }

    notifyListeners();
  }

  Future<void> login({
    required String username,
    required String password,
  }) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
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
    } catch (error) {
      _errorMessage = error.toString();
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
      _errorMessage = 'core-go is not initialized';
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
    } catch (error) {
      _errorMessage = error.toString();
    } finally {
      _authInFlight = false;
      notifyListeners();
    }
  }

  Future<void> logout() async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
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
      _statusMessage = 'Signed out';
      _page = AppPage.landing;
    } catch (error) {
      _errorMessage = error.toString();
    } finally {
      _authInFlight = false;
      notifyListeners();
    }
  }

  Future<bool> sendMessage(String body) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
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
      _errorMessage = error.toString();
      notifyListeners();
      return false;
    } finally {
      notifyListeners();
    }
  }

  Future<void> startDirectConversation(String friendUserId) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      final conversationId = await coreApi.startDirectConversation(
        friendUserId,
      );
      _applySnapshot(coreApi.snapshot());
      _page = AppPage.chat;
      await openConversation(conversationId);
      return;
    } catch (error) {
      _errorMessage = error.toString();
    }

    notifyListeners();
  }

  Future<void> createConversation({
    required String name,
    required List<String> participantUserIds,
  }) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
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
      _errorMessage = error.toString();
    }

    notifyListeners();
  }

  Future<void> sendFriendRequest(String friendCode) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.sendFriendRequest(friendCode);
      _infoMessage = 'Friend request sent';
    } catch (error) {
      _errorMessage = error.toString();
    }

    notifyListeners();
  }

  Future<void> acceptFriendRequest(String fromUserId) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.acceptFriendRequest(fromUserId);
      _infoMessage = 'Friend request accepted';
    } catch (error) {
      _errorMessage = error.toString();
    }

    notifyListeners();
  }

  Future<void> rejectFriendRequest(String fromUserId) async {
    final coreApi = _coreApi;
    if (coreApi == null || !_initialized) {
      _errorMessage = 'core-go is not initialized';
      notifyListeners();
      return;
    }

    _errorMessage = '';
    _infoMessage = '';
    notifyListeners();

    try {
      await coreApi.rejectFriendRequest(fromUserId);
      _infoMessage = 'Friend request rejected';
    } catch (error) {
      _errorMessage = error.toString();
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
    _infoMessage = '';
    notifyListeners();
  }

  void selectConversation(String conversationId, String conversationName) {
    _activeConversationId = conversationId;
    _activeConversationName = conversationName;
    _sidebarOpen = false;
    notifyListeners();
  }

  void toggleSidebar() {
    _sidebarOpen = !_sidebarOpen;
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
    _snapshot = snapshot;

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
    } catch (error) {
      _errorMessage = error.toString();
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
}
