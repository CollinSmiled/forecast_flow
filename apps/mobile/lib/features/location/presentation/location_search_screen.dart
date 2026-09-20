import 'package:flutter/material.dart';

import '../data/models/location_result.dart';
import '../domain/supported_country.dart';
import 'location_search_controller.dart';

class LocationSearchScreen extends StatefulWidget {
  const LocationSearchScreen({
    required this.controller,
    required this.onLocationSelected,
    this.recentLocations = const [],
    super.key,
  });

  final LocationSearchController controller;
  final ValueChanged<LocationResult> onLocationSelected;
  final List<LocationResult> recentLocations;

  @override
  State<LocationSearchScreen> createState() => _LocationSearchScreenState();
}

class _LocationSearchScreenState extends State<LocationSearchScreen> {
  final _queryController = TextEditingController();
  late SupportedCountry _country;
  String? _queryError;

  @override
  void initState() {
    super.initState();
    _country = widget.controller.state.country;
  }

  @override
  void dispose() {
    _queryController.dispose();
    super.dispose();
  }

  void _search() {
    if (_queryController.text.trim().isEmpty) {
      setState(() => _queryError = 'Enter a city name');
      return;
    }

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
      appBar: AppBar(title: const Text('Choose a city')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
          child: Column(
            children: [
              DropdownButtonFormField<SupportedCountry>(
                value: _country,
                decoration: const InputDecoration(
                  labelText: 'Country',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.public_rounded),
                ),
                items: SupportedCountry.values
                    .map(
                      (country) => DropdownMenuItem(
                        value: country,
                        child: Text(country.displayName),
                      ),
                    )
                    .toList(growable: false),
                onChanged: (country) {
                  if (country != null) {
                    setState(() => _country = country);
                  }
                },
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _queryController,
                textInputAction: TextInputAction.search,
                onSubmitted: (_) => _search(),
                decoration: InputDecoration(
                  labelText: 'City',
                  hintText: 'e.g. Jakarta',
                  errorText: _queryError,
                  border: const OutlineInputBorder(),
                  prefixIcon: const Icon(Icons.search_rounded),
                ),
              ),
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: FilledButton.icon(
                  onPressed: _search,
                  icon: const Icon(Icons.search_rounded),
                  label: const Text('Search cities'),
                ),
              ),
              const SizedBox(height: 20),
              Expanded(
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
            ],
          ),
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
                Text(
                  'Recent cities',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 4),
                Text(
                  'Tap a city to load its latest forecast.',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ],
            ),
          );
        }

        final location = locations[index - 1];
        return Card(
          child: ListTile(
            key: ValueKey('recent-location-${location.locationId}'),
            leading: const CircleAvatar(child: Icon(Icons.history_rounded)),
            title: Text(location.city),
            subtitle: Text(_locationSubtitle(location)),
            trailing: const Icon(Icons.chevron_right_rounded),
            onTap: () => onSelect(location),
          ),
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
          padding: const EdgeInsets.only(left: 4, bottom: 8),
          child: Text(
            '${state.results.length} result${state.results.length == 1 ? '' : 's'}',
            style: Theme.of(context).textTheme.titleMedium,
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

              return Card(
                child: ListTile(
                  key: ValueKey('location-${location.openMeteoLocationId}'),
                  leading: const CircleAvatar(
                    child: Icon(Icons.location_on_rounded),
                  ),
                  title: Text(location.city),
                  subtitle: Text(_locationSubtitle(location)),
                  trailing: isSaving
                      ? const SizedBox.square(
                          dimension: 22,
                          child: CircularProgressIndicator(strokeWidth: 2.5),
                        )
                      : const Icon(Icons.chevron_right_rounded),
                  onTap: state.isSaving ? null : () => onSelect(location),
                ),
              );
            },
          ),
        ),
      ],
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
