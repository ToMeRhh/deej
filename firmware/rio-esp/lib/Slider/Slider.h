#ifndef SLIDER_H
#define SLIDER_H

#include <Arduino.h>
#include "MuteButton.h"

class Slider {
private:
    int _dataPin;               // Analog pin for the potentiometer
    int _lastPosition;          // Last known position of the slider
    int _threshold;             // Sensitivity threshold
    MuteButton* _muteButton;    // Pointer to the mute button (optional)
    bool* _isOutputActive;

    int invertAnalogValue(int value);

public:
    // Constructor
    Slider(int dataPin, MuteButton* muteButton = nullptr, bool* isOutputActive = nullptr);

    // Initialize the slider
    void init(int threshold);

    // Get the current position of the slider (0-1023)
    int getState();
};

#endif // SLIDER_H