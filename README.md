# Mercury Client

Mercury Client is a desktop utility for connecting to a Mercury HF modem over TCP and sending ARQ or broadcast messages.

## What it does

- Connects to a Mercury HF modem running in any of the supported radios using TCP.
- Starts an ARQ session with a remote station and sends ARQ data or files.
- Sends broadcast KISS frames over a separate broadcast channel.
- Shows connection status and incoming modem messages in a log.

## Installation

### Prerequisites

- Go 1.20+ installed
- Fyne-compatible desktop environment

### Build

```bash
git clone https://github.com/Rhizomatica/mercury-client.git
cd mercury-client
go build -o mercury-client .
```

### Run

```bash
./mercury-client
```

## Usage

1. Enter your callsign, the target callsign, IP address of your radio running Mercury HF modem, the ARQ port (default is 8300) and the broadcast port (default is 8100).
2. Click `Connect TCP` to connect to the modem.
3. Click `Connect ARQ` to start the ARQ session.
4. Type ARQ messages in the ARQ message field and click `Send ARQ Message`.
5. Type broadcast messages in the Broadcast field and click `Send Broadcast (KISS)`.

## About

Mercury Client is developed by Rhizomatica's HERMES team, namely:

- Rafael Diniz
- Pedro Messetti

Help to support the ongoing development of this project: <https://www.paypal.com/donate/?hosted_button_id=EKY4LRAH64Z9S>

## LICENSE

Mercury Client is a free software, licensed under the GNU General Public License, version 3 or (at your option) any later version (GPL-3.0-or-later). See `LICENSE` file for details.