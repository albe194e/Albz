import 'package:flutter/material.dart';

import '../../app/app_page.dart';
import '../../app/app_scope.dart';
import '../components/base/app_button.dart';
import '../components/base/app_card.dart';
import '../components/base/app_image_picker.dart';
import '../components/base/app_input.dart';
import '../theme/app_theme.dart';

class RegisterPage extends StatefulWidget {
  const RegisterPage({super.key});

  @override
  State<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends State<RegisterPage> {
  final TextEditingController _nameController = TextEditingController();
  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  String _profilePicturePath = '';

  @override
  void dispose() {
    _nameController.dispose();
    _usernameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 560),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: AppCard(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Create account',
                  style: Theme.of(context).textTheme.headlineMedium,
                ),
                const SizedBox(height: 10),
                const Text(
                  'Set up a local-first identity and start building your network.',
                  style: TextStyle(color: AppColors.textSecondary),
                ),
                const SizedBox(height: 20),
                AppInput(
                  label: 'Name',
                  hint: 'Display name',
                  controller: _nameController,
                ),
                const SizedBox(height: 12),
                AppInput(
                  label: 'Username',
                  hint: 'Choose a username',
                  controller: _usernameController,
                ),
                const SizedBox(height: 12),
                AppInput(
                  label: 'Password',
                  hint: 'Create a password',
                  controller: _passwordController,
                  obscureText: true,
                ),
                const SizedBox(height: 12),
                AppImagePicker(
                  label: 'Profile picture',
                  buttonLabel: 'Upload profile picture',
                  onPicked: (file) async {
                    final savedPath = await controller.saveProfileImage(
                      data: file.bytes,
                      filename: file.name,
                    );
                    _profilePicturePath = savedPath;
                  },
                ),
                const SizedBox(height: 18),
                AppButton.primary(
                  label: 'Create Account',
                  onPressed: controller.isAuthInFlight
                      ? null
                      : () => controller.register(
                          name: _nameController.text,
                          username: _usernameController.text,
                          password: _passwordController.text,
                          profilePicturePath: _profilePicturePath,
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
