#include "mute_button.h"

#include <Arduino.h>

#include <optional>

namespace lib {
namespace input_components {

std::tuple<bool, bool> MuteButton::getValue() {
  if (digitalRead(_button_gpio_pin) == LOW) {
    // Debounce if needed.
    while (digitalRead(_button_gpio_pin) == LOW) {
      delay(40);
    }
    setState(!this->_is_pressed[_active_session]);
    return std::tuple(true, this->_is_pressed[_active_session]);
  }
  return std::tuple(false, this->_is_pressed[_active_session]);
}

void MuteButton::setState(bool is_pressed) {
  this->_is_pressed[_active_session] = is_pressed;
  digitalWrite(_led_gpio_pin, is_pressed ? LOW : HIGH);
}

void MuteButton::setActiveSession(int active_session) {
  _active_session = active_session;
  MuteButton::setState(this->_is_pressed[_active_session]);
}

std::tuple<std::string, std::string> MuteButton::getState() {
  return {std::make_tuple(std::to_string(this->_button_index),
                          std::to_string(_is_pressed[_active_session]))};
}

}  // namespace input_components
}  // namespace lib