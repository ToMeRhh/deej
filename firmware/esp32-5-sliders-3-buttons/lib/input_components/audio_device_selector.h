#ifndef LIB_INPUT_COMPONENTS_AUDIO_DEVICE_SELECTOR_H
#define LIB_INPUT_COMPONENTS_AUDIO_DEVICE_SELECTOR_H

#include <Arduino.h>
#include <mute_button.h>
#include <slider.h>

namespace lib {
namespace input_components {

class AudioDeviceSelector {
 public:
  AudioDeviceSelector(int button_gpio_pin, int dev_0_led_pin, int dev_1_led_pin,
                      MuteButton* multi_session_mute_button)
      : _button_gpio_pin(button_gpio_pin),
        _dev_0_led_pin(dev_0_led_pin),
        _dev_1_led_pin(dev_1_led_pin),
        _multi_session_mute_button(multi_session_mute_button),
        _selected_device(0) {
    _multi_session_mute_button->setActiveSession(_selected_device);
    // Configure button pin as input with pull-up resistor
    pinMode(_button_gpio_pin, INPUT_PULLUP);
    pinMode(_dev_0_led_pin, OUTPUT);
    pinMode(_dev_1_led_pin, OUTPUT);

    // Turn led off at start.
    // TODO 1: Set the selected device on startup, so the LED is correct.
    // TODO 2 (Optional): Fetch the correct state from the server and set it
    // here accordingly.
    digitalWrite(_dev_0_led_pin, HIGH);
    digitalWrite(_dev_1_led_pin, HIGH);

    setActiveDevice(_selected_device);
  }

  // Initializes this instance. Should be called at 'setup()'.
  void init();

  // Sets the given device index as the selected device, updates the LED
  // indication and updates the mute button pointer.
  void setActiveDevice(int selected_device);

  std::tuple<bool, int> getValue();
  std::tuple<std::string, std::string> getState();

 private:
  const int _button_gpio_pin;
  const int _dev_0_led_pin;
  const int _dev_1_led_pin;

  MuteButton* _multi_session_mute_button;
  int _selected_device;
  int _previous_device_mute_value;
};

}  // namespace input_components
}  // namespace lib

#endif