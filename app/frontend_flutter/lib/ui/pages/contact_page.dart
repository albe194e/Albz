import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/app_controller.dart';
import '../../app/app_scope.dart';
import '../../core/core_api.dart';
import '../components/base/app_button.dart';
import '../components/base/contact_card.dart';
import '../components/base/section.dart';
import '../components/base/mobile_page_header.dart';
import '../components/contacts/contact_qr_dialog.dart';
import '../theme/app_theme.dart';

enum _ContactTab { contacts, requests }

class ContactPage extends StatefulWidget {
  const ContactPage({super.key});

  @override
  State<ContactPage> createState() => _ContactPageState();
}

class _ContactPageState extends State<ContactPage> {
  _ContactTab _selectedTab = _ContactTab.contacts;
  final TextEditingController _contactCodeController = TextEditingController(
    text: 'HADDLE-',
  );
  bool _addContactDialogOpen = false;
  bool _pendingDialogScheduled = false;

  @override
  void dispose() {
    _contactCodeController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);
    final width = MediaQuery.sizeOf(context).width;
    final mobile = width < 760;
    _schedulePendingContactDialogIfNeeded(controller);

    return Padding(
      padding: EdgeInsets.all(mobile ? 16 : 24),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 1080),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              mobile
                  ? const MobilePageHeader(title: 'Contacts')
                  : Text(
                      'Contacts',
                      style: Theme.of(context).textTheme.headlineMedium,
                    ),
              const SizedBox(height: 18),
              _ContactSubnav(
                selectedTab: _selectedTab,
                onTabSelected: (tab) => setState(() => _selectedTab = tab),
                onAddContact: () => _showAddContactDialog(context),
              ),
              const SizedBox(height: 22),
              Expanded(
                child: SingleChildScrollView(
                  child: switch (_selectedTab) {
                    _ContactTab.contacts => _ContactsGrid(
                      contacts: controller.contacts,
                    ),
                    _ContactTab.requests => _RequestsSection(
                      requests: controller.contactRequests,
                    ),
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _showAddContactDialog(BuildContext context) async {
    if (_addContactDialogOpen) {
      return;
    }

    _addContactDialogOpen = true;
    final controller = AppScope.of(context);
    unawaited(controller.loadContactQrCode());

    try {
      await showDialog<void>(
        context: context,
        builder: (_) =>
            ContactQrDialog(contactCodeController: _contactCodeController),
      );
    } finally {
      _addContactDialogOpen = false;
    }
  }

  void _schedulePendingContactDialogIfNeeded(AppController controller) {
    if (_pendingDialogScheduled ||
        _addContactDialogOpen ||
        controller.pendingContactCode.isEmpty) {
      return;
    }

    _pendingDialogScheduled = true;
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      _pendingDialogScheduled = false;
      if (!mounted || _addContactDialogOpen) {
        return;
      }

      final pendingContactCode = controller.takePendingContactCode();
      if (pendingContactCode.isEmpty) {
        return;
      }

      _contactCodeController.text = pendingContactCode;
      await _showAddContactDialog(context);
    });
  }
}

class _ContactSubnav extends StatelessWidget {
  const _ContactSubnav({
    required this.selectedTab,
    required this.onTabSelected,
    required this.onAddContact,
  });

  final _ContactTab selectedTab;
  final ValueChanged<_ContactTab> onTabSelected;
  final VoidCallback onAddContact;

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    final mobile = width < 760;

    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: AppColors.cardMuted,
        borderRadius: BorderRadius.circular(22),
      ),
      child: mobile
          ? Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    _ContactTabButton(
                      label: 'Contacts',
                      selected: selectedTab == _ContactTab.contacts,
                      onPressed: () => onTabSelected(_ContactTab.contacts),
                    ),
                    const SizedBox(width: 8),
                    _ContactTabButton(
                      label: 'Requests',
                      selected: selectedTab == _ContactTab.requests,
                      onPressed: () => onTabSelected(_ContactTab.requests),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                AppButton.primary(
                  label: 'Add Contact',
                  onPressed: onAddContact,
                ),
              ],
            )
          : Row(
              children: [
                _ContactTabButton(
                  label: 'Contacts',
                  selected: selectedTab == _ContactTab.contacts,
                  onPressed: () => onTabSelected(_ContactTab.contacts),
                ),
                const SizedBox(width: 8),
                _ContactTabButton(
                  label: 'Requests',
                  selected: selectedTab == _ContactTab.requests,
                  onPressed: () => onTabSelected(_ContactTab.requests),
                ),
                const Spacer(),
                SizedBox(
                  width: 160,
                  child: AppButton.primary(
                    label: 'Add Contact',
                    onPressed: onAddContact,
                  ),
                ),
              ],
            ),
    );
  }
}

class _ContactTabButton extends StatelessWidget {
  const _ContactTabButton({
    required this.label,
    required this.selected,
    required this.onPressed,
  });

  final String label;
  final bool selected;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: selected ? AppColors.accentSoft : Colors.transparent,
      borderRadius: BorderRadius.circular(16),
      child: InkWell(
        onTap: onPressed,
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
          child: Text(
            label,
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              color: selected ? AppColors.textPrimary : AppColors.textSecondary,
            ),
          ),
        ),
      ),
    );
  }
}

class _ContactsGrid extends StatelessWidget {
  const _ContactsGrid({required this.contacts});

  final List<CoreContact> contacts;

  @override
  Widget build(BuildContext context) {
    if (contacts.isEmpty) {
      return const Section(
        title: 'Contacts',
        child: Text(
          'No contacts',
          style: TextStyle(color: AppColors.textMuted),
        ),
      );
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        var columns = 4;
        if (constraints.maxWidth < 980) {
          columns = 3;
        }
        if (constraints.maxWidth < 760) {
          columns = 2;
        }

        return GridView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: columns,
            crossAxisSpacing: 18,
            mainAxisSpacing: 18,
            childAspectRatio: 1,
          ),
          itemCount: contacts.length,
          itemBuilder: (context, index) {
            final contact = contacts[index];
            return ContactCard(name: contact.name, onTap: () {});
          },
        );
      },
    );
  }
}

class _RequestsSection extends StatelessWidget {
  const _RequestsSection({required this.requests});

  final List<CoreContactRequest> requests;

  @override
  Widget build(BuildContext context) {
    return Section(
      title: 'Requests',
      child: requests.isEmpty
          ? const Text(
              'No requests',
              style: TextStyle(color: AppColors.textMuted),
            )
          : Column(
              children: requests
                  .map((request) => _ContactRequestRow(request: request))
                  .toList(growable: false),
            ),
    );
  }
}

class _ContactRequestRow extends StatelessWidget {
  const _ContactRequestRow({required this.request});

  final CoreContactRequest request;

  @override
  Widget build(BuildContext context) {
    final controller = AppScope.of(context);

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  request.name,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                if (request.username.isNotEmpty)
                  Text(
                    '@${request.username}',
                    style: const TextStyle(color: AppColors.textSecondary),
                  ),
              ],
            ),
          ),
          SizedBox(
            width: 110,
            child: AppButton.primary(
              label: 'Accept',
              onPressed: () =>
                  controller.acceptContactRequest(request.fromUserId),
            ),
          ),
          const SizedBox(width: 10),
          SizedBox(
            width: 110,
            child: AppButton.secondary(
              label: 'Reject',
              onPressed: () =>
                  controller.rejectContactRequest(request.fromUserId),
            ),
          ),
        ],
      ),
    );
  }
}
