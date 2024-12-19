#include "Slider.h"

Slider::Slider(int dataPin, MuteButton* muteButton, bool* isOutputActive)
    : _dataPin(dataPin), _muteButton(muteButton), _isOutputActive(isOutputActive), _lastPosition(-1) {}

// Initialize the slider
void Slider::init(int threshold) {
    _threshold = threshold;
    pinMode(_dataPin, INPUT); // Set the analog pin for input
    _lastPosition = invertAnalogValue(analogRead(_dataPin)); // Initialize the position
    Serial.printf("Slider initialized on PIN %d with threshold %d\n", _dataPin, _threshold);
}

// Update the slider state and handle mute/unmute logic
int Slider::getState() {
    int currentPosition = invertAnalogValue(analogRead(_dataPin));

    if (abs(currentPosition - _lastPosition) >= _threshold) {
        _lastPosition = currentPosition;

        if (_muteButton != nullptr) {
            if (*_isOutputActive && currentPosition < _threshold && !_muteButton->getState()) {
                _muteButton->setMute(true);
            } else if (*_isOutputActive && currentPosition > 0 && _muteButton->getState()) {
                _muteButton->setMute(false);
            }
        }

        return currentPosition;
    }

    return _lastPosition;
}

int Slider::invertAnalogValue(int value) {
    const int MAX_ANALOG_VALUE = 1023; // Max value for 10-bit ADC
    return MAX_ANALOG_VALUE - value;
}

