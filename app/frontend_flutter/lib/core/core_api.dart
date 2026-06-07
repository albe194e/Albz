import 'dart:convert';
import 'dart:ffi';
import 'dart:io';
import 'dart:typed_data';

import 'package:ffi/ffi.dart';

typedef _CoreCreateNative =
    Uint64 Function(
      Pointer<Utf8> profileName,
      Pointer<Utf8> dataDir,
      Pointer<Utf8> serverUrl,
    );
typedef _CoreCreateDart =
    int Function(
      Pointer<Utf8> profileName,
      Pointer<Utf8> dataDir,
      Pointer<Utf8> serverUrl,
    );

typedef _CoreCloseNative = Void Function(Uint64 handle);
typedef _CoreCloseDart = void Function(int handle);

typedef _CoreJSONNative = Pointer<Utf8> Function(Uint64 handle);
typedef _CoreJSONDart = Pointer<Utf8> Function(int handle);

typedef _CoreTryLoadSessionNative = Int32 Function(Uint64 handle);
typedef _CoreTryLoadSessionDart = int Function(int handle);

typedef _CoreLoginNative =
    Int32 Function(
      Uint64 handle,
      Pointer<Utf8> username,
      Pointer<Utf8> password,
    );
typedef _CoreLoginDart =
    int Function(int handle, Pointer<Utf8> username, Pointer<Utf8> password);

typedef _CoreRegisterNative =
    Int32 Function(
      Uint64 handle,
      Pointer<Utf8> name,
      Pointer<Utf8> username,
      Pointer<Utf8> password,
      Pointer<Utf8> profilePicturePath,
    );
typedef _CoreRegisterDart =
    int Function(
      int handle,
      Pointer<Utf8> name,
      Pointer<Utf8> username,
      Pointer<Utf8> password,
      Pointer<Utf8> profilePicturePath,
    );
typedef _CoreLogoutNative = Int32 Function(Uint64 handle);
typedef _CoreLogoutDart = int Function(int handle);

typedef _CoreOpenConversationNative =
    Int32 Function(Uint64 handle, Pointer<Utf8> conversationId);
typedef _CoreOpenConversationDart =
    int Function(int handle, Pointer<Utf8> conversationId);

typedef _CoreSendMessageNative =
    Int32 Function(
      Uint64 handle,
      Pointer<Utf8> conversationId,
      Pointer<Utf8> body,
    );
typedef _CoreSendMessageDart =
    int Function(int handle, Pointer<Utf8> conversationId, Pointer<Utf8> body);

typedef _CoreStartDirectConversationNative =
    Pointer<Utf8> Function(Uint64 handle, Pointer<Utf8> friendUserId);
typedef _CoreStartDirectConversationDart =
    Pointer<Utf8> Function(int handle, Pointer<Utf8> friendUserId);

typedef _CoreCreateConversationNative =
    Pointer<Utf8> Function(
      Uint64 handle,
      Pointer<Utf8> name,
      Pointer<Utf8> participantUserIdsJson,
    );
typedef _CoreCreateConversationDart =
    Pointer<Utf8> Function(
      int handle,
      Pointer<Utf8> name,
      Pointer<Utf8> participantUserIdsJson,
    );

typedef _CoreFriendActionNative =
    Int32 Function(Uint64 handle, Pointer<Utf8> value);
typedef _CoreFriendActionDart = int Function(int handle, Pointer<Utf8> value);

typedef _CoreSaveProfileImageNative =
    Pointer<Utf8> Function(
      Uint64 handle,
      Pointer<Uint8> data,
      Int32 length,
      Pointer<Utf8> filename,
    );
typedef _CoreSaveProfileImageDart =
    Pointer<Utf8> Function(
      int handle,
      Pointer<Uint8> data,
      int length,
      Pointer<Utf8> filename,
    );

typedef _CorePollEventJSONNative = Pointer<Utf8> Function(Uint64 handle);
typedef _CorePollEventJSONDart = Pointer<Utf8> Function(int handle);

typedef _CoreTakeLastErrorNative = Pointer<Utf8> Function();
typedef _CoreTakeLastErrorDart = Pointer<Utf8> Function();

typedef _CoreStringFreeNative = Void Function(Pointer<Utf8> value);
typedef _CoreStringFreeDart = void Function(Pointer<Utf8> value);

class CoreApiError implements Exception {
  CoreApiError(this.message);

  final String message;

  @override
  String toString() => 'CoreApiError: $message';
}

class CoreConfig {
  const CoreConfig({
    required this.profileName,
    required this.dataDir,
    required this.dbPath,
    required this.serverUrl,
  });

  final String profileName;
  final String dataDir;
  final String dbPath;
  final String serverUrl;

  factory CoreConfig.fromJson(Map<String, dynamic> json) {
    return CoreConfig(
      profileName: json['profile_name'] as String? ?? '',
      dataDir: json['data_dir'] as String? ?? '',
      dbPath: json['db_path'] as String? ?? '',
      serverUrl: json['server_url'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'profile_name': profileName,
      'data_dir': dataDir,
      'db_path': dbPath,
      'server_url': serverUrl,
    };
  }
}

class CoreUser {
  const CoreUser({
    required this.id,
    required this.name,
    required this.username,
    required this.profilePictureUrl,
    required this.friendCode,
  });

  final String id;
  final String name;
  final String username;
  final String profilePictureUrl;
  final String friendCode;

  factory CoreUser.fromJson(Map<String, dynamic> json) {
    return CoreUser(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      username: json['username'] as String? ?? '',
      profilePictureUrl: json['profile_picture_url'] as String? ?? '',
      friendCode: json['friend_code'] as String? ?? '',
    );
  }
}

class CoreConversation {
  const CoreConversation({required this.id, required this.name});

  final String id;
  final String name;

  factory CoreConversation.fromJson(Map<String, dynamic> json) {
    return CoreConversation(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
    );
  }
}

class CoreMessage {
  const CoreMessage({
    required this.id,
    required this.conversationId,
    required this.senderId,
    required this.clientMessageId,
    required this.body,
    required this.createdAt,
    required this.deliveryState,
  });

  final int id;
  final String conversationId;
  final String senderId;
  final String clientMessageId;
  final String body;
  final int createdAt;
  final String deliveryState;

  factory CoreMessage.fromJson(Map<String, dynamic> json) {
    return CoreMessage(
      id: json['id'] as int? ?? 0,
      conversationId: json['conversation_id'] as String? ?? '',
      senderId: json['sender_id'] as String? ?? '',
      clientMessageId: json['client_message_id'] as String? ?? '',
      body: json['body'] as String? ?? '',
      createdAt: json['created_at'] as int? ?? 0,
      deliveryState: json['delivery_state'] as String? ?? '',
    );
  }
}

class CoreFriend {
  const CoreFriend({
    required this.id,
    required this.userId,
    required this.name,
    required this.username,
    required this.profilePictureUrl,
    required this.friendCode,
    required this.createdAt,
  });

  final int id;
  final String userId;
  final String name;
  final String username;
  final String profilePictureUrl;
  final String friendCode;
  final int createdAt;

  factory CoreFriend.fromJson(Map<String, dynamic> json) {
    return CoreFriend(
      id: json['id'] as int? ?? 0,
      userId: json['user_id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      username: json['username'] as String? ?? '',
      profilePictureUrl: json['profile_picture_url'] as String? ?? '',
      friendCode: json['friend_code'] as String? ?? '',
      createdAt: json['created_at'] as int? ?? 0,
    );
  }
}

class CoreFriendRequest {
  const CoreFriendRequest({
    required this.id,
    required this.fromUserId,
    required this.name,
    required this.username,
    required this.fromFriendCode,
    required this.createdAt,
  });

  final int id;
  final String fromUserId;
  final String name;
  final String username;
  final String fromFriendCode;
  final int createdAt;

  factory CoreFriendRequest.fromJson(Map<String, dynamic> json) {
    return CoreFriendRequest(
      id: json['id'] as int? ?? 0,
      fromUserId: json['from_user_id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      username: json['username'] as String? ?? '',
      fromFriendCode: json['from_friend_code'] as String? ?? '',
      createdAt: json['created_at'] as int? ?? 0,
    );
  }
}

class CoreSnapshot {
  const CoreSnapshot({
    required this.currentUser,
    required this.messages,
    required this.conversations,
    required this.friends,
    required this.friendRequests,
    required this.loadedConversationId,
    required this.serverConnected,
    required this.lastNetworkError,
  });

  final CoreUser? currentUser;
  final List<CoreMessage> messages;
  final List<CoreConversation> conversations;
  final List<CoreFriend> friends;
  final List<CoreFriendRequest> friendRequests;
  final String loadedConversationId;
  final bool serverConnected;
  final String lastNetworkError;

  factory CoreSnapshot.fromJson(Map<String, dynamic> json) {
    return CoreSnapshot(
      currentUser: json['current_user'] is Map<String, dynamic>
          ? CoreUser.fromJson(json['current_user'] as Map<String, dynamic>)
          : null,
      messages: ((json['messages'] as List<dynamic>? ?? const <dynamic>[])
          .whereType<Map<String, dynamic>>()
          .map(CoreMessage.fromJson)
          .toList(growable: false)),
      conversations:
          ((json['conversations'] as List<dynamic>? ?? const <dynamic>[])
              .whereType<Map<String, dynamic>>()
              .map(CoreConversation.fromJson)
              .toList(growable: false)),
      friends: ((json['friends'] as List<dynamic>? ?? const <dynamic>[])
          .whereType<Map<String, dynamic>>()
          .map(CoreFriend.fromJson)
          .toList(growable: false)),
      friendRequests:
          ((json['friend_requests'] as List<dynamic>? ?? const <dynamic>[])
              .whereType<Map<String, dynamic>>()
              .map(CoreFriendRequest.fromJson)
              .toList(growable: false)),
      loadedConversationId: json['loaded_conversation_id'] as String? ?? '',
      serverConnected: json['server_connected'] as bool? ?? false,
      lastNetworkError: json['last_network_error'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'current_user': currentUser == null
          ? null
          : {
              'id': currentUser!.id,
              'name': currentUser!.name,
              'username': currentUser!.username,
              'profile_picture_url': currentUser!.profilePictureUrl,
              'friend_code': currentUser!.friendCode,
            },
      'messages': messages
          .map(
            (message) => {
              'id': message.id,
              'conversation_id': message.conversationId,
              'sender_id': message.senderId,
              'client_message_id': message.clientMessageId,
              'body': message.body,
              'created_at': message.createdAt,
              'delivery_state': message.deliveryState,
            },
          )
          .toList(growable: false),
      'conversations': conversations
          .map(
            (conversation) => {
              'id': conversation.id,
              'name': conversation.name,
            },
          )
          .toList(growable: false),
      'friends': friends
          .map(
            (friend) => {
              'id': friend.id,
              'user_id': friend.userId,
              'name': friend.name,
              'username': friend.username,
              'profile_picture_url': friend.profilePictureUrl,
              'friend_code': friend.friendCode,
              'created_at': friend.createdAt,
            },
          )
          .toList(growable: false),
      'friend_requests': friendRequests
          .map(
            (request) => {
              'id': request.id,
              'from_user_id': request.fromUserId,
              'name': request.name,
              'username': request.username,
              'from_friend_code': request.fromFriendCode,
              'created_at': request.createdAt,
            },
          )
          .toList(growable: false),
      'loaded_conversation_id': loadedConversationId,
      'server_connected': serverConnected,
      'last_network_error': lastNetworkError,
    };
  }
}

class CoreEvent {
  const CoreEvent({required this.type, this.snapshot});

  final String type;
  final CoreSnapshot? snapshot;

  factory CoreEvent.fromJson(Map<String, dynamic> json) {
    return CoreEvent(
      type: json['type'] as String? ?? '',
      snapshot: json['snapshot'] is Map<String, dynamic>
          ? CoreSnapshot.fromJson(json['snapshot'] as Map<String, dynamic>)
          : null,
    );
  }
}

class CoreApi {
  CoreApi._(DynamicLibrary library)
    : _coreCreate = library.lookupFunction<_CoreCreateNative, _CoreCreateDart>(
        'core_create',
      ),
      _coreClose = library.lookupFunction<_CoreCloseNative, _CoreCloseDart>(
        'core_close',
      ),
      _coreConfigJSON = library.lookupFunction<_CoreJSONNative, _CoreJSONDart>(
        'core_config_json',
      ),
      _coreSnapshotJSON = library
          .lookupFunction<_CoreJSONNative, _CoreJSONDart>('core_snapshot_json'),
      _coreTryLoadSession = library
          .lookupFunction<_CoreTryLoadSessionNative, _CoreTryLoadSessionDart>(
            'core_try_load_session',
          ),
      _coreLogin = library.lookupFunction<_CoreLoginNative, _CoreLoginDart>(
        'core_login',
      ),
      _coreRegister = library
          .lookupFunction<_CoreRegisterNative, _CoreRegisterDart>(
            'core_register',
          ),
      _coreLogout = library.lookupFunction<_CoreLogoutNative, _CoreLogoutDart>(
        'core_logout',
      ),
      _coreOpenConversation = library
          .lookupFunction<
            _CoreOpenConversationNative,
            _CoreOpenConversationDart
          >('core_open_conversation'),
      _coreSendMessage = library
          .lookupFunction<_CoreSendMessageNative, _CoreSendMessageDart>(
            'core_send_message',
          ),
      _coreStartDirectConversation = library
          .lookupFunction<
            _CoreStartDirectConversationNative,
            _CoreStartDirectConversationDart
          >('core_start_direct_conversation'),
      _coreCreateConversation = library
          .lookupFunction<
            _CoreCreateConversationNative,
            _CoreCreateConversationDart
          >('core_create_conversation'),
      _coreSendFriendRequest = library
          .lookupFunction<_CoreFriendActionNative, _CoreFriendActionDart>(
            'core_send_friend_request',
          ),
      _coreAcceptFriendRequest = library
          .lookupFunction<_CoreFriendActionNative, _CoreFriendActionDart>(
            'core_accept_friend_request',
          ),
      _coreRejectFriendRequest = library
          .lookupFunction<_CoreFriendActionNative, _CoreFriendActionDart>(
            'core_reject_friend_request',
          ),
      _coreSaveProfileImage = library
          .lookupFunction<
            _CoreSaveProfileImageNative,
            _CoreSaveProfileImageDart
          >('core_save_profile_image'),
      _corePollEventJSON = library
          .lookupFunction<_CorePollEventJSONNative, _CorePollEventJSONDart>(
            'core_poll_event_json',
          ),
      _coreTakeLastError = library
          .lookupFunction<_CoreTakeLastErrorNative, _CoreTakeLastErrorDart>(
            'core_take_last_error',
          ),
      _coreStringFree = library
          .lookupFunction<_CoreStringFreeNative, _CoreStringFreeDart>(
            'core_string_free',
          );
  final _CoreCreateDart _coreCreate;
  final _CoreCloseDart _coreClose;
  final _CoreJSONDart _coreConfigJSON;
  final _CoreJSONDart _coreSnapshotJSON;
  final _CoreTryLoadSessionDart _coreTryLoadSession;
  final _CoreLoginDart _coreLogin;
  final _CoreRegisterDart _coreRegister;
  final _CoreLogoutDart _coreLogout;
  final _CoreOpenConversationDart _coreOpenConversation;
  final _CoreSendMessageDart _coreSendMessage;
  final _CoreStartDirectConversationDart _coreStartDirectConversation;
  final _CoreCreateConversationDart _coreCreateConversation;
  final _CoreFriendActionDart _coreSendFriendRequest;
  final _CoreFriendActionDart _coreAcceptFriendRequest;
  final _CoreFriendActionDart _coreRejectFriendRequest;
  final _CoreSaveProfileImageDart _coreSaveProfileImage;
  final _CorePollEventJSONDart _corePollEventJSON;
  final _CoreTakeLastErrorDart _coreTakeLastError;
  final _CoreStringFreeDart _coreStringFree;

  int? _handle;

  static CoreApi load() {
    return CoreApi._(DynamicLibrary.open(_defaultLibraryPath()));
  }

  bool get isInitialized => _handle != null && _handle != 0;

  Future<void> initialize({
    String profileName = '',
    String? dataDir,
    String serverUrl = 'ws://localhost:8080/ws',
  }) async {
    final profileNamePointer = profileName.toNativeUtf8();
    final dataDirPointer = (dataDir ?? '').toNativeUtf8();
    final serverUrlPointer = serverUrl.toNativeUtf8();

    try {
      final handle = _coreCreate(
        profileNamePointer,
        dataDirPointer,
        serverUrlPointer,
      );
      if (handle == 0) {
        throw CoreApiError(_takeLastErrorMessage());
      }

      _handle = handle;
    } finally {
      calloc.free(profileNamePointer);
      calloc.free(dataDirPointer);
      calloc.free(serverUrlPointer);
    }
  }

  CoreConfig config() {
    return CoreConfig.fromJson(_readJSON(_coreConfigJSON));
  }

  CoreSnapshot snapshot() {
    return CoreSnapshot.fromJson(_readJSON(_coreSnapshotJSON));
  }

  bool tryLoadSession() {
    final handle = _requireHandle();
    final result = _coreTryLoadSession(handle);
    if (result < 0) {
      throw CoreApiError(_takeLastErrorMessage());
    }

    return result == 1;
  }

  Future<void> login({
    required String username,
    required String password,
  }) async {
    final handle = _requireHandle();
    final usernamePointer = username.toNativeUtf8();
    final passwordPointer = password.toNativeUtf8();

    try {
      final result = _coreLogin(handle, usernamePointer, passwordPointer);
      if (result != 1) {
        throw CoreApiError(_takeLastErrorMessage());
      }
    } finally {
      calloc.free(usernamePointer);
      calloc.free(passwordPointer);
    }
  }

  Future<void> register({
    required String name,
    required String username,
    required String password,
    String profilePicturePath = '',
  }) async {
    final handle = _requireHandle();
    final namePointer = name.toNativeUtf8();
    final usernamePointer = username.toNativeUtf8();
    final passwordPointer = password.toNativeUtf8();
    final profilePicturePointer = profilePicturePath.toNativeUtf8();

    try {
      final result = _coreRegister(
        handle,
        namePointer,
        usernamePointer,
        passwordPointer,
        profilePicturePointer,
      );
      if (result != 1) {
        throw CoreApiError(_takeLastErrorMessage());
      }
    } finally {
      calloc.free(namePointer);
      calloc.free(usernamePointer);
      calloc.free(passwordPointer);
      calloc.free(profilePicturePointer);
    }
  }

  Future<void> logout() async {
    final handle = _requireHandle();
    final result = _coreLogout(handle);
    if (result != 1) {
      throw CoreApiError(_takeLastErrorMessage());
    }
  }

  Future<void> openConversation(String conversationId) async {
    final handle = _requireHandle();
    final conversationPointer = conversationId.toNativeUtf8();

    try {
      final result = _coreOpenConversation(handle, conversationPointer);
      if (result != 1) {
        throw CoreApiError(_takeLastErrorMessage());
      }
    } finally {
      calloc.free(conversationPointer);
    }
  }

  Future<void> sendMessage({
    required String conversationId,
    required String body,
  }) async {
    final handle = _requireHandle();
    final conversationPointer = conversationId.toNativeUtf8();
    final bodyPointer = body.toNativeUtf8();

    try {
      final result = _coreSendMessage(handle, conversationPointer, bodyPointer);
      if (result != 1) {
        throw CoreApiError(_takeLastErrorMessage());
      }
    } finally {
      calloc.free(conversationPointer);
      calloc.free(bodyPointer);
    }
  }

  Future<String> startDirectConversation(String friendUserId) async {
    final handle = _requireHandle();
    final friendPointer = friendUserId.toNativeUtf8();

    try {
      final pointer = _coreStartDirectConversation(handle, friendPointer);
      if (pointer == nullptr) {
        throw CoreApiError(_takeLastErrorMessage());
      }

      final conversationId = pointer.toDartString();
      _coreStringFree(pointer);
      return conversationId;
    } finally {
      calloc.free(friendPointer);
    }
  }

  Future<String> createConversation({
    required String name,
    required List<String> participantUserIds,
  }) async {
    final handle = _requireHandle();
    final namePointer = name.toNativeUtf8();
    final participantJsonPointer = jsonEncode(
      participantUserIds,
    ).toNativeUtf8();

    try {
      final pointer = _coreCreateConversation(
        handle,
        namePointer,
        participantJsonPointer,
      );
      if (pointer == nullptr) {
        throw CoreApiError(_takeLastErrorMessage());
      }

      final conversationId = pointer.toDartString();
      _coreStringFree(pointer);
      return conversationId;
    } finally {
      calloc.free(namePointer);
      calloc.free(participantJsonPointer);
    }
  }

  Future<void> sendFriendRequest(String friendCode) async {
    final handle = _requireHandle();
    final friendCodePointer = friendCode.toNativeUtf8();

    try {
      final result = _coreSendFriendRequest(handle, friendCodePointer);
      if (result != 1) {
        throw CoreApiError(_takeLastErrorMessage());
      }
    } finally {
      calloc.free(friendCodePointer);
    }
  }

  Future<void> acceptFriendRequest(String fromUserId) async {
    final handle = _requireHandle();
    final fromUserIdPointer = fromUserId.toNativeUtf8();

    try {
      final result = _coreAcceptFriendRequest(handle, fromUserIdPointer);
      if (result != 1) {
        throw CoreApiError(_takeLastErrorMessage());
      }
    } finally {
      calloc.free(fromUserIdPointer);
    }
  }

  Future<void> rejectFriendRequest(String fromUserId) async {
    final handle = _requireHandle();
    final fromUserIdPointer = fromUserId.toNativeUtf8();

    try {
      final result = _coreRejectFriendRequest(handle, fromUserIdPointer);
      if (result != 1) {
        throw CoreApiError(_takeLastErrorMessage());
      }
    } finally {
      calloc.free(fromUserIdPointer);
    }
  }

  Future<String> saveProfileImage({
    required Uint8List data,
    required String filename,
  }) async {
    final handle = _requireHandle();
    final dataPointer = calloc<Uint8>(data.length);
    final filenamePointer = filename.toNativeUtf8();

    try {
      dataPointer.asTypedList(data.length).setAll(0, data);
      final pointer = _coreSaveProfileImage(
        handle,
        dataPointer,
        data.length,
        filenamePointer,
      );
      if (pointer == nullptr) {
        throw CoreApiError(_takeLastErrorMessage());
      }

      final savedPath = pointer.toDartString();
      _coreStringFree(pointer);
      return savedPath;
    } finally {
      calloc.free(dataPointer);
      calloc.free(filenamePointer);
    }
  }

  CoreEvent? pollEvent() {
    final handle = _requireHandle();
    final pointer = _corePollEventJSON(handle);
    if (pointer == nullptr) {
      return null;
    }

    final eventJSON = pointer.toDartString();
    _coreStringFree(pointer);
    final decoded = jsonDecode(eventJSON);
    if (decoded is! Map<String, dynamic>) {
      throw CoreApiError('unexpected event payload from core-go');
    }

    return CoreEvent.fromJson(decoded);
  }

  void dispose() {
    final handle = _handle;
    if (handle == null || handle == 0) {
      return;
    }

    _coreClose(handle);
    _handle = null;
  }

  Map<String, dynamic> _readJSON(_CoreJSONDart reader) {
    final handle = _requireHandle();
    final pointer = reader(handle);
    if (pointer == nullptr) {
      throw CoreApiError(_takeLastErrorMessage());
    }

    final jsonString = pointer.toDartString();
    _coreStringFree(pointer);
    final decoded = jsonDecode(jsonString);
    if (decoded is! Map<String, dynamic>) {
      throw CoreApiError('unexpected JSON payload from core-go');
    }

    return decoded;
  }

  String _takeLastErrorMessage() {
    final pointer = _coreTakeLastError();
    if (pointer == nullptr) {
      return 'core-go returned an unknown error';
    }

    final value = pointer.toDartString();
    _coreStringFree(pointer);
    return value;
  }

  int _requireHandle() {
    final handle = _handle;
    if (handle == null || handle == 0) {
      throw CoreApiError('core-go is not initialized');
    }

    return handle;
  }

  static String _defaultLibraryPath() {
    if (Platform.isWindows) {
      return 'albz_core.dll';
    }
    if (Platform.isLinux) {
      return 'libalbz_core.so';
    }
    if (Platform.isMacOS) {
      return 'libalbz_core.dylib';
    }

    throw UnsupportedError('Unsupported platform for core-go library loading');
  }
}
