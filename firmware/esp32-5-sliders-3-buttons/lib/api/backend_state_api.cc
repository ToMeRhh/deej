#include <WiFi.h>
#include <backend_state.h>
#include <backend_state_api.h>

#include <iostream>
#include <sstream>
#include <string>

namespace lib {
namespace api {

int BackendStateApi::getCurrentOutputDevice() {
  WiFiClient client;
  if (!client.connect(this->_server_ip, this->_server_port)) {
    Serial.println("Connection failed");
    return -1;
  }
  client.setTimeout(2);

  static const char* request_prefix = "GetCurrentOutputDevice";

  client.print(request_prefix);  // Use print to send a string
  client.print("\n");
  // while (!client.available());  // wait for response
  Serial.println("Client Available");
  String str = client.readString();                      // read entire response
  Serial.println("**********************************");  // Print the response
  Serial.println(str);                                   // Print the response
  client.stop();                                         // Close the connection
  int ret = std::stoi(str.c_str());
  return ret;
}

std::optional<std::vector<bool>> BackendStateApi::setMuteButtonsState(
    const std::vector<bool>& new_state) {
  // Create a TCP client
  WiFiClient client;

  if (!client.connect(this->_server_ip, this->_server_port)) {
    Serial.println("Connection failed");
    return std::nullopt;
  }

  static const char* request_prefix = "MuteButtons";

  client.print(request_prefix);  // Use print to send a string
  for (int value : new_state) {
    client.print("|");
    client.print(value);
  }
  client.print("\n");

  while (!client.available());  // wait for response

  String response = client.readStringUntil('\n');  // read entire response

  std::stringstream str(response.c_str());
  std::string segment;
  std::vector<bool> ret;

  while (std::getline(str, segment, '|')) {
    ret.push_back(segment == "true");
  }

  client.stop();  // Close the connection

  return ret;
}

std::optional<int> BackendStateApi::setOutputDeviceState(const int new_state) {
  // Create a TCP client
  WiFiClient client;

  if (!client.connect(this->_server_ip, this->_server_port)) {
    Serial.println("Connection failed");
    return std::nullopt;
  }

  static const char* request_prefix = "SwitchOutput|";

  client.print(request_prefix);  // Use print to send a string
  client.print(new_state);
  client.print("\n");

  while (!client.available());  // wait for response

  String str = client.readStringUntil('\n');  // read entire response

  client.stop();  // Close the connection
  return std::stoi(str.c_str());
}
}  // namespace api
}  // namespace lib