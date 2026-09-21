# Forecast Flow Mobile

Flutter client for Forecast Flow. The app lets users search supported Asian
cities, view current, hourly, and daily forecasts, switch between recent
cities, and refresh the latest operational forecast.

## Prerequisites

- Flutter and the Android SDK
- An Android emulator or device
- The Forecast Flow API and its PostgreSQL/Kafka hot path

Run Flutter commands from `apps/mobile` unless a command says otherwise.

## Start the backend

From the repository root:

```powershell
docker compose up -d postgres kafka
docker compose --profile tools run --rm migrate
docker compose up -d --build api hotpath scheduler
```

Check the API before opening the app:

```powershell
Invoke-RestMethod http://localhost:8080/health/live
Invoke-RestMethod http://localhost:8080/health/ready
```

The app reads forecasts published to Kafka and consumed into PostgreSQL. The
scheduler checks every minute for newly added cities and forecasts older than
one hour. A new city's first forecast can take about a minute to pass through
that pipeline.

When a newly saved city does not have a forecast yet, the app displays a
`Preparing forecast` state and checks again automatically for up to 80 seconds.
The user can also check immediately or return to city selection.

To ingest a known PostgreSQL `location_id` from the repository root:

```powershell
$env:LOCATION_ID = '4'
docker compose --profile tools run --rm ingester
Remove-Item Env:LOCATION_ID
```

## Run on the Android emulator

Install dependencies once:

```powershell
flutter pub get
```

List available devices and run the app:

```powershell
flutter devices
flutter run -d emulator-5554 --dart-define=API_BASE_URL=http://10.0.2.2:8080
```

`10.0.2.2` is the Android emulator's route to `localhost` on the Windows host.
Do not use it on a physical phone; use the computer's reachable LAN address
instead.

While `flutter run` is active:

- Press `r` for hot reload.
- Press `R` for hot restart.
- Press `q` to stop the app.

After adding assets or changing `pubspec.yaml`, stop and rerun the app so the
asset bundle is rebuilt.

## Run on a physical Android phone

The phone and computer must be connected to the same local network. Find the
computer's active Wi-Fi IPv4 address:

```powershell
ipconfig
```

Then pass that address instead of the emulator-only `10.0.2.2` address. For
example, when the computer's Wi-Fi address is `192.168.1.5`:

```powershell
flutter devices
flutter run -d <phone-device-id> --dart-define=API_BASE_URL=http://192.168.1.5:8080
```

If Windows asks whether the API may accept private-network connections, allow
it. The Docker API publishes port `8080` on all host interfaces.

## Validate changes

The repository keeps Flutter temporary output on the `D:` drive:

```powershell
$env:TEMP = 'D:\Personal\forecast_flow\.tmp\flutter-test-temp'
$env:TMP = $env:TEMP
flutter analyze --no-pub
flutter test --no-pub
```

Build a debug APK with the emulator API address:

```powershell
flutter build apk --debug --dart-define=API_BASE_URL=http://10.0.2.2:8080
```

The APK is written to:

```text
build/app/outputs/flutter-apk/app-debug.apk
```

## Troubleshooting

### Location request timed out

Confirm both health endpoints work from Windows. If they do but the location
endpoint hangs, verify that a physical phone uses the computer's LAN IPv4
address rather than `10.0.2.2`. Then check `docker compose logs api`.

### Forecast unavailable

The city exists in PostgreSQL but does not have a consumed operational
forecast yet. Wait for the scheduler's next one-minute poll, then refresh the
app. Check `docker compose logs scheduler hotpath` if it remains unavailable.
The one-shot ingester command above remains available for manual retries.

### UI changes do not appear

Use hot restart (`R`). For newly added images, fonts, plugins, or other
`pubspec.yaml` changes, stop and rerun `flutter run`.
