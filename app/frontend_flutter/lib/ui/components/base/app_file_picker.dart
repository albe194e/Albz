import 'dart:io';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';
import '../../theme/responsive.dart';
import 'app_button.dart';

class PickedFileData {
  const PickedFileData({required this.name, required this.bytes, this.path});

  final String name;
  final Uint8List bytes;
  final String? path;
}

class AppFilePicker extends StatefulWidget {
  const AppFilePicker({
    super.key,
    required this.label,
    required this.buttonLabel,
    required this.onPicked,
    this.type = FileType.any,
    this.allowedExtensions,
  });

  final String label;
  final String buttonLabel;
  final FileType type;
  final List<String>? allowedExtensions;
  final Future<void> Function(PickedFileData file) onPicked;

  @override
  State<AppFilePicker> createState() => _AppFilePickerState();
}

class _AppFilePickerState extends State<AppFilePicker> {
  bool _picking = false;
  String _selectedFileName = 'No file selected';
  String _errorMessage = '';

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    final narrow = AppBreakpoints.isNarrow(width);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(widget.label, style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 10),
        narrow
            ? Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    _selectedFileName,
                    style: const TextStyle(color: AppColors.textSecondary),
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 12),
                  AppButton.secondary(
                    label: _picking ? 'Picking...' : widget.buttonLabel,
                    onPressed: _picking ? null : _pickFile,
                  ),
                ],
              )
            : Row(
                children: [
                  Expanded(
                    child: Text(
                      _selectedFileName,
                      style: const TextStyle(color: AppColors.textSecondary),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  const SizedBox(width: 12),
                  SizedBox(
                    width: 200,
                    child: AppButton.secondary(
                      label: _picking ? 'Picking...' : widget.buttonLabel,
                      onPressed: _picking ? null : _pickFile,
                    ),
                  ),
                ],
              ),
        if (_errorMessage.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text(_errorMessage, style: const TextStyle(color: AppColors.error)),
        ],
      ],
    );
  }

  Future<void> _pickFile() async {
    setState(() {
      _picking = true;
      _errorMessage = '';
    });

    try {
      final result = await FilePicker.platform.pickFiles(
        type: widget.type,
        allowedExtensions: widget.allowedExtensions,
        withData: true,
      );
      if (result == null || result.files.isEmpty) {
        return;
      }

      final file = result.files.single;
      final bytes = file.bytes ?? await _loadBytesFromPath(file.path);
      if (bytes == null || bytes.isEmpty) {
        throw Exception('Picked file is empty');
      }

      await widget.onPicked(
        PickedFileData(name: file.name, bytes: bytes, path: file.path),
      );

      setState(() {
        _selectedFileName = file.name;
      });
    } catch (error) {
      setState(() {
        _errorMessage = error.toString();
      });
    } finally {
      if (mounted) {
        setState(() {
          _picking = false;
        });
      }
    }
  }

  Future<Uint8List?> _loadBytesFromPath(String? path) async {
    if (path == null || path.isEmpty) {
      return null;
    }

    return File(path).readAsBytes();
  }
}
