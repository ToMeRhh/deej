#include "slider.h"

#include <Arduino.h>

#include <optional>
#include <string>
#include <tuple>

namespace lib {
namespace input_components {
namespace {
bool valueIsChanged(int new_val, int old_val) {
  if (old_val == -1) {
    return true;
  }
  if (new_val == old_val) {
    return false;
  }
  if (new_val >= old_val + 25 || new_val <= old_val - 25 || new_val == 4095 ||
      new_val == 0) {
    return true;
  }
  return false;
}
}  // namespace

std::tuple<bool, int> Slider::getValue() {
  int percentValue = analogRead(_gpioPinNumber);
  if (valueIsChanged(percentValue, _previous_value)) {
    _previous_value = percentValue;
    if (_session_mute_button.has_value()) {
      _session_mute_button->button->setLedState(_session_mute_button->session,
                                                percentValue == 0);
    }
    return std::tuple(true, percentValue);
  }
  return std::tuple(false, percentValue);
}

}  // namespace input_components
}  // namespace lib