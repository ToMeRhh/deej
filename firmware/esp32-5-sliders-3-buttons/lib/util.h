#ifndef LIB_UTIL_H
#define LIB_UTIL_H

#include <Arduino.h>

namespace util {

inline void blinkLeds(int ledPin1, int ledPin2, int ledPin3,
                      int delayMs = 600) {
  // Store the current output states of the pins
  int previousState1 = digitalRead(ledPin1);
  int previousState2 = digitalRead(ledPin2);
  int previousState3 = digitalRead(ledPin3);

  digitalWrite(ledPin1, HIGH);
  digitalWrite(ledPin2, HIGH);
  digitalWrite(ledPin3, HIGH);
  delay(delayMs);

  digitalWrite(ledPin1, LOW);
  delay(delayMs);
  digitalWrite(ledPin2, LOW);
  delay(delayMs);
  digitalWrite(ledPin3, LOW);
  delay(delayMs);

  // Restore the previous output states
  digitalWrite(ledPin1, previousState1);
  digitalWrite(ledPin2, previousState2);
  digitalWrite(ledPin3, previousState3);
}
}  // namespace util

#endif