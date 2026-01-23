# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Regex/literal toggle - Switch between regex and fixed-string mode
- File type filters - Filter by extension using ripgrep's --type flag
- Match highlighting - Highlight the actual matched text within lines
- Syntax highlighting - Highlight code in preview pane using chroma
- History persistence - Save search patterns to XDG data directory
- Add comprehensive test suite

## [0.0.1] - TBD

### Added
- Initial alpha release of igrep - Interactive Grep TUI
- Real-time search results as you type with ripgrep integration
- Split-pane interface: results list (left) and file preview (right)
- Debounced input (200ms) for responsive typing experience
- Context preview showing 5 lines above and below each match
- Smart-case search support (case-insensitive unless uppercase is used)
- Case sensitivity toggle (Smart → Sensitive → Insensitive) via Ctrl+T
- Path scoping via dedicated input field
- Streaming results with batching for smooth UI updates
- Results capped at 10,000 for performance
- Editor integration - press Enter to open files at matched lines
- Built with Charm's Bubble Tea framework for elegant TUI

### Key Bindings
- **Tab**: Switch between pattern and path input
- **Up/Down** or **Ctrl+P/Ctrl+N**: Navigate results
- **PgUp/PgDn**: Jump 10 results at a time
- **Ctrl+T**: Toggle case sensitivity mode
- **Enter**: Open selected result in default editor
- **Ctrl+C**: Quit

### Requirements
- Go 1.23.4+ (for building from source)
- ripgrep (rg) must be installed and available in PATH

### Known Limitations
- No tests yet (alpha release)
- Limited error handling for edge cases
- No configuration file support
- No match highlighting within lines (submatch data parsed but not displayed)

[Unreleased]: https://github.com/William9923/igrep/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/William9923/igrep/releases/tag/v0.0.1