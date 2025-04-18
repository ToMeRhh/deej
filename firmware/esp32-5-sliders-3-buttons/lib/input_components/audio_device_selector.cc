#include "audio_device_selector.h"

#include <Arduino.h>

namespace lib {
namespace input_components {

std::tuple<bool, int> AudioDeviceSelector::getValue() {
  if (digitalRead(_button_gpio_pin) == LOW) {
    // Debounce if needed.
    int debounce_count = 0;
    while (digitalRead(_button_gpio_pin) == LOW) {
      debounce_count++;
      if (debounce_count > 20) {  // 2 seconds == 2000 ms == 20 * delay(100)
        _on_longpress_override_callback();
        return std::tuple(false, _selected_device);
      }
      delay(100);
    }
    return std::tuple(true, _selected_device ^ 1);
  }
  return std::tuple(false, _selected_device);
}

void AudioDeviceSelector::setActiveDevice(int selected_device) {
  _selected_device = selected_device;
  if (_selected_device == 0) {
    digitalWrite(_dev_0_led_pin, HIGH);
    digitalWrite(_dev_1_led_pin, LOW);
  } else {
    digitalWrite(_dev_0_led_pin, LOW);
    digitalWrite(_dev_1_led_pin, HIGH);
  }
  _multi_session_mute_button->setActiveSession(selected_device);
}

}  // namespace input_components
}  // namespace lib