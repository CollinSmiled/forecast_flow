class WeatherSceneResolver {
  const WeatherSceneResolver._();

  static const dayAsset = 'assets/scenes/day.png';
  static const nightAsset = 'assets/scenes/night.png';

  static String resolve({required bool isDay}) {
    return isDay ? dayAsset : nightAsset;
  }
}
