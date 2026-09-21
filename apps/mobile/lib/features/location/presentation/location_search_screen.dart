import 'dart:async';

import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../weather/presentation/weather_scene_resolver.dart';
import '../data/models/location_result.dart';
import '../domain/supported_country.dart';
import 'location_search_controller.dart';

class LocationSearchScreen extends StatefulWidget {
  const LocationSearchScreen({
    required this.controller,
    required this.onLocationSelected,
    this.recentLocations = const [],
    this.backgroundAsset = WeatherSceneResolver.dayAsset,
    this.backgroundAssetResolver,
    super.key,
  });

  final LocationSearchController controller;
  final ValueChanged<LocationResult> onLocationSelected;
  final List<LocationResult> recentLocations;
  final String backgroundAsset;
  final ValueGetter<String>? backgroundAssetResolver;

  @override
  State<LocationSearchScreen> createState() => _LocationSearchScreenState();
}

class _LocationSearchScreenState extends State<LocationSearchScreen>
    with WidgetsBindingObserver {
  final _queryController = TextEditingController();
  late SupportedCountry _country;
  Timer? _sceneClock;
  String? _queryError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _country = widget.controller.state.country;
    _sceneClock = Timer.periodic(const Duration(minutes: 1), (_) {
      if (mounted) {
        setState(() {});
      }
    });
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && mounted) {
      setState(() {});
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _sceneClock?.cancel();
    _queryController.dispose();
    super.dispose();
  }

  String get _backgroundAsset =>
      widget.backgroundAssetResolver?.call() ?? widget.backgroundAsset;

  Future<void> _chooseCountry() async {
    final selected = await showModalBottomSheet<SupportedCountry>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      barrierColor: const Color(0x9905182B),
      builder: (context) => _CountryPickerSheet(selected: _country),
    );

    if (selected != null && mounted) {
      setState(() => _country = selected);
    }
  }

  void _search() {
    if (_queryController.text.trim().isEmpty) {
      setState(() => _queryError = 'Enter a city name');
      return;
    }

    FocusManager.instance.primaryFocus?.unfocus();
    setState(() => _queryError = null);
    widget.controller.search(query: _queryController.text, country: _country);
  }

  Future<void> _select(LocationResult location) async {
    final selected = await widget.controller.save(location);
    if (selected != null && mounted) {
      widget.onLocationSelected(selected);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: false,
      body: Stack(
        fit: StackFit.expand,
        children: [
          Image.asset(
            _backgroundAsset,
            key: const ValueKey('location-scene-background'),
            fit: BoxFit.cover,
          ),
          const DecoratedBox(
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [Color(0x3D05182B), Color(0x7505182B)],
              ),
            ),
          ),
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(20, 18, 20, 20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Row(
                    children: [
                      Icon(
                        Icons.location_on_rounded,
                        color: Colors.white,
                        size: 30,
                      ),
                      SizedBox(width: 10),
                      Text(
                        'Choose a city',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 30,
                          fontWeight: FontWeight.w700,
                          letterSpacing: -0.8,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 14),
                  Container(
                    key: const ValueKey('location-search-form'),
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: AppColors.card,
                      borderRadius: BorderRadius.circular(28),
                      border: Border.all(color: AppColors.cardBorder),
                      boxShadow: const [
                        BoxShadow(
                          color: AppColors.cardShadow,
                          blurRadius: 18,
                          offset: Offset(0, 8),
                        ),
                      ],
                    ),
                    child: Column(
                      children: [
                        _CountryField(country: _country, onTap: _chooseCountry),
                        const SizedBox(height: 12),
                        TextField(
                          controller: _queryController,
                          textInputAction: TextInputAction.search,
                          onSubmitted: (_) => _search(),
                          decoration: _fieldDecoration(
                            label: 'City',
                            icon: Icons.search_rounded,
                            hint: 'e.g. Jakarta',
                            errorText: _queryError,
                          ),
                        ),
                        const SizedBox(height: 12),
                        SizedBox(
                          width: double.infinity,
                          height: 52,
                          child: FilledButton.icon(
                            onPressed: _search,
                            style: FilledButton.styleFrom(
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(18),
                              ),
                            ),
                            icon: const Icon(Icons.search_rounded),
                            label: const Text('Search cities'),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  Expanded(
                    child: Container(
                      key: const ValueKey('location-results-panel'),
                      width: double.infinity,
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: AppColors.card,
                        borderRadius: BorderRadius.circular(28),
                        border: Border.all(color: AppColors.cardBorder),
                        boxShadow: const [
                          BoxShadow(
                            color: AppColors.cardShadow,
                            blurRadius: 18,
                            offset: Offset(0, 8),
                          ),
                        ],
                      ),
                      child: ListenableBuilder(
                        listenable: widget.controller,
                        builder: (context, _) {
                          return _SearchBody(
                            state: widget.controller.state,
                            recentLocations: widget.recentLocations,
                            onRetry: widget.controller.retry,
                            onSelect: _select,
                          );
                        },
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  InputDecoration _fieldDecoration({
    required String label,
    required IconData icon,
    String? hint,
    String? errorText,
  }) {
    return InputDecoration(
      labelText: label,
      hintText: hint,
      errorText: errorText,
      filled: true,
      fillColor: Colors.white.withValues(alpha: 0.78),
      prefixIcon: Icon(icon),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(18),
        borderSide: BorderSide.none,
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(18),
        borderSide: const BorderSide(color: Color(0x1A17253D)),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(18),
        borderSide: const BorderSide(color: AppColors.primary, width: 1.5),
      ),
      errorBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(18),
        borderSide: BorderSide(color: Theme.of(context).colorScheme.error),
      ),
      focusedErrorBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(18),
        borderSide: BorderSide(
          color: Theme.of(context).colorScheme.error,
          width: 1.5,
        ),
      ),
    );
  }
}

class _CountryField extends StatelessWidget {
  const _CountryField({required this.country, required this.onTap});

  final SupportedCountry country;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white.withValues(alpha: 0.78),
      borderRadius: BorderRadius.circular(18),
      child: InkWell(
        key: const ValueKey('country-selector'),
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Container(
          height: 62,
          padding: const EdgeInsets.symmetric(horizontal: 16),
          decoration: BoxDecoration(
            border: Border.all(color: const Color(0x1A17253D)),
            borderRadius: BorderRadius.circular(18),
          ),
          child: Row(
            children: [
              _CountryBadge(country: country, small: true),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Country',
                      style: TextStyle(color: AppColors.mutedInk, fontSize: 12),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      country.displayName,
                      style: const TextStyle(
                        color: AppColors.ink,
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
              ),
              const Icon(
                Icons.keyboard_arrow_down_rounded,
                color: AppColors.mutedInk,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _CountryPickerSheet extends StatelessWidget {
  const _CountryPickerSheet({required this.selected});

  final SupportedCountry selected;

  @override
  Widget build(BuildContext context) {
    return FractionallySizedBox(
      heightFactor: 0.78,
      child: Container(
        key: const ValueKey('country-picker-sheet'),
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
        decoration: const BoxDecoration(
          color: Color(0xFFF7FAFD),
          borderRadius: BorderRadius.vertical(top: Radius.circular(30)),
        ),
        child: SafeArea(
          top: false,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 42,
                  height: 5,
                  decoration: BoxDecoration(
                    color: const Color(0xFFD5DCE7),
                    borderRadius: BorderRadius.circular(99),
                  ),
                ),
              ),
              const SizedBox(height: 20),
              const Text(
                'Choose a country',
                style: TextStyle(
                  color: AppColors.ink,
                  fontSize: 24,
                  fontWeight: FontWeight.w700,
                  letterSpacing: -0.6,
                ),
              ),
              const SizedBox(height: 4),
              const Text(
                'Select the region where you want to search.',
                style: TextStyle(color: AppColors.mutedInk, fontSize: 13),
              ),
              const SizedBox(height: 18),
              Expanded(
                child: ListView.separated(
                  itemCount: SupportedCountry.values.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 8),
                  itemBuilder: (context, index) {
                    final country = SupportedCountry.values[index];
                    final isSelected = country == selected;

                    return Material(
                      color: isSelected
                          ? const Color(0xFFE4EEFF)
                          : Colors.white,
                      borderRadius: BorderRadius.circular(18),
                      child: InkWell(
                        key: ValueKey('country-option-${country.code}'),
                        onTap: () => Navigator.of(context).pop(country),
                        borderRadius: BorderRadius.circular(18),
                        child: Padding(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 14,
                            vertical: 11,
                          ),
                          child: Row(
                            children: [
                              _CountryBadge(country: country),
                              const SizedBox(width: 14),
                              Expanded(
                                child: Text(
                                  country.displayName,
                                  style: TextStyle(
                                    color: AppColors.ink,
                                    fontSize: 16,
                                    fontWeight: isSelected
                                        ? FontWeight.w700
                                        : FontWeight.w600,
                                  ),
                                ),
                              ),
                              if (isSelected)
                                const Icon(
                                  Icons.check_circle_rounded,
                                  color: AppColors.primary,
                                )
                              else
                                const Icon(
                                  Icons.chevron_right_rounded,
                                  color: AppColors.mutedInk,
                                ),
                            ],
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _CountryBadge extends StatelessWidget {
  const _CountryBadge({required this.country, this.small = false});

  final SupportedCountry country;
  final bool small;

  @override
  Widget build(BuildContext context) {
    final size = small ? 34.0 : 42.0;
    return Container(
      width: size,
      height: size,
      alignment: Alignment.center,
      decoration: const BoxDecoration(
        color: Color(0xFFDCE8FF),
        shape: BoxShape.circle,
      ),
      child: Text(
        country.code,
        style: TextStyle(
          color: AppColors.primary,
          fontSize: small ? 11 : 12,
          fontWeight: FontWeight.w800,
          letterSpacing: 0.3,
        ),
      ),
    );
  }
}

class _SearchBody extends StatelessWidget {
  const _SearchBody({
    required this.state,
    required this.recentLocations,
    required this.onRetry,
    required this.onSelect,
  });

  final LocationSearchState state;
  final List<LocationResult> recentLocations;
  final VoidCallback onRetry;
  final ValueChanged<LocationResult> onSelect;

  @override
  Widget build(BuildContext context) {
    return switch (state.status) {
      LocationSearchStatus.initial =>
        recentLocations.isEmpty
            ? const _Message(
                icon: Icons.location_city_rounded,
                title: 'Find your city',
                message: 'Choose a country and search by city name.',
              )
            : _RecentLocations(locations: recentLocations, onSelect: onSelect),
      LocationSearchStatus.searching => const Center(
        child: CircularProgressIndicator(),
      ),
      LocationSearchStatus.empty => _Message(
        icon: Icons.travel_explore_rounded,
        title: 'No cities found',
        message: 'Try a different spelling for “${state.query}”.',
      ),
      LocationSearchStatus.failure => _Failure(
        message: state.errorMessage ?? 'Locations could not be loaded.',
        onRetry: onRetry,
      ),
      LocationSearchStatus.results => _Results(
        state: state,
        onSelect: onSelect,
      ),
    };
  }
}

class _RecentLocations extends StatelessWidget {
  const _RecentLocations({required this.locations, required this.onSelect});

  final List<LocationResult> locations;
  final ValueChanged<LocationResult> onSelect;

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      itemCount: locations.length + 1,
      separatorBuilder: (_, index) => SizedBox(height: index == 0 ? 4 : 8),
      itemBuilder: (context, index) {
        if (index == 0) {
          return Padding(
            padding: const EdgeInsets.fromLTRB(4, 0, 4, 8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Row(
                  children: [
                    Icon(
                      Icons.history_rounded,
                      color: AppColors.primary,
                      size: 21,
                    ),
                    SizedBox(width: 8),
                    Text(
                      'Recent cities',
                      style: TextStyle(
                        color: AppColors.ink,
                        fontSize: 18,
                        fontWeight: FontWeight.w700,
                        letterSpacing: -0.3,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                const Text(
                  'Tap a city to load its latest forecast.',
                  style: TextStyle(
                    color: AppColors.mutedInk,
                    fontSize: 13,
                    height: 1.3,
                  ),
                ),
              ],
            ),
          );
        }

        final location = locations[index - 1];
        return _LocationCard(
          key: ValueKey('recent-location-${location.locationId}'),
          location: location,
          icon: Icons.history_rounded,
          onTap: () => onSelect(location),
        );
      },
    );
  }
}

class _Results extends StatelessWidget {
  const _Results({required this.state, required this.onSelect});

  final LocationSearchState state;
  final ValueChanged<LocationResult> onSelect;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (state.errorMessage case final message?) ...[
          MaterialBanner(
            content: Text(message),
            actions: const [SizedBox.shrink()],
          ),
          const SizedBox(height: 8),
        ],
        Padding(
          padding: const EdgeInsets.fromLTRB(4, 0, 4, 10),
          child: Row(
            children: [
              const Icon(
                Icons.travel_explore_rounded,
                color: AppColors.primary,
                size: 21,
              ),
              const SizedBox(width: 8),
              const Expanded(
                child: Text(
                  'Search results',
                  style: TextStyle(
                    color: AppColors.ink,
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                    letterSpacing: -0.3,
                  ),
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 5,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFFDCE8FF),
                  borderRadius: BorderRadius.circular(99),
                ),
                child: Text(
                  '${state.results.length} result${state.results.length == 1 ? '' : 's'}',
                  style: const TextStyle(
                    color: AppColors.primary,
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: ListView.separated(
            itemCount: state.results.length,
            separatorBuilder: (_, _) => const SizedBox(height: 8),
            itemBuilder: (context, index) {
              final location = state.results[index];
              final isSaving =
                  state.savingOpenMeteoLocationId ==
                  location.openMeteoLocationId;

              return _LocationCard(
                key: ValueKey('location-${location.openMeteoLocationId}'),
                location: location,
                icon: Icons.location_on_rounded,
                isLoading: isSaving,
                onTap: state.isSaving ? null : () => onSelect(location),
              );
            },
          ),
        ),
      ],
    );
  }
}

class _LocationCard extends StatelessWidget {
  const _LocationCard({
    required this.location,
    required this.icon,
    required this.onTap,
    this.isLoading = false,
    super.key,
  });

  final LocationResult location;
  final IconData icon;
  final VoidCallback? onTap;
  final bool isLoading;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white.withValues(alpha: 0.82),
      borderRadius: BorderRadius.circular(20),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(20),
        child: Container(
          constraints: const BoxConstraints(minHeight: 72),
          padding: const EdgeInsets.fromLTRB(14, 10, 12, 10),
          decoration: BoxDecoration(
            border: Border.all(color: const Color(0x1205182B)),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Row(
            children: [
              Container(
                width: 44,
                height: 44,
                decoration: const BoxDecoration(
                  color: Color(0xFFDCE8FF),
                  shape: BoxShape.circle,
                ),
                child: Icon(icon, color: AppColors.primary, size: 22),
              ),
              const SizedBox(width: 13),
              Expanded(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      location.city,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: AppColors.ink,
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        letterSpacing: -0.2,
                      ),
                    ),
                    const SizedBox(height: 3),
                    Text(
                      _locationSubtitle(location),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: AppColors.mutedInk,
                        fontSize: 13,
                        height: 1.25,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              if (isLoading)
                const SizedBox.square(
                  dimension: 22,
                  child: CircularProgressIndicator(strokeWidth: 2.5),
                )
              else
                Container(
                  width: 32,
                  height: 32,
                  decoration: const BoxDecoration(
                    color: Color(0xFFF0F5FC),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.chevron_right_rounded,
                    color: AppColors.ink,
                    size: 21,
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

String _locationSubtitle(LocationResult location) {
  final area = location.administrativeArea;
  if (area == null || area.isEmpty || area == location.city) {
    return location.country;
  }

  return '$area, ${location.country}';
}

class _Failure extends StatelessWidget {
  const _Failure({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return _Message(
      icon: Icons.cloud_off_rounded,
      title: 'Search unavailable',
      message: message,
      action: OutlinedButton.icon(
        onPressed: onRetry,
        icon: const Icon(Icons.refresh_rounded),
        label: const Text('Try again'),
      ),
    );
  }
}

class _Message extends StatelessWidget {
  const _Message({
    required this.icon,
    required this.title,
    required this.message,
    this.action,
  });

  final IconData icon;
  final String title;
  final String message;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 52, color: Theme.of(context).colorScheme.primary),
            const SizedBox(height: 16),
            Text(title, style: Theme.of(context).textTheme.headlineSmall),
            const SizedBox(height: 8),
            Text(
              message,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            if (action case final action?) ...[
              const SizedBox(height: 20),
              action,
            ],
          ],
        ),
      ),
    );
  }
}
