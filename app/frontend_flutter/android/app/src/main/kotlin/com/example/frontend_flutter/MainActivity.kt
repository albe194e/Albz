package com.example.frontend_flutter

import android.content.Intent
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.EventChannel
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {
    private var deepLinkEvents: EventChannel.EventSink? = null

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        MethodChannel(
            flutterEngine.dartExecutor.binaryMessenger,
            METHOD_CHANNEL_NAME,
        ).setMethodCallHandler { call, result ->
            when (call.method) {
                "getInitialLink" -> result.success(intent?.dataString)
                else -> result.notImplemented()
            }
        }

        EventChannel(
            flutterEngine.dartExecutor.binaryMessenger,
            EVENT_CHANNEL_NAME,
        ).setStreamHandler(
            object : EventChannel.StreamHandler {
                override fun onListen(arguments: Any?, events: EventChannel.EventSink) {
                    deepLinkEvents = events
                }

                override fun onCancel(arguments: Any?) {
                    deepLinkEvents = null
                }
            },
        )
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        publishDeepLink(intent)
    }

    private fun publishDeepLink(intent: Intent?) {
        val deepLink = intent?.dataString ?: return
        if (deepLink.isBlank()) {
            return
        }

        deepLinkEvents?.success(deepLink)
    }

    companion object {
        private const val METHOD_CHANNEL_NAME = "haddle/deep_links/methods"
        private const val EVENT_CHANNEL_NAME = "haddle/deep_links/events"
    }
}
