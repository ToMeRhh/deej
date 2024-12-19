#include "MuteButton.h"

// Constructor
MuteButton::MuteButton(int switchPin, int ledPin) 
    : _switchPin(switchPin), _ledPin(ledPin), _isMuted(false) {}

// Initialize the button and LED
void MuteButton::init() {
    pinMode(_switchPin, INPUT_PULLUP); // Button input with pull-up resistor
    pinMode(_ledPin, OUTPUT);         // LED as output
    digitalWrite(_ledPin, HIGH);       // LED off initially
    Serial.printf("Mute Button initialized successfully (PIN %d)\n", _switchPin);
}

// Get the current mute state
bool MuteButton::getState() {
    // Check if the button is pressed
    if (digitalRead(_switchPin) == LOW) { // Button press detected
        delay(50); // Debounce delay
        if (digitalRead(_switchPin) == LOW) { // Still pressed
            _isMuted = !_isMuted; // Toggle the mute state
            digitalWrite(_ledPin, _isMuted ? LOW : HIGH); // Update LED
        }
        while (digitalRead(_switchPin) == LOW) {
            delay(10); // Wait for the button to be released
        }
    }
    return _isMuted;
}

void MuteButton::setMute(bool isMuted) {
    _isMuted = isMuted; // Update the mute state
    digitalWrite(_ledPin, _isMuted ? LOW : HIGH); // Update the LED
}