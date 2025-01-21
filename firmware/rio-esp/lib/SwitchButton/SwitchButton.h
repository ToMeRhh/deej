#ifndef SWITCH_BUTTON_H
#define SWITCH_BUTTON_H

#include <Arduino.h>
#include <WiFiUdp.h>
#include "MuteButton.h"

class SwitchButton {
private:
    int _switchPin;
    int _led1Pin;
    int _led2Pin;
    bool& _isSpeakerActive;
    bool& _isHeadphonesActive;
    bool _speakerMuted;
    bool _headphonesMuted;
    MuteButton& _muteButton;
    WiFiUDP& _udp;
    const char* _udpAddress;
    uint16_t _udpPort;  

public:
    // Constructor
    SwitchButton(int switchPin, int led1Pin, int led2Pin, MuteButton& muteButton, bool& _isSpeakerActive, bool& _isHeadphonesActive, WiFiUDP& udp, const char* udpAddress, uint16_t udpPort);

    // Initialize the button and LEDs
    void init();

    // Update the state of the button and send UDP message if toggled
    void update();

    void sendState();

    bool isHeadphonesActive();
};

#endif // SWITCH_BUTTON_H