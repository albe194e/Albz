import 'package:file_picker/file_picker.dart';
import 'package:flutter/widgets.dart';

import 'app_file_picker.dart';

class AppImagePicker extends StatelessWidget {
  const AppImagePicker({
    super.key,
    required this.label,
    required this.buttonLabel,
    required this.onPicked,
  });

  final String label;
  final String buttonLabel;
  final Future<void> Function(PickedFileData file) onPicked;

  @override
  Widget build(BuildContext context) {
    return AppFilePicker(
      label: label,
      buttonLabel: buttonLabel,
      type: FileType.custom,
      allowedExtensions: const ['png', 'jpg', 'jpeg', 'webp', 'gif'],
      onPicked: onPicked,
    );
  }
}
