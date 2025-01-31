#ifndef LIB_API_BACKEND_STATE_H
#define LIB_API_BACKEND_STATE_H

namespace lib {
namespace api {

struct BackendState {
  int selected_device;
  int slider_values[5];
  int mute_button_values[2];
};

}  // namespace api
}  // namespace lib
#endif