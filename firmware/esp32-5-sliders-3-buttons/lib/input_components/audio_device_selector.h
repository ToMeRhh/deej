#ifndef LIB_INPUT_COMPONENTS_AUDIO_DEVICE_SELECTOR_H
#define LIB_INPUT_COMPONENTS_AUDIO_DEVICE_SELECTOR_H

#include <slider.h>

namespace lib {
namespace input_components {

class AudioDeviceSelector {
 public:
  AudioDeviceSelector(int button_gpio_pin, int dev_0_led_pin, int dev_1_led_pin)
      : _button_gpio_pin(button_gpio_pin),
        _dev_0_led_pin(dev_0_led_pin),
        _dev_1_led_pin(dev_1_led_pin),
        _selected_device(0) {}

  // Initializes this instance. Should be called at 'setup()'.
  void init();

  // Activates the led corresponds to the current active device.
  void setActiveLed();

  std::tuple<bool, int> getValue();
  std::tuple<std::string, std::string> getState();

 private:
  const int _button_gpio_pin;
  const int _dev_0_led_pin;
  const int _dev_1_led_pin;
  int _selected_device;
};

}  // namespace input_components
}  // namespace lib

#endif