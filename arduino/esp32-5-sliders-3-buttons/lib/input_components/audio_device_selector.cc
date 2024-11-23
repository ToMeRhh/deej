#include "audio_device_selector.h"

#include <Arduino.h>

namespace lib {
namespace input_components {

void AudioDeviceSelector::init() {
  // Configure button pin as input with pull-up resistor
  pinMode(_gpioPinNumber, INPUT_PULLUP);
}

std::tuple<bool, int> AudioDeviceSelector::getValue() {
  if (digitalRead(_gpioPinNumber) == LOW) {
    // Debounce if needed.
    while (digitalRead(_gpioPinNumber) == LOW) {
      delay(100);
    }
    this->_selected_device ^= 1;
    return std::tuple(true, this->_selected_device);
  }
  return std::tuple(false, this->_selected_device);
}

std::tuple<std::string, std::string> AudioDeviceSelector::getState() {
  return {std::make_tuple("SELECTOR", std::to_string(this->_selected_device))};
}

}  // namespace input_components
}  // namespace lib