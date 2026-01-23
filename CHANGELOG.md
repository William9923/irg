# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release preparation
- MIT License
- Enhanced documentation

## [1.0.0] - TBD

### Added
- Interactive grep TUI with real-time search results
- Split-pane interface: results list and file preview
- Ripgrep integration with JSON output parsing
- Debounced input (200ms) for responsive typing
- Context preview showing 5 lines above and below matches
- Smart-case search support
- Case sensitivity toggle (Smart → Sensitive → Insensitive)
- Path scoping via dedicated input field
- Streaming results with batching for smooth UI
- Results capped at 10,000 for performance
- Editor integration (open files at matched lines)
- Built with Charm's Bubble Tea framework

### Key Features
- **Tab**: Switch between pattern and path input
- **Up/Down** or **Ctrl+P/Ctrl+N**: Navigate results
- **PgUp/PgDn**: Jump 10 results at a time
- **Ctrl+T**: Toggle case sensitivity mode
- **Enter**: Open selected result in default editor
- **Ctrl+C**: Quit

### Technical Details
- Go 1.23.4+ requirement
- Ripgrep dependency for fast text search
- Clean separation of concerns: search backend and TUI frontend
- Comprehensive error handling and graceful degradation

[Unreleased]: https://github.com/william-nobara/igrep/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/william-nobara/igrep/releases/tag/v1.0.0