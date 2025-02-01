#ifndef LIB_API_BACKEND_STATE_H
#define LIB_API_BACKEND_STATE_H

#include <iostream>
#include <optional>
#include <vector>

namespace lib {
namespace api {

struct BackendState {
  std::optional<int> selected_device;
  std::optional<std::vector<bool>> mute_button_values;
  // Note: Sliders states are not supported.

  operator std::string() const {
    std::string out = "";
    if (selected_device.has_value()) {
      out += "SwitchOutput|";
      out += std::to_string(selected_device.value());
      out += "\n";
    }
    if (mute_button_values.has_value()) {
      out += "MuteButtons";
      for (bool value : mute_button_values.value()) {
        out += "|";
        out += std::to_string(value);
      }
      out += "\n";
    }
    return out;
  }
};

}  // namespace api
}  // namespace lib
#endif