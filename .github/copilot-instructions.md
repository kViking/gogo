<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

This project is a user-friendly command snippet manager for Windows (PowerShell) and Linux (Bash). It allows low-skill users to store, manage, and run command snippets (aliases) with variable injection, without needing to write scripts or use the command line directly.

The CLI app is now feature-complete and serves as the backend for the GUI app. All business logic, validation, and persistence is implemented in Rust in a modular way for easy integration.

Current focus: Building out the GUI using Slint, which will use the CLI/backend as its core logic. The GUI will provide:
- Visual gadget management (add, edit, delete, run)
- Dynamic variable forms (fields generated based on backend-detected variables in the command)
- Validation feedback and error highlighting
- Live command preview with injected variables
- All changes and runs go through the backend logic for consistency

All business logic, variable parsing, and validation should remain in Rust. The GUI should only reflect state and emit user actions, avoiding Slint scripting wherever possible. Within Rust, prefer using gogo-core library functions for consistency. Ensure logic is modular and testable. 

Slint is version 1.12.0, and the Rust toolchain is stable. There were breaking changes in Slint 1.10.0, so ensure compatibility with the 1.12.0 docs. The Slint docs are mirrored here in the project root at `slint_docs/`.

Questions to consider:
- State and state management: learn about slint's state management options and consider building custom systems if needed.
- Dynamic forms: investigate how to create dynamic forms in Slint based on variable input from the backend
- UX must be intuitive for low-skill users: consider user flows, error handling, and feedback mechanisms. Keep interface simple and clear, but based in familiar UI patterns. Users are Windows based, so consider Windows UI conventions (probably windows 10).
- Integration: how to best structure the integration between Slint and Rust backend. Consider using channels, callbacks, or other mechanisms to keep UI responsive while performing backend operations, and reflecting state changes in the UI.

Scratchpad for ideas:
use [popups](https://releases.slint.dev/1.7.2/docs/slint/src/language/builtins/elements.html#popupwindow) to create an alert class that can be reused for error messages and confirmations, maybe for edits/adds/deletes/runs
use [ListView](https://releases.slint.dev/1.7.2/docs/slint/src/language/builtins/elements.html#listview) to display the list of gadgets/aliases. Remember to alpha sort the list before displaying.
Consider using [states](https://releases.slint.dev/1.7.2/docs/slint/src/language/builtins/states.html) to manage different UI states, like loading, error, empty list, validation errors, etc. Is this React-like? Or is it more like a state machine? Research more.
use [property bindings](https://releases.slint.dev/1.7.2/docs/slint/src/language/builtins/property_bindings.html) to bind UI elements to Rust backend properties, ensuring that changes in the backend are reflected in the UI automatically.
