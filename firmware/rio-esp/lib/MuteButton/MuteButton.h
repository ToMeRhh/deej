#ifndef MUTE_BUTTON_H
#define MUTE_BUTTON_H

#include <Arduino.h>

class MuteButton {
public:
    // Constructor
    MuteButton(int switchPin, int ledPin);

    // Initialize the button and LED
    void init();

    // Returns the current state: 1 if muted, 0 otherwise
    bool getState();

    // Manually set the mute state
    void setMute(bool isMuted);

private:
    int _switchPin;   // GPIO pin for the button
    int _ledPin;      // GPIO pin for the LED
    bool _isMuted;    // True if muted, false otherwise
};

#endif // MUTE_BUTTON_H