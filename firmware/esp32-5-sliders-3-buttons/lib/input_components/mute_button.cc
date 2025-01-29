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
    bool new_state = !this->_buttons_states[_active_session].is_pressed;
    setMuteState(_active_session, new_state);
    return std::tuple(true, new_state);
  }
  return std::tuple(false, this->_buttons_states[_active_session].is_pressed);
}

void MuteButton::setMuteState(int session, bool mute_state) {
  this->_buttons_states[session].is_pressed = mute_state;
  updateLedState();
}

void MuteButton::setLedState(int session, bool muted) {
  this->_buttons_states[session].led_state = muted;
  updateLedState();
}

void MuteButton::setActiveSession(int new_session) {
  _active_session = new_session;
  updateLedState();
}

void MuteButton::updateLedState() {
  const auto& current_state = this->_buttons_states[_active_session];
  digitalWrite(
      _led_gpio_pin,
      (current_state.is_pressed || current_state.led_state) ? LOW : HIGH);
}

std::tuple<std::string, std::string> MuteButton::getState() {
  return {std::make_tuple(
      std::to_string(this->_button_index),
      std::to_string(_buttons_states[_active_session].is_pressed))};
}

}  // namespace input_components
}  // namespace lib