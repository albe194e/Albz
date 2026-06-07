import 'package:flutter/material.dart';

import '../../app/app_page.dart';
import '../../app/app_scope.dart';
import '../components/base/app_button.dart';
import '../components/base/app_card.dart';
import '../components/base/app_input.dart';
import '../theme/app_theme.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: AppCard(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Welcome back',
                  style: Theme.of(context).textTheme.headlineMedium,
                ),
                const SizedBox(height: 10),
                const Text(
                  'Sign in to reconnect with your local conversations.',
                  style: TextStyle(color: AppColors.textSecondary),
                ),
                const SizedBox(height: 20),
                AppInput(
                  label: 'Username',
                  hint: 'Enter username',
                  controller: _usernameController,
                ),
                const SizedBox(height: 12),
                AppInput(
                  label: 'Password',
                  hint: 'Enter password',
                  controller: _passwordController,
                  obscureText: true,
                ),
                const SizedBox(height: 18),
                AppButton.primary(
                  label: 'Login',
                  onPressed: controller.isAuthInFlight
                      ? null
                      : () => controller.login(
                          username: _usernameController.text,
                          password: _passwordController.text,
                        ),
                ),
                const SizedBox(height: 12),
                AppButton.secondary(
                  label: 'Back',
                  onPressed: () => controller.navigateTo(AppPage.landing),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
