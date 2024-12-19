#ifndef INITIALIZATIONS_H
#define INITIALIZATIONS_H

#include <WiFi.h>
#include <WiFiUdp.h>
#include "MuteButton.h"
#include "MuteButtons.h"
#include "SwitchButton.h"
#include "Slider.h"
#include "Sliders.h"

#define DEST_IP "192.168.0.178"
#define SSID "Mama 2.4"
#define PASSWORD "maytheforcebewithyou"
#define HOSTNAME "deejcontroller"
#define UDP_PORT 16990

// WiFi settings
extern const char* ssid;
extern const char* password;

// UDP settings
extern const char* udpAddress;
extern const uint16_t udpPort;

// MuteButton instances and MuteButtons manager
extern MuteButton button1;
extern MuteButton button2;
extern MuteButton* buttons[];
extern MuteButtons muteButtons;
extern SwitchButton switchButton;
extern Slider slider1;
extern Slider slider2;
extern Slider slider3;
extern Slider slider4;
extern Slider slider5;
extern Slider* slidersArr[];
extern Sliders sliders;

// WiFiUDP instance
extern WiFiUDP udp;

#endif // INITIALIZATIONS_H