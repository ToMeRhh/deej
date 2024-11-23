import socket
import tkinter as tk

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM) 
# # sock.sendto(bytes("MuteButtons|false|false", "utf-8"), ("127.0.0.1", 16990))
# # sock.sendto(bytes("Sliders|200|400|600", "utf-8"), ("127.0.0.1", 16990))


current_output_device = 0
sliders = []

def send_udp_data(data):
    sock.sendto(bytes(data, "utf-8"), ("127.0.0.1", 16990))
    

def create_slider(frame, label_text, command):
    label = tk.Label(frame, text=label_text)
    label.pack()

    slider = tk.Scale(frame, from_=0, to=1023, orient=tk.HORIZONTAL, command=command)
    slider.set(1023)
    slider.pack()
    return slider

def create_button(frame, text, command=None):
    button = tk.Button(frame, text=text, command=command)
    button.pack()
    return button

def switch_output():
    global current_output_device
    send_udp_data("SwitchOutput|{}".format(current_output_device))
    current_output_device = (current_output_device + 1) % 2

def send_slider_values(arg):
    print(arg)
    global sliders
    data = "Sliders|{}".format("|".join(str(x.get()) for x in sliders))
    print("Sending: ", data)
    send_udp_data(data)

def main():
    root = tk.Tk()
    root.title("Volume Mixer")

    frame = tk.Frame(root)
    frame.pack()

    global sliders
    for i in range(5):
        slider = create_slider(frame, f"Slider {i+1}", send_slider_values)
        sliders.append(slider)
    create_button(frame, "Send slider values", send_slider_values)

    mute_button1 = create_button(frame, "Mute 1")
    mute_button2 = create_button(frame, "Mute 2")
    output_button = create_button(frame, "Toggle Output", switch_output)

    root.mainloop()

if __name__ == "__main__":
    main()