import 'package:flutter/material.dart';

import '../../../app/app_scope.dart';
import '../../theme/app_theme.dart';
import '../base/app_button.dart';
import '../base/app_input.dart';
import '../base/byte_image_panel.dart';

class ContactQrDialog extends StatelessWidget {
  const ContactQrDialog({super.key, required this.contactCodeController});

  final TextEditingController contactCodeController;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final currentUser = controller.currentUser;

    return Dialog(
      backgroundColor: AppColors.card,
      insetPadding: const EdgeInsets.all(24),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                'Add Contact',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 18),
              Center(
                child: ByteImagePanel(
                  bytes: controller.contactQrCodeBytes,
                  loading: controller.isContactQrCodeLoading,
                  emptyLabel: 'Unable to load contact QR code',
                  size: 210,
                ),
              ),
              const SizedBox(height: 16),
              Text(
                currentUser?.contactCode.isNotEmpty == true
                    ? currentUser!.contactCode
                    : 'Contact code unavailable',
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 16),
              AppInput(
                label: 'Contact code',
                hint: 'Enter a contact code',
                controller: contactCodeController,
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  const SizedBox(width: 12),
                  Expanded(
                    child: AppButton.primary(
                      label: 'Send',
                      onPressed: () async {
                        await controller.sendContactRequest(
                          contactCodeController.text,
                        );
                        if (!context.mounted) {
                          return;
                        }
                        if (controller.errorMessage.isEmpty) {
                          contactCodeController.text = 'HADDLE-';
                          Navigator.of(context).pop();
                        }
                      },
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
