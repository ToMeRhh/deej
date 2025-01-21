#include "mute_button.h"

#include <Arduino.h>

#include <optional>

namespace lib {
namespace input_components {

void MuteButton::init() {
  // Configure button pin as input with pull-up resistor
  pinMode(_gpioPinNumber, INPUT_PULLUP);
}

std::tuple<bool, bool> MuteButton::getValue() {
  if (digitalRead(_gpioPinNumber) == LOW) {
    // Debounce if needed.
    while (digitalRead(_gpioPinNumber) == LOW) {
      delay(40);
    }
    this->_is_pressed = !this->_is_pressed;
    return std::tuple(true, this->_is_pressed);
  }
  return std::tuple(false, this->_is_pressed);
}

std::tuple<std::string, std::string> MuteButton::getState() {
  return {std::make_tuple(std::to_string(this->_button_index),
                          std::to_string(_is_pressed))};
}

}  // namespace input_components
}  // namespace lib