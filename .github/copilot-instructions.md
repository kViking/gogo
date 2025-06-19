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

# Logical Model

## Data Model
- **Gadget**: name, command (Command struct), description, variables (Vec<GadgetVariable>)
- **Command**: raw string, parsed parts (Vec<CommandPart>), methods for variable extraction, renaming, and command building
- **GadgetVariable**: name, description, default
- **AppState**: holds the current list of gadgets, selected gadget, edit buffer, and UI state
- **Edit Buffer**: all edits are performed on a working copy of the selected gadget; original is restored on cancel

## State Management
- **Edit-on-copy**: When editing, operate on a deep copy of the gadget (edit buffer), leaving the original untouched until save/commit
- **Cancel**: Discard the edit buffer and restore the original
- **Live Preview**: UI is always bound to the edit buffer; all changes (including variable renames) are reflected live
- **Debouncing**: Not required for most operations, but can be added for expensive validation or backend sync

## Variable Sync Logic
- When the command string is edited:
  - Parse the new command into parts (Command::parse_parts)
  - Extract variable names from the new and old command
  - If the number of variables and the text parts are the same, mutate variable names in place (preserve focus and metadata)
  - If variables are added/removed or text parts change, rebuild the variable array, preserving metadata by name or index where possible
- When a variable name is edited via the variable field:
  - Mutate the variable name in place in the array
  - Update the command string using Command::replace_variable_name
  - Do not rebuild the array

## UI Update
- UI fields are always bound to the edit buffer (not the committed gadget)
- Only commit changes to the main store on save
- Only reset the edit buffer on cancel
- Use Slint's ListView or for loop for dynamic variable fields; mutate array members in place to preserve focus

## Slint Integration
- Use property bindings to bind UI elements to Rust backend properties
- Use callbacks for user actions (save, cancel, edit variable, edit command)
- Use ModelRc/VecModel for dynamic lists if needed
- Only update the model in place for renames/reorders; rebuild only for add/remove

## Error Handling & Validation
- All validation and error feedback is handled in Rust
- UI displays error messages and highlights fields as directed by backend state

## UX Principles
- Intuitive for low-skill users (Windows 10 conventions)
- Live feedback and preview
- Robust cancel/undo
- No focus loss or data loss on variable renames

# Questions to Consider
- State and state management: learn about slint's state management options and consider building custom systems if needed.
- Dynamic forms: investigate how to create dynamic forms in Slint based on variable input from the backend
- UX must be intuitive for low-skill users: consider user flows, error handling, and feedback mechanisms. Keep interface simple and clear, but based in familiar UI patterns. Users are Windows based, so consider Windows UI conventions (probably windows 10).
- Integration: how to best structure the integration between Slint and Rust backend. Consider using channels, callbacks, or other mechanisms to keep UI responsive while performing backend operations, and reflecting state changes in the UI.

# Implementation Notes
- Use [popups](https://releases.slint.dev/1.7.2/docs/slint/src/language/builtins/elements.html#popupwindow) for error messages and confirmations
- Use [ListView](https://releases.slint.dev/1.7.2/docs/slint/src/language/builtins/elements.html#listview) for the gadget/alias list
- Use [states](https://releases.slint.dev/1.7.2/docs/slint/src/language/builtins/states.html) for UI state management (loading, error, empty, etc.)
- Use [property bindings](https://releases.slint.dev/1.7.2/docs/slint/src/language/builtins/property_bindings.html) to keep UI and backend in sync
