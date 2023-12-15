# SOCKS5 Server
A simple, fast SOCKS5 server written in Go

## HowTo
Compile the program and supply the public IP of the server:
```shell
./socks5-server [UDP Relay Address] <username> <password> <TCP Listening Address>
```
if you supply the username and password, the server will be only accessible using the username/password authentication method.

## Features
- Username/password authentication
- IPv4 and domain address type
- UDP association

The following features are NOT supported:
- GSSAPI authentication
- Bind command
- IPv6 address type

## Sponsor
It takes a lot of time to create and maintain a project. If you think it helped you, could you buy me a cup of coffee? 😉
You can use any of the following methods to donate:

| [![PayPal](/images/paypal.svg)](https://www.paypal.com/paypalme/tianchentang)<br/>Click [here](https://www.paypal.com/paypalme/tianchentang) to donate | ![Wechat Pay](/images/wechat.jpg)<br/>Wechat Pay | ![Alipay](/images/alipay.jpg) Alipay |
|--------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------|--------------------------------------|
