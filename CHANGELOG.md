# Changelog

Notable changes to Numen will be documented in this file.
See ./phrases/CHANGELOG.md for changes to the default phrases.

## [0.5](https://git.sr.ht/~geb/numen/refs/0.5)

### Changed

- Reimplemented @transcribe and @cancel to slice the audio once the
results have been finalized, rather than hotplugging the audio between
commanding/transcribing and relying on unfinalized results.
- Now doesn't exit if the microphone is unplugged and continues when it is
plugged back in.

### Fixed

- Fixed paths in ./install-numen.sh that ignored the DESTDIR argument.

### Removed

- Removed the @instant tag now it's no longer needed with @transcribe and
@cancel.

## [0.4](https://git.sr.ht/~geb/numen/refs/0.4)

### Added

- A --version flag.

### Changed

- Now exits less confusingly if no model is installed.

## [0.3](https://git.sr.ht/~geb/numen/refs/0.3)

### Added

- This changelog har har har.

### Changed

- The @transcribe tag now sets an environment variable to your sentence
instead of typing it. This lets you apply a filter to the result.
- Simplified the runit service.

### Fixed

- You no longer need to leave a slight pause before transcribing a sentence.
- Sticky mode now releases keys properly when using --kernel.
- Replaced a redundant udev in ./install-dotool.sh with ./install-user-udev-rule.sh.
