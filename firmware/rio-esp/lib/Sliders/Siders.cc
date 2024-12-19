#include "Sliders.h"

Sliders::Sliders(Slider** sliders, size_t numSliders, int threshold, WiFiUDP& udp, const char* udpAddress, uint16_t udpPort)
    : _sliders(sliders), _numSliders(numSliders), _threshold(threshold), _udp(udp), _udpAddress(udpAddress), _udpPort(udpPort) {

    _lastStates = new int[_numSliders];
    for (size_t i = 0; i < _numSliders; i++) {
        _lastStates[i] = -1; 
    }
}

void Sliders::init() {
    for (size_t i = 0; i < _numSliders; i++) {
        if (_sliders[i]) {
            _sliders[i]->init(_threshold);
        }
    }
    Serial.printf("Initialized %d sliders.\n", _numSliders);
}

// Check for changes and update sliders
void Sliders::update() {
    bool stateChanged = false;
    String message = "Sliders";

    for (size_t i = 0; i < _numSliders; i++) {
        int currentState = _sliders[i]->getState();
        message += "|" + String(currentState);

        if (_lastStates[i] != currentState) {
            _lastStates[i] = currentState;
            stateChanged = true;
        }
    }

    if (stateChanged) {
        _udp.beginPacket(_udpAddress, _udpPort);
        _udp.print(message);
        _udp.endPacket();
    }
}