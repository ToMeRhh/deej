#ifndef LIB_INPUT_COMPONENTS_SLIDER_H
#define LIB_INPUT_COMPONENTS_SLIDER_H

#include <optional>
#include <tuple>

namespace lib {
namespace input_components {

class Slider {
 public:
  Slider(int slider_index, int gpioPinNumber)
      : _slider_index(slider_index),
        _gpioPinNumber(gpioPinNumber),
        _previous_value(0) {}

  // Initializes this instance. Should be called at 'setup()'.
  void init();

  std::tuple<bool, int> getValue();
  std::tuple<std::string, std::string> getState();

  const int _slider_index;

 private:
  const int _gpioPinNumber;
  int _previous_value;
};

}  // namespace input_components
}  // namespace lib

#endif