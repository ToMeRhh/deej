#include <Arduino.h>
#include <WiFi.h>
#include <audio_device_selector.h>
#include <mute_button.h>
#include <slider.h>
#include <string.h>
#include <volume_controller_api.h>

#include <string>
#include <vector>

#define SLIDER_0_PIN 34
#define SLIDER_1_PIN 35
#define SLIDER_2_PIN 32
#define SLIDER_3_PIN 33
// #define SLIDER_4_PIN 5
#define MUTE_BUTTON_0_PIN 21
#define MUTE_BUTTON_1_PIN 19
#define AUDIO_DEVICE_SELECTOR_BUTTON_PIN 22
#define AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_PIN 0
#define AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_PIN 0

using lib::api::VolumeControllerApi;
using lib::input_components::AudioDeviceSelector;
using lib::input_components::MuteButton;
using lib::input_components::Slider;

const char *ssid = "TnM";
const char *password = "tm5971088tm";
const char *server_address = "192.168.1.10";
const int server_port = 16990;

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

void setup() {
  Serial.begin(115200);

  setupWifi();

  Serial.println("Initializing API:");
  api = new VolumeControllerApi(server_address, server_port);
  Serial.println("API Initialized!");

  Serial.println("Initializing components:");

  sliders = new std::vector<Slider *>();
  sliders->push_back(new Slider(0, SLIDER_0_PIN));
  sliders->push_back(new Slider(1, SLIDER_1_PIN));
  sliders->push_back(new Slider(2, SLIDER_2_PIN));
  sliders->push_back(new Slider(3, SLIDER_3_PIN));
  // sliders->push_back(new Slider(4, SLIDER_4_PIN));
  for (auto *slider : *sliders) {
    slider->init();
  }
  Serial.println("Sliders initialized!");

  mute_buttons = new std::vector<MuteButton *>();
  mute_buttons->push_back(new MuteButton(0, MUTE_BUTTON_0_PIN));
  mute_buttons->push_back(new MuteButton(1, MUTE_BUTTON_1_PIN));
  for (auto *button : *mute_buttons) {
    button->init();
  }
  Serial.println("Mute Buttons initialized!");

  audio_device_selector = new AudioDeviceSelector(
      AUDIO_DEVICE_SELECTOR_BUTTON_PIN, AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_PIN,
      AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_PIN);
  Serial.println("Audio device selector button initialized!");
}

void loop() {
  std::string mute_buttons_data = "MuteButtons";
  bool mute_buttons_changed;
  for (int i = 0; i < mute_buttons->size(); i++) {
    auto [changed, value] = mute_buttons->at(i)->getValue();
    mute_buttons_changed |= changed;
    mute_buttons_data += "|";
    mute_buttons_data += std::to_string(value);
  }
  if (mute_buttons_changed) {
    api->sendUdpData(mute_buttons_data);
  }

  std::string sliders_data = "Sliders";
  bool sliders_changed;
  for (int i = 0; i < sliders->size(); i++) {
    auto [changed, value] = sliders->at(i)->getValue();
    sliders_changed |= changed;
    sliders_data += "|";
    sliders_data += std::to_string(value);
  }
  if (sliders_changed) {
    api->sendUdpData(sliders_data);
  }

  if (auto [changed, value] = audio_device_selector->getValue(); changed) {
    char data[20] = "SwitchOutput|";
    api->sendUdpData(strcat(data, std::to_string(value).c_str()));
  }

  delay(500);
}
