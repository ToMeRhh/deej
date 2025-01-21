#include "./volume_controller_api.h"

#include <WiFiUdp.h>

#include <string>

namespace lib {
namespace api {

void VolumeControllerApi::sendUdpData(const std::string& data) {
  // this->_udp_client.beginPacket(this->_server_ip, this->_server_port);
  Serial.println(data.c_str());
  // this->_udp_client.print(data.c_str());
  // this->_udp_client.endPacket();
}

}  // namespace api
}  // namespace lib
