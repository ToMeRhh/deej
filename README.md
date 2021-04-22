# deej (with UDP support)

![deej](assets/deejudp-logo.png)

deej is an **open-source hardware volume mixer** for Windows and Linux PCs. It lets you use real-life sliders (like a DJ!) to **seamlessly control the volumes of different apps** (such as your music player, the game you're playing and your voice chat session) without having to stop what you're doing.

**In addition to the usual serial communication, this fork supports receiving fader data via UDP.** If you have an ESP8266 or ESP32-based microcontroller, this will allow you to build a wireless controller that will talk to deej via Wi-Fi.

For thorough documentation on the basics, please check out [the README of the original project](https://github.com/omriharel/deej).

**[Download the latest release](https://github.com/iamjackg/deej/releases/latest)**

## UDP support

The configuration file for this fork has two extra options.

```yaml
# set this to "serial" to use deej with a board connected via a serial port
# set this to "udp" to listen for fader movements via a UDP network connection
controller_type: serial

# settings for the UDP connection
udp_port: 16990
```

If you set `controller_type` to `udp`, deej will start listening for UDP packets on the specified `udp_port` instead of opening the serial interface.

The deej protocol is very simple: each packet must consist of a series of numbers between 0 and 1023 for each slider, separated by a `|`. No newline is necessary at the end of each packet.

For example, if you have 5 faders, and the second and third one are currently at the midpoint, your packets would look like this:

```text
0|512|512|0|0
```

This means that you can control deej with anything that can send UDP packets, including Wi-Fi enabled microcontrollers like the ESP8266 and ESP32, or scripts, or your home automation system, or anything you want!

### Building the controller

The basics are exactly the same as what is listed in the repo for the original project. The main difference is that the final string should be sent via UDP instead of through the serial port.

See [my sample firmware repo](https://github.com/iamjackg/deej-esp32) for an example of how to accomplish this on an ESP32 board. 

## License

deej is released under the [MIT license](./LICENSE).
