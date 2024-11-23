#ifndef LIB_INPUT_COMPONENTS_MUTE_BUTTON_H
#define LIB_INPUT_COMPONENTS_MUTE_BUTTON_H

#include <optional>

namespace lib {
namespace input_components {

class MuteButton {
 public:
  MuteButton(int button_index, int gpioPinNumber)
      : _button_index(button_index),
        _gpioPinNumber(gpioPinNumber),
        _is_pressed(false) {}

  // Initializes this instance. Should be called at 'setup()'.
  void init();

  std::tuple<bool, bool> getValue();
  std::tuple<std::string, std::string> getState();

  const int _button_index;

 private:
  const int _gpioPinNumber;
  bool _is_pressed;
};

}  // namespace input_components
}  // namespace lib

#endif