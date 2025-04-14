#include <WiFi.h>
#include <backend_state.h>
#include <backend_state_api.h>

#include <iostream>
#include <sstream>
#include <string>

namespace lib {
namespace api {

int BackendStateApi::getCurrentOutputDevice() {
  static const char* request_prefix = "GetCurrentOutputDevice";

  const auto& response = this->sendBackendRequest<int>(
      request_prefix, {});  // Use print to send a string
  if (!response) {
    Serial.println("Failed to get response from server");
    return -1;
  }

  try {
    return std::stoi(response->at(0).c_str());
  } catch (const std::exception& e) {
    Serial.println("Error converting string");
  }
  return -1;
}

std::optional<std::vector<bool>> BackendStateApi::setMuteButtonsState(
    const std::vector<bool>& new_state) {
  static const char* request_prefix = "MuteButtons";

  const auto& response = this->sendBackendRequest<bool>(
      request_prefix, new_state);  // Use print to send a string
  if (!response) {
    Serial.println("Failed to get response from server");
    return std::nullopt;
  }

  std::vector<bool> ret;
  for (const auto& str : *response) {
    ret.push_back(str == "true");
  }
  return ret;
}

std::optional<int> BackendStateApi::setOutputDeviceState(const int new_state) {
  static const char* request_prefix = "SwitchOutput";

  const auto& response = this->sendBackendRequest<int>(
      request_prefix, {new_state});  // Use print to send a string
  if (!response || response->size() != 1) {
    Serial.println("Failed to get response from server");
    return std::nullopt;
  }

  return std::stoi(response->at(0).c_str());
}
}  // namespace api
}  // namespace lib