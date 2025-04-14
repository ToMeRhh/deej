#ifndef LIB_API_BACKEND_STATE_API_H
#define LIB_API_BACKEND_STATE_API_H

#include <WiFiUdp.h>
#include <backend_state.h>

#include <iostream>
#include <optional>
#include <sstream>
#include <string>
#include <vector>

namespace lib {
namespace api {

class BackendStateApi {
 public:
  BackendStateApi(const char* server_ip, const int _server_port)
      : _server_ip(server_ip), _server_port(_server_port) {}

  int getCurrentOutputDevice();
  std::optional<std::vector<bool>> setMuteButtonsState(
      const std::vector<bool>& new_state);
  std::optional<int> setOutputDeviceState(const int new_state);

 private:
  template <typename T>
  std::optional<std::vector<std::string>> sendBackendRequest(
      const std::string& request_prefix, const std::vector<T>& params) {
    WiFiClient client;
    if (!client.connect(this->_server_ip, this->_server_port)) {
      Serial.println("Connection failed");
      return std::nullopt;
    }

    client.setTimeout(2);

    client.print(request_prefix.c_str());  // Use print to send a string
    for (const auto& param : params) {
      client.print("|");
      client.print(param);
    }
    client.print('\n');  // Send a newline to indicate the end of the request

    while (!client.available());  // wait for response

    String response = client.readStringUntil('\n');  // read entire response
    client.stop();                                   // Close the connection

    std::stringstream str(response.c_str());
    std::string segment;
    std::vector<std::string> ret;

    while (std::getline(str, segment, '|')) {
      ret.push_back(segment);
    }
    return std::make_optional(ret);
  }

  const char* _server_ip;
  const int _server_port;
};

}  // namespace api
}  // namespace lib

#endif
