#ifndef MUTE_BUTTONS_H
#define MUTE_BUTTONS_H

#include <Arduino.h>
#include <WiFiUdp.h>
#include "MuteButton.h"

class MuteButtons {
public:
    // Constructor
    MuteButtons(MuteButton** buttons, size_t numButtons, WiFiUDP& udp, const char* udpAddress, uint16_t udpPort);

    // Initialize all buttons
    void init();

    // Check for changes and send UDP messages if necessary
    void update(bool forceUpdate = false);

    void sendState();

private:
    MuteButton** _buttons;      // Array of MuteButton pointers
    size_t _numButtons;         // Number of buttons
    WiFiUDP& _udp;              // Reference to a WiFiUDP object for sending messages
    const char* _udpAddress;    // UDP destination address
    uint16_t _udpPort;          // UDP destination port
    bool* _lastStates;          // Store the last states of the buttons
};

#endif // MUTE_BUTTONS_H