#ifndef LIB_UTILS_UTIL_H
#define LIB_UTILS_UTIL_H

#include <Arduino.h>

namespace util {

inline void sequentialLEDOn(int ledPin1, int ledPin2, int ledPin3, int led4Pin,
                            int delayMs = 600) {
  // Store the current output states of the pins
  int previousState1 = digitalRead(ledPin1);
  int previousState2 = digitalRead(ledPin2);
  int previousState3 = digitalRead(ledPin3);
  int previousState4 = digitalRead(led4Pin);

  digitalWrite(ledPin1, HIGH);
  digitalWrite(ledPin2, HIGH);
  digitalWrite(ledPin3, HIGH);
  digitalWrite(led4Pin, HIGH);
  delay(delayMs);

  digitalWrite(ledPin1, LOW);
  delay(delayMs);
  digitalWrite(ledPin2, LOW);
  delay(delayMs);
  digitalWrite(ledPin3, LOW);
  delay(delayMs);
  digitalWrite(led4Pin, LOW);
  delay(delayMs);

  // Restore the previous output states
  digitalWrite(ledPin1, previousState1);
  digitalWrite(ledPin2, previousState2);
  digitalWrite(ledPin3, previousState3);
  digitalWrite(led4Pin, previousState4);
}

inline void blinkLed(int ledPin, int delayMs = 600) {
  // Store the current output state of the pin
  int previousState = digitalRead(ledPin);

  digitalWrite(ledPin, HIGH);
  delay(delayMs);
  digitalWrite(ledPin, LOW);
  delay(delayMs);
  digitalWrite(ledPin, HIGH);
  delay(delayMs);
  digitalWrite(ledPin, LOW);
  delay(delayMs);
  digitalWrite(ledPin, HIGH);
  delay(delayMs);
  digitalWrite(ledPin, LOW);
  delay(delayMs);

  // Restore the previous output state
  digitalWrite(ledPin, previousState);
}

inline void blink2Leds(int ledPin1, int ledPin2, int delayMs = 600) {
  // Store the current output state of the pins
  int previousState1 = digitalRead(ledPin1);
  int previousState2 = digitalRead(ledPin2);

  digitalWrite(ledPin1, HIGH);
  digitalWrite(ledPin2, HIGH);
  delay(delayMs);
  digitalWrite(ledPin1, LOW);
  digitalWrite(ledPin2, LOW);
  delay(delayMs);
  digitalWrite(ledPin1, HIGH);
  digitalWrite(ledPin2, HIGH);
  delay(delayMs);
  digitalWrite(ledPin1, LOW);
  digitalWrite(ledPin2, LOW);
  delay(delayMs);
  digitalWrite(ledPin1, HIGH);
  digitalWrite(ledPin2, HIGH);
  delay(delayMs);
  digitalWrite(ledPin1, LOW);
  digitalWrite(ledPin2, LOW);
  delay(delayMs);

  // Restore the previous output states
  digitalWrite(ledPin1, previousState1);
  digitalWrite(ledPin2, previousState2);
}
}  // namespace util

#endif