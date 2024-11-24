This is a cloned fork of the Deej project (https://github.com/omriharel/deej); and specifically, iamjackg's fork that added UDP support (https://github.com/iamjackg/deej).

The main significant changes in the fork, from the original Deej project:
1. UDP support
2. Support for mute buttons (ones that actually mutes, not just lowers the volume to 0)
3. Support for output device toggle (i.e. toggle between sound devices)
4. A quick and dirty python UI for testing, while developing the Go code

# deej (with UDP support)

![deej](assets/deejudp-logo.png)

deej is an **open-source hardware volume mixer** for Windows and Linux PCs. It lets you use real-life sliders (like a DJ!) to **seamlessly control the volumes of different apps** (such as your music player, the game you're playing and your voice chat session) without having to stop what you're doing.

**This fork only supports receiving fader data via UDP.** This means that you can control deej with anything that can send UDP packets, including Wi-Fi enabled microcontrollers like the ESP8266 and ESP32, or scripts, or your home automation system, or anything you want!

For thorough documentation on the basics, please check out [the README of the original project](https://github.com/omriharel/deej).

**[Download the latest release](https://github.com/iamjackg/deej/releases/latest)**

## Configuration
In `config.yaml` edit the following properties:

### UDP Port

```yaml
# settings for the UDP connection
udp_port: 16990
```

### Sliders
an index based list of volume targets that will be controlled from the deej board.
See notes below on target names.

```yaml
slider_mapping:
  0: "Headphones (HyperX Cloud Flight Wireless Headset)"
  1: "Speakers (Realtek(R) Audio)"
  2: chrome.exe
  3: discord.exe
```

### Mute buttons
an index based list of targets that will be muted from the deej board.
See notes below on target names.

```yaml
mute_button_mapping:
  0: master
  1: mic
```

### Output device toggeling
an index based list of device IDs that will be available to choose from the deej board.
See notes below on target names.

```yaml
available_output_device:
  0: "{0.0.0.00000000}.{02870d54-9a1a-4478-a1a9-ced980df96fe}"
  1: "{0.0.0.00000000}.{717a777c-1b75-456b-bde9-2f4c6c4bf866}"
```

### Notes on target names
**TODO: Briefly document on how to get device names.**
* process names are case-insensitive
* you can use 'master' to indicate the master channel, or a list of process names to create a group
* you can use 'mic' to control your mic input level (uses the default recording device)
* you can use 'deej.unmapped' to control all apps that aren't bound to any slider (this ignores master, system, mic and device-targeting sessions)
* windows only - you can use 'deej.current' to control the currently active app (whether full-screen or not)
* windows only - you can use a device's full name, i.e. "Speakers (Realtek High Definition Audio)", to bind it. this works for both output and input devices
* windows only - you can use 'system' to control the "system sounds" volume
* important: slider indexes start at 0, regardless of which analog pins you're using!


## The modified Deej protocol
The deej protocol is very simple: each packet must consist of a command name, followed by a series of values separated by a `|`. No newline is necessary at the end of each packet.

The available commands are:
> Sliders
> 
> MuteButtons
> 
> SwitchOutput

### Examples
#### Sliders
If you have 5 sliders, and the second and third one are currently at the midpoint, your packets would look like this:

```text
Sliders|0|512|512|0|0
```
#### Mute buttons
If you have 2 mute buttons, 2 are muted and the last one is not, the packet would look like this:

```text
MuteButton|true|true|false
```
#### Toggle output device
To choose the device at index 1 in the config, send the packet:

```text
SwitchOutput|1
```  


### Building the controller

The basics are exactly the same as what is listed in the repo for the original project. The main difference is that the final string should be sent via UDP instead of through the serial port.

**TODO: add a sample firmware**

## License

deej is released under the [MIT license](./LICENSE).
