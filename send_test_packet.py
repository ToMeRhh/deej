import socket
sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM) 
sock.sendto(bytes("MuteButtons|true|false|true", "utf-8"), ("127.0.0.1", 16990))