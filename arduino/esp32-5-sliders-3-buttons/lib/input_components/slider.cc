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
  int percentValue = map(analogRead(_gpioPinNumber), 0, 1023, 0, 100);
  if (percentValue == _previous_value) {
    return std::tuple(false, percentValue);
  }
  _previous_value = percentValue;
  return std::tuple(true, percentValue);
}

std::tuple<std::string, std::string> Slider::getState() {
  if (auto [changed, value] = getValue(); changed) {
    this->_previous_value = value;
  }
  return {std::make_tuple(std::to_string(this->_slider_index),
                          std::to_string(_previous_value))};
}

}  // namespace input_components
}  // namespace lib