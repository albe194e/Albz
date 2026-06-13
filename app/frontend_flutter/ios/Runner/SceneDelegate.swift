import Flutter
import UIKit

class SceneDelegate: FlutterSceneDelegate {
  private var deepLinkEventSink: FlutterEventSink?
  private var initialLink: String?

  override func scene(
    _ scene: UIScene,
    willConnectTo session: UISceneSession,
    options connectionOptions: UIScene.ConnectionOptions
  ) {
    initialLink = connectionOptions.urlContexts.first?.url.absoluteString
    super.scene(scene, willConnectTo: session, options: connectionOptions)
    configureDeepLinkChannels()
  }

  override func scene(_ scene: UIScene, openURLContexts URLContexts: Set<UIOpenURLContext>) {
    super.scene(scene, openURLContexts: URLContexts)

    guard let deepLink = URLContexts.first?.url.absoluteString,
          !deepLink.isEmpty
    else {
      return
    }

    deepLinkEventSink?(deepLink)
  }

  private func configureDeepLinkChannels() {
    guard let controller = window?.rootViewController as? FlutterViewController else {
      return
    }

    let methodChannel = FlutterMethodChannel(
      name: "haddle/deep_links/methods",
      binaryMessenger: controller.binaryMessenger
    )
    methodChannel.setMethodCallHandler { [weak self] call, result in
      switch call.method {
      case "getInitialLink":
        result(self?.initialLink)
      default:
        result(FlutterMethodNotImplemented)
      }
    }

    let eventChannel = FlutterEventChannel(
      name: "haddle/deep_links/events",
      binaryMessenger: controller.binaryMessenger
    )
    eventChannel.setStreamHandler(
      DeepLinkStreamHandler(
        onListen: { [weak self] sink in
          self?.deepLinkEventSink = sink
        },
        onCancel: { [weak self] in
          self?.deepLinkEventSink = nil
        }
      )
    )
  }
}

private final class DeepLinkStreamHandler: NSObject, FlutterStreamHandler {
  init(
    onListen: @escaping (FlutterEventSink) -> Void,
    onCancel: @escaping () -> Void
  ) {
    self.onListenHandler = onListen
    self.onCancelHandler = onCancel
  }

  private let onListenHandler: (FlutterEventSink) -> Void
  private let onCancelHandler: () -> Void

  func onListen(withArguments arguments: Any?, eventSink events: @escaping FlutterEventSink) -> FlutterError? {
    onListenHandler(events)
    return nil
  }

  func onCancel(withArguments arguments: Any?) -> FlutterError? {
    onCancelHandler()
    return nil
  }
}
