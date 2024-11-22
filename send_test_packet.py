import socket
sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM) 
# sock.sendto(bytes("MuteButtons|false|false", "utf-8"), ("127.0.0.1", 16990))
# sock.sendto(bytes("Sliders|200|400|600", "utf-8"), ("127.0.0.1", 16990))
sock.sendto(bytes("SwitchOutput|1", "utf-8"), ("127.0.0.1", 16990))