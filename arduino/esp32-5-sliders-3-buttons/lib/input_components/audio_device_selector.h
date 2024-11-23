#ifndef LIB_INPUT_COMPONENTS_AUDIO_DEVICE_SELECTOR_H
#define LIB_INPUT_COMPONENTS_AUDIO_DEVICE_SELECTOR_H

#include <slider.h>

namespace lib {
namespace input_components {

class AudioDeviceSelector {
 public:
  AudioDeviceSelector(const int gpioPinNumber)
      : _gpioPinNumber(gpioPinNumber), _selected_device(0) {}

  // Initializes this instance. Should be called at 'setup()'.
  void init();

  std::tuple<bool, int> getValue();
  std::tuple<std::string, std::string> getState();

 private:
  const int _gpioPinNumber;
  int _selected_device;
};

}  // namespace input_components
}  // namespace lib

#endif