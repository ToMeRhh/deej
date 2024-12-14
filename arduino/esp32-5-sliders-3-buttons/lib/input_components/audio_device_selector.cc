#include "audio_device_selector.h"

#include <Arduino.h>

namespace lib {
namespace input_components {

void AudioDeviceSelector::init() {
  // Configure button pin as input with pull-up resistor
  pinMode(_button_gpio_pin, INPUT);
}

void AudioDeviceSelector::setActiveLed() {
  if (_selected_device == 0) {
    digitalWrite(_dev_0_gpio_pin, HIGH);
    digitalWrite(_dev_1_gpio_pin, LOW);
  } else if (_selected_device == 1) {
    digitalWrite(_dev_0_gpio_pin, LOW);
    digitalWrite(_dev_1_gpio_pin, HIGH);
  }
}

std::tuple<bool, int> AudioDeviceSelector::getValue() {
  if (digitalRead(_button_gpio_pin) == HIGH) {
    // Debounce if needed.
    while (digitalRead(_button_gpio_pin) == HIGH) {
      delay(100);
    }
    _selected_device ^= 1;
    setActiveLed();
    return std::tuple(true, _selected_device);
  }
  return std::tuple(false, _selected_device);
}

std::tuple<std::string, std::string> AudioDeviceSelector::getState() {
  return {std::make_tuple("SELECTOR", std::to_string(_selected_device))};
}

}  // namespace input_components
}  // namespace lib