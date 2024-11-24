#include <Arduino.h>
#include <AsyncTCP.h>
#include <ESPAsyncWebServer.h>
#include <WiFi.h>
#include <audio_device_selector.h>
#include <mute_button.h>
#include <slider.h>
#include <string.h>
#include <volume_controller_api.h>

#include <string>
#include <vector>

#define OUTPUT_DEVICE_0_MASTER_VOLUME_SLIDER_PIN 34
#define AUDIO_DEVICE_SELECTOR_PIN 17

using lib::api::VolumeControllerApi;
using lib::input_components::AudioDeviceSelector;
using lib::input_components::MuteButton;
using lib::input_components::Slider;

const char *ssid = "TnM";
const char *password = "tm5971088tm";
const char *server_address = "192.168.1.10";
const int server_port = 5000;

AsyncWebServer server(80);

VolumeControllerApi *api = nullptr;
std::vector<MuteButton *> *mute_buttons = nullptr;
std::vector<Slider *> *sliders = nullptr;
AudioDeviceSelector *audio_device_selector = nullptr;

void setupWifi() {
  Serial.println("Starting");
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    Serial.println(".");
    delay(500);
  }
  Serial.println("");
  Serial.println("Connected to WiFi!");
  Serial.println("IP address: ");
  Serial.println(WiFi.localIP());
}

bool get_sliders_data(std::string *out) {
  *out = "Sliders|";
  bool sliders_changed;
  for (int i = 0; i < sliders->size(); i++) {
    auto [changed, value] = sliders->at(i)->getValue();
    sliders_changed |= changed;
    *out += value;
    if (i < sliders->size() - 1) {
      *out += "|";
    }
  }
  *out += 100;
  *out += "|";
  *out += 50;
  *out += "|";
  *out += 10;
  *out += "|";
  return sliders_changed;
}

bool get_mute_buttons_data(std::string *out) {
  *out = "MuteButtons|";
  bool mute_buttons_changed;
  for (int i = 0; i < mute_buttons->size(); i++) {
    auto [changed, value] = mute_buttons->at(i)->getValue();
    mute_buttons_changed |= changed;
    *out += value;
    if (i < mute_buttons->size() - 1) {
      *out += "|";
    }
  }
  return mute_buttons_changed;
}

bool get_output_device_data(std::string *out) {
  auto [changed, value] = audio_device_selector->getValue();
  *out = "SwitchOutput|" + value;
  return changed;
}

void setup() {
  Serial.begin(115200);

  setupWifi();

  Serial.println("Initializing API:");
  api = new VolumeControllerApi(server_address, server_port);
  Serial.println("API Initialized!");

  Serial.println("Initializing components:");

  sliders = new std::vector<Slider *>();
  sliders->push_back(new Slider(0, OUTPUT_DEVICE_0_MASTER_VOLUME_SLIDER_PIN));
  for (auto *slider : *sliders) {
    slider->init();
  }
  Serial.println("Sliders initialized!");

  mute_buttons = new std::vector<MuteButton *>();
  for (auto *button : *mute_buttons) {
    button->init();
  }
  Serial.println("Mute Buttons initialized!");

  audio_device_selector = new AudioDeviceSelector(AUDIO_DEVICE_SELECTOR_PIN);
  Serial.println("Audio device selector button initialized!");

  Serial.println("Initializing server...");
  // send a file when /index is requested
  server.on("/index", HTTP_ANY, [](AsyncWebServerRequest *request) {
    std::string *sliders = new std::string();
    get_sliders_data(sliders);
    request->send(200, "text/plain", sliders->c_str());
  });
  Serial.println("Starting server...");
  server.begin();
}

void loop() {
  std::string mute_buttons_data;
  if (bool mute_buttons_changed = get_mute_buttons_data(&mute_buttons_data);
      mute_buttons_changed) {
    api->sendUdpData(mute_buttons_data);
  }

  std::string sliders_data;
  if (bool sliders_changed = get_sliders_data(&sliders_data); sliders_changed) {
    api->sendUdpData(sliders_data);
  }

  std::string output_device_data;
  if (bool output_device_changed = get_output_device_data(&output_device_data);
      output_device_changed) {
    api->sendUdpData(output_device_data);
  }

  delay(100);
}
