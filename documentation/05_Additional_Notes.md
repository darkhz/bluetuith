- Ensure that the bluetooth service is up and running, and it is visible to DBus before launching the application. With systemd you can find out the status using the following command: `systemctl status bluetooth.service`.
- Only one transfer (either of send or receive) can happen on an adapter. Attempts to start another transfer while a transfer is already running (for example, trying to send files to a device when a transfer is already in progress) will be silently ignored.

