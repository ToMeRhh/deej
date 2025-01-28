#ifndef LIB_INPUT_COMPONENTS_SLIDER_H
#define LIB_INPUT_COMPONENTS_SLIDER_H

#include <mute_button.h>

#include <optional>
#include <tuple>

namespace lib {
namespace input_components {

class Slider {
 public:
  struct SessionMuteButton {
    MuteButton* button;
    int session;
  };

  Slider(int slider_index, int gpioPinNumber)
      : Slider(slider_index, gpioPinNumber, std::nullopt) {}

  Slider(int slider_index, int gpioPinNumber,
         std::optional<SessionMuteButton> session_mute_button)
      : _slider_index(slider_index),
        _gpioPinNumber(gpioPinNumber),
        _session_mute_button(session_mute_button),
        _previous_value(0) {
    this->getState();  // Force update new state.
  }

  std::tuple<bool, int> getValue();
  std::tuple<std::string, std::string> getState();

  const int _slider_index;

 private:
  const int _gpioPinNumber;
  int _previous_value;
  const std::optional<SessionMuteButton> _session_mute_button;
};

}  // namespace input_components
}  // namespace lib

#endif