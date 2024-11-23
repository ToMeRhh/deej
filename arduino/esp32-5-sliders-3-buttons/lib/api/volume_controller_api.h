#ifndef LIB_API_VOLUME_CONTROLLER_API_H
#define LIB_API_VOLUME_CONTROLLER_API_H

#include <WiFiUdp.h>

#include <string>

namespace lib {
namespace api {

class VolumeControllerApi {
 public:
  VolumeControllerApi(const char* server_ip, const int _server_port)
      : _server_ip(server_ip), _server_port(_server_port) {}
  void sendUdpData(const std::string& data);

 private:
  const char* _server_ip;
  const int _server_port;
  WiFiUDP _udp_client;
};

}  // namespace api
}  // namespace lib

#endif
