# Packaging Numen For Distribution

Numen uses the [vosk](https://alphacephei.com/vosk) speech recognition
library.  You can find the vosk build template I wrote for Void Linux
[here](https://github.com/void-linux/void-packages/pull/39015).

There's also the simple templates for
[dotool](https://github.com/void-linux/void-packages/pull/40115) and
[numen](https://github.com/void-linux/void-packages/pull/39716) itself.

There's a service for runit in the `./service/` directory. It would be great
to have services that can optionally run as a specific user for other init
systems.

Thank you!
