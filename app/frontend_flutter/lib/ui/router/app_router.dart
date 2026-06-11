import 'package:flutter/material.dart';

import '../../app/app_controller.dart';
import '../../app/app_page.dart';
import '../pages/chat_page.dart';
import '../pages/contact_page.dart';
import '../pages/landing_page.dart';
import '../pages/login_page.dart';
import '../pages/profile_page.dart';
import '../pages/register_page.dart';

class AppRouter extends StatelessWidget {
  const AppRouter({super.key, required this.controller});

  final AppController controller;

  @override
  Widget build(BuildContext context) {
    switch (controller.page) {
      case AppPage.landing:
        return const LandingPage();
      case AppPage.login:
        return const LoginPage();
      case AppPage.register:
        return const RegisterPage();
      case AppPage.chat:
        return const ChatPage();
      case AppPage.profile:
        return const ProfilePage();
      case AppPage.contacts:
        return const ContactPage();
    }
  }
}
