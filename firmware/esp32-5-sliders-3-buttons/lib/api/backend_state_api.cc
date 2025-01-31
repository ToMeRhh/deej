#include <WiFi.h>
#include <backend_state.h>
#include <backend_state_api.h>

namespace lib {
namespace api {

BackendState BackendStateApi::getState() {
  // Create a TCP client
  WiFiClient client;

  if (!client.connect(this->_server_ip, this->_server_port)) {
    Serial.println("Connection failed");
    return BackendState();
  }

  const char* request = "get_state";
  client.println(request);  // Use print to send a string

  while (!client.available());  // wait for response

  String str = client.readStringUntil('\n');  // read entire response
  Serial.println(str);

  client.stop();  // Close the connection
  return BackendState();
}
}  // namespace api
}  // namespace lib