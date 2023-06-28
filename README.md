# bittorrent-with-tcp

## Overview
This project is a bittorent client written in Go (golang). The client uses **Bittorent Wire Protocol** and messages are sent over **tcp**.

## Running locally
1. Clone the repository:
```
git clone https://github.com/Ehab-24/bittorrent-with-tcp.git
```
2.
```
cd bittorent-with-tcp
```
3.
```
go run main.go [inPath] [outPath]
```
**inPath** is the path to _.torrent_ file and **outPaht** is where the downloaded content will be stored.

## Caution
The client does not handle **udp** connections and therefore any _.torrent_ files provided to the software must use **tcp** as well. Technologies such as DHT, PEX or magnetic links are not supported. If these limitations are not taken into consideration, you may see an error message like:
```
Get "udp://tracker.leechers-paradise.org:6969?compact=1&downloaded=0&info_hash=%26%A0RiN%F9%1C%2B%0CT%16z%7D%DAc%C2T%3C%DD%19&left=0&peer_id=%C5%A1%98KLV%BC%FD%C0%BC%08%5BX%D0%80%E5Q%0B%C1%A6&port=6881&uploaded=0": unsupported protocol scheme "udp"
```
