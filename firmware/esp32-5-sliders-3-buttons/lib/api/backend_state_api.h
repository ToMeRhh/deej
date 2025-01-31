#ifndef LIB_API_BACKEND_STATE_API_H
#define LIB_API_BACKEND_STATE_API_H

#include <WiFiUdp.h>
#include <backend_state.h>

#include <string>

namespace lib {
namespace api {

class BackendStateApi {
 public:
  BackendStateApi(const char* server_ip, const int _server_port)
      : _server_ip(server_ip), _server_port(_server_port) {}

  BackendState getState();

 private:
  const char* _server_ip;
  const int _server_port;
};

}  // namespace api
}  // namespace lib

#endif
