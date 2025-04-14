#include <Arduino.h>
#include <WiFi.h>
#include <audio_device_selector.h>
#include <backend_state.h>
#include <backend_state_api.h>
#include <mute_button.h>
#include <slider.h>
#include <string.h>
#include <util.h>
#include <volume_controller_api.h>

#include <string>
#include <vector>

#define SLIDER_0_PIN 33
#define SLIDER_1_PIN 35
#define SLIDER_2_PIN 32
#define SLIDER_3_PIN 36
#define SLIDER_4_PIN 34
#define MUTE_BUTTON_0_PIN 4
#define MUTE_BUTTON_0_LED_PIN 12
#define MUTE_BUTTON_1_PIN 14
#define MUTE_BUTTON_1_LED_PIN 21
#define AUDIO_DEVICE_SELECTOR_BUTTON_PIN 5
#define AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_LED_PIN 18
#define AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN 19

using lib::api::BackendState;
using lib::api::BackendStateApi;
using lib::api::VolumeControllerApi;
using lib::input_components::AudioDeviceSelector;
using lib::input_components::MuteButton;
using lib::input_components::Slider;

const char *ssid = "TnM";
const char *password = "tm5971088tm";
const char *server_address = "192.168.1.10";
const int udp_server_port = 16990;
const int tcp_server_port = 16991;

VolumeControllerApi *udp_api = nullptr;
BackendStateApi *tcp_api = nullptr;
std::vector<MuteButton *> *mute_buttons = nullptr;
std::vector<Slider *> *sliders = nullptr;
AudioDeviceSelector *audio_device_selector = nullptr;

void maybeApplyBackendState() {
  if (tcp_api == nullptr || audio_device_selector == nullptr) {
    Serial.println("TCP API or Audio Device Selector is not initialized.");
    return;
  }

  const auto current_output_device = tcp_api->getCurrentOutputDevice();
  if (current_output_device != -1) {
    Serial.print("Setting current output device to: ");
    Serial.println(current_output_device);
    audio_device_selector->setActiveDevice(current_output_device);
  } else {
    Serial.println("Failed to get current output device.");
  }
}

void setupWifi() {
  Serial.println("Starting");
  WiFi.setHostname("Deej");
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
  udp_api = new VolumeControllerApi(server_address, udp_server_port);
  tcp_api = new BackendStateApi(server_address, tcp_server_port);
  Serial.println("API Initialized!");

  Serial.println("Initializing components:");

  mute_buttons = new std::vector<MuteButton *>();
  // Controls two sessions (speakers and headphones).
  MuteButton *output_devices_mute_button =
      new MuteButton(0, MUTE_BUTTON_0_PIN, MUTE_BUTTON_0_LED_PIN, 2);
  MuteButton *mic_mute_button =
      new MuteButton(1, MUTE_BUTTON_1_PIN, MUTE_BUTTON_1_LED_PIN);
  mute_buttons->push_back(output_devices_mute_button);
  mute_buttons->push_back(mic_mute_button);
  Serial.println("Mute Buttons initialized!");

  sliders = new std::vector<Slider *>();
  sliders->push_back(new Slider(0, SLIDER_0_PIN,
                                std::make_optional<Slider::SessionMuteButton>(
                                    {output_devices_mute_button, 0})));
  sliders->push_back(new Slider(1, SLIDER_1_PIN,
                                std::make_optional<Slider::SessionMuteButton>(
                                    {output_devices_mute_button, 1})));
  sliders->push_back(new Slider(2, SLIDER_2_PIN));
  sliders->push_back(new Slider(3, SLIDER_3_PIN));
  sliders->push_back(new Slider(4, SLIDER_4_PIN));
  Serial.println("Sliders initialized!");

  audio_device_selector = new AudioDeviceSelector(
      AUDIO_DEVICE_SELECTOR_BUTTON_PIN,
      AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_LED_PIN,
      AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN, output_devices_mute_button,
      []() { esp_restart(); });

  maybeApplyBackendState();
  Serial.println("Audio device selector button initialized!");

  Serial.println("Initialization complete!");

  // Visually indicate that the system is ready.
  util::sequentialLEDOn(MUTE_BUTTON_0_LED_PIN, MUTE_BUTTON_1_LED_PIN,
                        AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_LED_PIN,
                        AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN, 300);
}

void loop() {
  std::string sliders_data = "Sliders";
  bool sliders_changed;
  for (int i = 0; i < sliders->size(); i++) {
    auto [changed, value] = sliders->at(i)->getValue();
    sliders_changed |= changed;
    sliders_data += "|";
    sliders_data += std::to_string(value);
  }
  if (sliders_changed) {
    udp_api->sendUdpData(sliders_data);
  }

  std::vector<bool> mute_buttons_state(mute_buttons->size());
  bool mute_buttons_changed = false;
  for (int i = 0; i < mute_buttons->size(); i++) {
    auto [changed, value] = mute_buttons->at(i)->getValue();
    mute_buttons_changed |= changed;
    mute_buttons_state[i] = value;
  }

  if (mute_buttons_changed) {
    const auto &updated_state =
        tcp_api->setMuteButtonsState(mute_buttons_state);

    if (updated_state.has_value() &&
        updated_state.value().size() == mute_buttons->size()) {
      for (int i = 0; i < mute_buttons->size(); i++) {
        mute_buttons->at(i)->setActiveSessionMuteState(
            updated_state.value()[i]);
      }
    } else {
      Serial.println("Failed to set mute buttons state.");
      util::blink2Leds(MUTE_BUTTON_0_LED_PIN, MUTE_BUTTON_1_LED_PIN, 150);
    }
  }

  if (auto [changed, value] = audio_device_selector->getValue(); changed) {
    const auto &updated_state = tcp_api->setOutputDeviceState(value);
    if (updated_state.has_value()) {
      // In case of failure - use the value from the server or fallback to 0.
      audio_device_selector->setActiveDevice(updated_state.value());
    } else {
      Serial.println("Failed to set output device state.");
      util::blinkLed(AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_LED_PIN, 150);
    }
  }

  delay(150);
}
