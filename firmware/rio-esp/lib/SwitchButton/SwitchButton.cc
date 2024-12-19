#include "SwitchButton.h"

SwitchButton::SwitchButton(int switchPin, int led1Pin, int led2Pin, MuteButton& muteButton, bool& isSpeakerActive, bool& isHeadphonesActive, WiFiUDP& udp, const char* udpAddress, uint16_t udpPort): 
    _switchPin(switchPin),
    _led1Pin(led1Pin),
    _led2Pin(led2Pin),
    _isSpeakerActive(isSpeakerActive), //isSpeaker
    _isHeadphonesActive(isHeadphonesActive), //isSpeaker
    _speakerMuted(false),
    _headphonesMuted(false),
    _muteButton(muteButton),
    _udp(udp),
    _udpAddress(udpAddress),
    _udpPort(udpPort) {}

// Initialize the button and LEDs
void SwitchButton::init() {
    pinMode(_switchPin, INPUT_PULLUP); // Button input with pull-up resistor
    pinMode(_led1Pin, OUTPUT);        // LED1 as output
    pinMode(_led2Pin, OUTPUT);        // LED2 as output
    digitalWrite(_led1Pin, LOW);     // LED1 on initially
    digitalWrite(_led2Pin, HIGH);      // LED2 off initially

    sendState();
    Serial.printf("Switch Button initialized successfully (PIN %d)\n", _switchPin);
}

// Update the state of the button and send UDP message if toggled
void SwitchButton::update() {
    if (digitalRead(_switchPin) == LOW) { // Button press detected
        delay(50); // Debounce delay
        if (digitalRead(_switchPin) == LOW) { // Still pressed
            _isSpeakerActive = !_isSpeakerActive; // Toggle the state
            _isHeadphonesActive = !_isHeadphonesActive; 
            // Update LEDs based on the state
            digitalWrite(_led1Pin, _isSpeakerActive ? LOW : HIGH);
            digitalWrite(_led2Pin, _isSpeakerActive ? HIGH : LOW);

            // Send UDP message
            sendState();
            
            if (_isSpeakerActive) {
                _headphonesMuted = _muteButton.getState();
                _muteButton.setMute(_speakerMuted);
            } else {
                _speakerMuted = _muteButton.getState();
                _muteButton.setMute(_headphonesMuted);
            }

            while (digitalRead(_switchPin) == LOW) {
                delay(10); // Wait for the button to be released
            }
        }
    }
}

void SwitchButton::sendState() {
    String message = "SwitchOutput|" + String(_isSpeakerActive ? "0" : "1");
    _udp.beginPacket(_udpAddress, _udpPort);
    _udp.print(message);
    Serial.printf("sending %s\n", message.c_str());
    _udp.endPacket();
}
