<!--
This is the main planning and requirements document for GoGoGadget. All requirements, design decisions, clarifications, and Q&A for the rewrite and new frontends are captured here. This file is the single source of truth for project planning and should be updated continuously as design decisions are made. It is important that this remain a living document—update it along the way with every new answer, decision, or requirement. Do not begin coding until this document is finalized and all open questions are resolved.
-->

# Project Goal
GoGoGadget aims to provide a safe, approachable, and user-friendly way for non-technical users to run, manage, and edit PowerShell command shortcuts ("gadgets") on Windows. The project abstracts away command-line complexity, enabling users to save, edit, and execute scripts with variable prompts and clear descriptions. The CLI is the foundation, while TUI and GUI frontends offer increasingly accessible and visual experiences. The ultimate goal is to empower users to automate and streamline their workflows without needing to understand scripting or command-line details, while maintaining safety, clarity, and extensibility. For more advanced users, the code should remain as modular as possible to allow for customization. 

# GoGoGadget Planning Notes

## Technical Summary & Implementation Notes
- **User Interface Strategy:**
  - CLI is the core, with a GUI as a secondary, user-friendly layer. The GUI should be able to invoke CLI commands and display their output, acting as a visual abstraction over the CLI.
  - TUI (if used) should mimic the CLI’s logic and flow, but present it in a column-based, navigable interface.

- **Feature Set:**
  - Core: Add, edit, delete gadgets; validate variables; execute scripts; analyze snippets for variable suggestions; colorized output/logging for user-facing messages.
  - See the 'Nice-to-Have Features' section below for additional ideas.

- **Config & Data:**
  - All UI layers (CLI, TUI, GUI) use a single config file (user_scripts.json), managed by the CLI tool. The GUI/TUI should never bypass the CLI for data changes.
  - For now, config is overwritten on each change.

- **Error Handling & Feedback:**
  - UI should show friendly error messages, but allow users to view full technical details if desired.
  - User feedback is directed to GitHub Issues for tracking and support.

- **Accessibility & UX:**
  - Use color, clear indicators, and minimal technical jargon. Prioritize clarity and confidence for non-technical users.
  - Help should be accessible at every turn. Assume users may not read documentation or remember instructions.
  - Onboarding: Show a disclaimer about script safety on first run or when adding/editing gadgets.

- **Extensibility & Modularity:**
  - All business logic (validation, parsing, script execution) must be UI-agnostic and testable independently of the UI.
  - UI layers (CLI, TUI, GUI) are separate apps, each acting as a frontend to the same business logic and config.

- **Testing:**
  - Unit tests for business logic. Manual testing for UI, primarily on Windows.

- **Platform & Release:**
  - Windows is the primary target. Linux support is a future consideration.
  - Manual release/update process to minimize package size and complexity.

- **Advanced Flows:**
  - Variable reordering is not needed, but recursive validation or auto-correction of variable names is important.

- **Logging & Debugging:**
  - No logging standard chosen yet; open to using a popular Go logging library.

- **Security, Privacy, and Permissions:**
  - No authentication or permissions planned. The app will execute any script provided by the user.

- **Update Notifications & Plugins:**
  - Update notifications are optional and should be lightweight if implemented. No plans for plugin or extension support.

- **Config File Location:**
  - No cloud sync planned, but allow users to move user_scripts.json to another location (e.g., for their own cloud backup solution).

---

## Nice-to-Have Features
- Tutorial mode (built-in help or onboarding for new users)
- GUI support for running gadgets and displaying output (not required for v1)
- Context-sensitive help popups or tooltips throughout the UI
- Rolling backup of config/user_scripts.json
- Batch operations (e.g., delete multiple gadgets at once in the GUI)
- Verbose/debug output toggles in CLI and GUI
- Update notifications (if lightweight)
- Any additional features as they arise
- **Themes (light/dark mode) (GUI only)**
- **Preview effect/output of a gadget before saving changes (GUI only)**
- **Favorite/pin frequently used gadgets (GUI only)**
- **Changelog or “what’s new” dialog after major updates**
- **Recently used/history list for gadgets (future release)**
- **Categories for gadgets (future release, one category per gadget, but consider method chaining for CLI: e.g., GoGoGadget counting.recursive.folders)**
- **Subcommands for gadgets (future release, generalized/hierarchical if possible)**
- **Backup/export of all gadgets/settings to JSON or CSV, with export options for columns/order/values**
- **Demo mode/safe mode for training/testing (include a sample gadget like Write-Host)**
- **Warnings for potentially dangerous PowerShell commands (future release)**
- **Sidebar customization in GUI (order of gadgets/categories)**
- **Breadcrumbs/path bar in GUI for navigation when using categories/subcategories**
- **Tooltip (not description) for gadgets in sidebar**
- **Restore defaults/reset all option in GUI/TUI (clobber user_scripts.json with default_scripts.json, including sample gadget)**
- **Trash/undo delete for gadgets (deleted gadgets go to a trash/recycle bin, with undo/restore in GUI and CLI)**
- **Track last run time and usage count for each gadget (visible in GUI, optionally in CLI/TUI)**

---

## Remaining/Clarifying Questions (Answered)
- Remember last opened gadget/view: **No**
- Keyboard shortcuts: **TUI will be fully keyboard-based, GUI will need hotkeys**
- Recently used/history list: **Nice to have, future release**
- Accessibility beyond color/clarity: **No, but modularity should allow for future support**
- Variable types: **All layers are on top of the CLI; types should match or be castable, but GUI can offer richer editors**
- Drag-and-drop for reordering: **Not needed, but categories and subcommands are nice-to-have for future**
- Export/backup: **Yes, with options for JSON/CSV and column selection/order/values**
- Dangerous command warnings: **Future feature, may require advanced logic**
- Demo/safe mode: **Yes, include a simple sample gadget**
- Categories: **One per gadget, but CLI could use method chaining for subcategories**
- Subcommands: **Should be generalized/hierarchical if possible**
- CSV export: **User can select columns/order/values, even if result is odd**
- Sidebar customization: **Yes, especially in GUI**
- Method chaining: **CLI only, not needed in GUI**
- Breadcrumbs/path bar: **Yes, in GUI**
- Sidebar tooltip: **Yes, tooltip only, not description**
- Restore defaults/reset all: **Yes, use default_scripts.json to restore**
- Duplicate/copy gadget as template: **Yes, in both GUI and CLI**
- Lock/protect gadgets: **Future feature**
- Search/filter gadgets: **Tab autocompletion in CLI if possible, search required in GUI**
- Multi-select for batch delete: **Yes, in GUI**

---

## New/Clarifying Questions
- Any other notes or requirements?
