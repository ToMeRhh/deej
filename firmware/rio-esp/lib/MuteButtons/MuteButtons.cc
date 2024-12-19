#include "MuteButtons.h"

// Constructor
MuteButtons::MuteButtons(MuteButton** buttons, size_t numButtons, WiFiUDP& udp, const char* udpAddress, uint16_t udpPort)
    : _buttons(buttons), _numButtons(numButtons), _udp(udp), _udpAddress(udpAddress), _udpPort(udpPort) {
    _lastStates = new bool[_numButtons];
    for (size_t i = 0; i < _numButtons; i++) {
        _lastStates[i] = false;
    }
}

// Initialize all buttons
void MuteButtons::init() {
    for (size_t i = 0; i < _numButtons; i++) {
        _buttons[i]->init();
    }
    update(true);
}

// Check for changes and send UDP messages if necessary
void MuteButtons::update(bool forceUpdate) {
    bool stateChanged = false;
    String message = "MuteButtons";

    // Check the state of each button
    for (size_t i = 0; i < _numButtons; i++) {
        bool currentState = _buttons[i]->getState();
        message += "|" + String(currentState ? "true" : "false");

        // Detect if any button state has changed
        if (currentState != _lastStates[i]) {
            _lastStates[i] = currentState;
            stateChanged = true;
        }
    }

    // Send UDP message if any state has changed
    if (stateChanged || forceUpdate) {
        _udp.beginPacket(_udpAddress, _udpPort);
        _udp.print(message);
        _udp.endPacket();
        Serial.printf("sending %s\n", message.c_str());
    }
}