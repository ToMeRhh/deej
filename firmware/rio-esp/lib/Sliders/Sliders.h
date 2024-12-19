#ifndef SLIDERS_H
#define SLIDERS_H

#include <Arduino.h>
#include <WiFiUdp.h>
#include "Slider.h"

class Sliders {
public:
    // Constructor
    Sliders(Slider** sliders, size_t numSliders, int threshold, WiFiUDP& udp, const char* udpAddress, uint16_t udpPort);

    // Initialize all sliders
    void init();

    // Check for changes and update sliders
    void update();

    void sendState();

private:
    Slider** _sliders;     // Array of Slider pointers
    size_t _numSliders;    // Number of sliders
    int* _lastStates;
    int _threshold;      // Store the last states of the sliders to detect changes
    WiFiUDP& _udp;              // Reference to a WiFiUDP object for sending messages
    const char* _udpAddress;    // UDP destination address
    uint16_t _udpPort;          // UDP destination port
};

#endif // SLIDERS_H