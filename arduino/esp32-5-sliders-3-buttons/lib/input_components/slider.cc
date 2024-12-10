#include "slider.h"

#include <Arduino.h>

#include <optional>
#include <string>
#include <tuple>

namespace lib {
namespace input_components {

void Slider::init() {
  this->getState();  // Force update new state.
}

std::tuple<bool, int> Slider::getValue() {
  int percentValue = analogRead(_gpioPinNumber);
  if (valueIsChanged(percentValue, _previous_value)) {
    _previous_value = percentValue;
    return std::tuple(true, percentValue);
  }
  return std::tuple(false, percentValue);
}

std::tuple<std::string, std::string> Slider::getState() {
  if (auto [changed, value] = getValue(); changed) {
    this->_previous_value = value;
  }
  return {std::make_tuple(std::to_string(this->_slider_index),
                          std::to_string(_previous_value))};
}

bool valueIsChanged(int new_val, int old_val) {
  if (new_val == old_val) {
    return false;
  }
  if (new_val >= old_val + 25 || new_val <= old_val - 25 || new_val == 4095 ||
      new_val == 0) {
    return true;
  }
  return false;
}

}  // namespace input_components
}  // namespace lib