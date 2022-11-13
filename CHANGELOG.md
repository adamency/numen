# Changelog

Notable changes to Numen will be documented in this file.
See ./phrases/CHANGELOG.md for changes to the default phrases.

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