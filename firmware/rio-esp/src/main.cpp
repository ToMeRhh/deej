// #include <Arduino.h>
#include "initializations.h"

bool isSpeakersActive = true;
bool isHeadphonesActive = false;

// Define MuteButton instances
MuteButton button1(14, 12);
MuteButton button2(4, 21);
MuteButton* buttons[] = { &button1, &button2};
MuteButtons muteButtons(buttons, 2, udp, DEST_IP, UDP_PORT);

SwitchButton switchButton(5, 18, 19, button1, isSpeakersActive, isHeadphonesActive, udp, DEST_IP, UDP_PORT);

// Define sliders
Slider slider1(34, &button1, &isHeadphonesActive); // start with speakers true
Slider slider2(35, &button1, &isSpeakersActive); // start with speakers false
Slider slider3(33);
Slider slider4(32);
Slider slider5(36);
Slider* slidersArr[] = { &slider1, &slider2, &slider3, &slider4, &slider5 };
Sliders sliders(slidersArr, 5, 15, udp, DEST_IP, UDP_PORT);

// WiFiUDP instance
WiFiUDP udp;

IPAddress destIp;

void connectToWifi() {
    destIp = IPAddress();
    destIp.fromString(DEST_IP);

    Serial.begin(9600);
    Serial.println();
    Serial.println("Configuring access point...");

    WiFi.mode(WIFI_STA);
    WiFi.disconnect(true);
    WiFi.config(INADDR_NONE, INADDR_NONE, INADDR_NONE);
    WiFi.setHostname(HOSTNAME);
    WiFi.begin(SSID, PASSWORD);

    while (WiFi.waitForConnectResult() != WL_CONNECTED) {
        Serial.println("Connection Failed! Rebooting...");
        delay(5000);
        ESP.restart();
    }

    Serial.println("Ready");
    Serial.print("IP address: ");
    Serial.println(WiFi.localIP());
}

void setup() {      
  Serial.begin(9600);

  connectToWifi();
  analogReadResolution(10);

  // Initialize all components
  switchButton.init();
  muteButtons.init();
  sliders.init();
}

void loop() {
  muteButtons.update();
  switchButton.update();
  sliders.update();
  delay(100); // Small delay for efficiency
}