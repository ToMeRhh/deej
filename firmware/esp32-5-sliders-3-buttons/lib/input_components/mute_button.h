#ifndef LIB_INPUT_COMPONENTS_MUTE_BUTTON_H
#define LIB_INPUT_COMPONENTS_MUTE_BUTTON_H

#include <Arduino.h>

#include <map>
#include <optional>

namespace lib {
namespace input_components {

class MuteButton {
  struct ButtonState {
    bool is_pressed;
    bool led_state;
  };

 public:
  MuteButton(int button_index, int button_gpio_pin, int led_gpio_pin)
      : MuteButton(button_index, button_gpio_pin, led_gpio_pin, 1) {}

  MuteButton(int button_index, int button_gpio_pin, int led_gpio_pin,
             int controlled_sessions)
      : _button_index(button_index),
        _button_gpio_pin(button_gpio_pin),
        _led_gpio_pin(led_gpio_pin),
        _active_session(0) {
    // Configure button pin as input with pull-up resistor
    pinMode(_button_gpio_pin, INPUT_PULLUP);
    pinMode(_led_gpio_pin, OUTPUT);
    digitalWrite(_led_gpio_pin, HIGH);  // Turn led off at start.
  }

  std::tuple<bool, bool> getValue();
  std::tuple<std::string, std::string> getState();

  void setMuteState(int session, bool mute_state);
  void setActiveSession(int active_session);
  // Requests the led to be turned on when the button is not already muted.
  void setLedState(int session, bool muted);
  void updateLedState();

  const int _button_index;

 private:
  const int _button_gpio_pin;
  const int _led_gpio_pin;

  int _active_session;
  std::map<int, ButtonState> _buttons_states;
};

}  // namespace input_components
}  // namespace lib

#endif