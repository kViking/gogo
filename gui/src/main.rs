slint::include_modules!();

mod gadget_convert {
    use slint::{ModelRc, VecModel, Model};
    use super::{Gadget as SlintGadget, GadgetVariable as SlintGadgetVariable};
    use super::{CoreGadget, CoreGadgetVariable, Command};

    pub fn to_slint_gadget(core: &CoreGadget) -> SlintGadget {
        SlintGadget {
            name: core.name.clone().into(),
            command: core.command.raw.clone().into(),
            description: core.description.clone().into(),
            variables: ModelRc::new(VecModel::from(
                core.variables.iter().map(|v| SlintGadgetVariable {
                    name: v.name.clone().into(),
                    description: v.description.clone().into(),
                    default: v.default.clone().unwrap_or_default().into(),
                }).collect::<Vec<_>>()
            )),
        }
    }

    pub fn _from_slint_gadget(slint_gadget: &SlintGadget) -> CoreGadget {
        CoreGadget {
            name: slint_gadget.name.to_string(),
            command: Command::new(slint_gadget.command.to_string()),
            description: slint_gadget.description.to_string(),
            variables: slint_gadget.variables.iter().map(|v| CoreGadgetVariable {
                name: v.name.to_string(),
                description: v.description.to_string(),
                default: if v.default.is_empty() { None } else { Some(v.default.to_string()) },
            }).collect(),
        }
    }
}

use std::rc::Rc;
use std::cell::RefCell;
use gogo_core::{Gadget as CoreGadget, GadgetVariable as CoreGadgetVariable, GadgetStore, Command};
use gogo_core::GadgetAddError;
use slint::Model;
use std::collections::HashMap;

// Centralized application state
struct AppState {
    store: Rc<RefCell<GadgetStore>>,
    gadgets: Vec<CoreGadget>,
    selected_gadget: Option<CoreGadget>,
    original_gadget: Option<CoreGadget>, // Added field for cancel support
    edit_name: String,
    edit_description: String,
    edit_command: String,
    edit_variables: Vec<EditableVariable>,
    error_message: String,
}

impl AppState {
    fn new(store: Rc<RefCell<GadgetStore>>) -> Self {
        let gadgets = store.borrow().list_gadgets().into_iter().cloned().collect();
        Self {
            store,
            gadgets,
            selected_gadget: None,
            original_gadget: None, // Initialize as None
            edit_name: String::new(),
            edit_description: String::new(),
            edit_command: String::new(),
            edit_variables: Vec::new(),
            error_message: String::new(),
        }
    }

    fn reload_gadgets(&mut self) {
        self.gadgets = self.store.borrow().list_gadgets().into_iter().cloned().collect();
    }

    fn set_selected_gadget(&mut self, name: &str) {
        self.selected_gadget = self.store.borrow().get_gadget(name).cloned();
        self.original_gadget = self.selected_gadget.clone();
        if let Some(gadget) = &self.selected_gadget {
            self.original_gadget = Some(gadget.clone());
            self.edit_name = gadget.name.clone();
            self.edit_description = gadget.description.clone();
            self.edit_command = gadget.command.raw.clone();
            self.edit_variables = gadget.variables.iter().map(|v| EditableVariable {
                name: v.name.clone().into(),
                description: v.description.clone().into(),
                default: v.default.clone().unwrap_or_default().into(),
                name_error: false,
                desc_error: false,
                default_error: false,
            }).collect();
        } else {
            self.edit_name.clear();
            self.edit_description.clear();
            self.edit_command.clear();
            self.edit_variables.clear();
        }
    }

    fn sync_to_ui(&self, app: &AppWindow) {
        use slint::{ModelRc, VecModel};
        app.set_gadgets(ModelRc::new(VecModel::from(self.gadgets.iter().map(|g| gadget_convert::to_slint_gadget(g)).collect::<Vec<_>>())));
        app.set_edit_name(self.edit_name.clone().into());
        app.set_edit_description(self.edit_description.clone().into());
        app.set_edit_command(self.edit_command.clone().into());
        app.set_edit_variables(ModelRc::new(VecModel::from(self.edit_variables.clone())));
        app.set_error_message(self.error_message.clone().into());
    }

    /// Syncs variables from the old list to the new command, preserving metadata when possible.
    fn sync_variables(
        old_vars: &[EditableVariable],
        old_parts: &[gogo_core::CommandPart],
        new_parts: &[gogo_core::CommandPart],
    ) -> Vec<EditableVariable> {
        let old_text: Vec<_> = old_parts.iter().filter_map(|p| match p {
            gogo_core::CommandPart::Text(t) => Some(t),
            _ => None,
        }).collect();
        let new_text: Vec<_> = new_parts.iter().filter_map(|p| match p {
            gogo_core::CommandPart::Text(t) => Some(t),
            _ => None,
        }).collect();
        let new_var_names: Vec<_> = new_parts.iter().filter_map(|p| match p {
            gogo_core::CommandPart::Variable(v) => Some(v.clone()),
            _ => None,
        }).collect();
        let mut result = Vec::new();
        let old_vars_map: HashMap<&str, &EditableVariable> = old_vars.iter().map(|v| (v.name.as_str(), v)).collect();
        if old_text == new_text {
            // Likely a rename or reorder: match by index
            let old_var_names: Vec<_> = old_parts.iter().filter_map(|p| match p {
                gogo_core::CommandPart::Variable(v) => Some(v),
                _ => None,
            }).collect();
            for (i, name) in new_var_names.iter().enumerate() {
                if let Some(old_name) = old_var_names.get(i) {
                    if let Some(var) = old_vars_map.get(old_name.as_str()) {
                        let mut new_var = (*var).clone();
                        new_var.name = name.clone().into();
                        result.push(new_var);
                        continue;
                    }
                }
                // fallback: new variable
                result.push(EditableVariable {
                    name: name.clone().into(),
                    description: "".into(),
                    default: "".into(),
                    name_error: false,
                    desc_error: false,
                    default_error: false,
                });
            }
        } else {
            // Treat as new command: match by name if possible
            for name in new_var_names {
                if let Some(var) = old_vars_map.get(name.as_str()) {
                    result.push((*var).clone());
                } else {
                    result.push(EditableVariable {
                        name: name.clone().into(),
                        description: "".into(),
                        default: "".into(),
                        name_error: false,
                        desc_error: false,
                        default_error: false,
                    });
                }
            }
        }
        result
    }
}

// Helper: update variable name in place and update command string
fn update_variable_name_in_place(
    edit_variables: &mut [EditableVariable],
    index: usize,
    old_name: &str,
    new_name: &str,
    command: &mut Command,
) {
    if let Some(var) = edit_variables.get_mut(index) {
        var.name = new_name.into();
    }
    command.replace_variable_name(old_name, new_name);
}

fn main() {
    use slint::{ModelRc, VecModel};
    use gadget_convert::{to_slint_gadget, _from_slint_gadget};

    // Store gadgets from gadgets.json in the project root
    let store = GadgetStore::from_path("gadgets.json").expect("Failed to load gadgets.json");
    let store = Rc::new(RefCell::new(store));

    let app = AppWindow::new().unwrap();
    let app_state = Rc::new(RefCell::new(AppState::new(store.clone())));
    let app_weak = app.as_weak();
    let app_state_save = app_state.clone();

    // Initial sync to UI so gadgets are displayed
    app_state.borrow().sync_to_ui(&app);

    // Save callback
    app.on_save_gadget(move |
        name: slint::SharedString,
        description: slint::SharedString,
        command: slint::SharedString,
        variables: slint::ModelRc<EditableVariable>
    | {
        let mut state = app_state_save.borrow_mut();
        // Convert ModelRc<EditableVariable> to Vec<EditableVariable>
        let variables_vec: Vec<EditableVariable> = (0..variables.row_count()).map(|i| variables.row_data(i).unwrap()).collect();
        let backend_vars: Vec<CoreGadgetVariable> = variables_vec.iter().map(|v| CoreGadgetVariable {
            name: v.name.to_string(),
            description: v.description.to_string(),
            default: if v.default.is_empty() { None } else { Some(v.default.to_string()) },
        }).collect();
        let edited = CoreGadget {
            name: name.to_string(),
            description: description.to_string(),
            command: Command::new(command.to_string()),
            variables: backend_vars,
        };
        let edited_name = edited.name.clone();
        let add_result = state.store.borrow_mut().add(edited);
        // Clear all error flags by default
        let mut new_edit_vars = variables_vec.clone();
        let mut error_message = slint::SharedString::default();
        if let Err(err) = add_result {
            match err {
                GadgetAddError::InvalidVariableName { name: bad_name } => {
                    for (i, v) in variables_vec.iter().enumerate() {
                        if v.name == bad_name {
                            new_edit_vars[i].name_error = true;
                        }
                    }
                    error_message = format!("Invalid variable name: {}", bad_name).into();
                }
                GadgetAddError::VariableCountMismatch { .. } => {
                    for var in new_edit_vars.iter_mut() {
                        var.name_error = var.name.is_empty();
                    }
                    error_message = "Variable count mismatch".into();
                }
                GadgetAddError::NoVariablesButDescriptions => {
                    error_message = "No variables found in command, but variable descriptions were provided.".into();
                }
            }
            state.edit_variables = new_edit_vars.clone();
            state.error_message = error_message.to_string();
            if let Some(app) = app_weak.upgrade() {
                state.sync_to_ui(&app);
            }
            return;
        } else {
            // Success: clear errors
            for var in new_edit_vars.iter_mut() {
                var.name_error = false;
                var.desc_error = false;
                var.default_error = false;
            }
            state.error_message.clear();
            state.edit_variables = new_edit_vars.clone();
        }
        state.reload_gadgets();
        // Find index by name
        let idx = state.gadgets.iter().position(|g| g.name == edited_name).map(|i| i as i32).unwrap_or(-1);
        if let Some(app) = app_weak.upgrade() {
            state.sync_to_ui(&app);
            app.set_selected_index(-1); // Force rebind
            app.set_selected_index(idx);
        }
    });

    let app_weak = app.as_weak();
    let app_state_details = app_state.clone();
    // Handle request_gadget_details from Slint
    app.on_request_gadget_details(move |name| {
        let mut state = app_state_details.borrow_mut();
        state.set_selected_gadget(&name);
        if let Some(app) = app_weak.upgrade() {
            state.sync_to_ui(&app);
        }
    });

    // Cancel edit callback
    let app_state_cancel = app_state.clone();
    let app_weak_cancel = app.as_weak();
    app.on_cancel_edit(move || {
        let orig = app_state_cancel.borrow().original_gadget.clone();
        if let Some(orig) = orig {
            let mut state = app_state_cancel.borrow_mut();
            state.edit_name = orig.name.clone();
            state.edit_description = orig.description.clone();
            state.edit_command = orig.command.raw.clone();
            state.edit_variables = orig.variables.iter().map(|v| EditableVariable {
                name: v.name.clone().into(),
                description: v.description.clone().into(),
                default: v.default.clone().unwrap_or_default().into(),
                name_error: false,
                desc_error: false,
                default_error: false,
            }).collect();
            state.error_message.clear();
            if let Some(app) = app_weak_cancel.upgrade() {
                state.sync_to_ui(&app);
            }
        }
    });

    // Add command_edited and variable_name_edited handlers
    // (Assume Slint callback wiring is present)

    // Command edited: update AppState with new Command and sync variables
    let app_state_command = app_state.clone();
    let app_weak_command = app.as_weak();
    app.on_command_edited(move |new_command| {
        let mut state = app_state_command.borrow_mut();
        let old_cmd = Command::new(state.edit_command.clone());
        let new_cmd = Command::new(new_command.to_string());
        state.edit_command = new_cmd.raw.clone();
        // Extract variable names from old and new command
        let old_var_names: Vec<_> = old_cmd.parts.iter().filter_map(|p| match p {
            gogo_core::CommandPart::Variable(v) => Some(v),
            _ => None,
        }).collect();
        let new_var_names: Vec<_> = new_cmd.parts.iter().filter_map(|p| match p {
            gogo_core::CommandPart::Variable(v) => Some(v),
            _ => None,
        }).collect();
        // If variable count and text parts are the same, mutate in place
        let old_text: Vec<_> = old_cmd.parts.iter().filter_map(|p| match p {
            gogo_core::CommandPart::Text(t) => Some(t),
            _ => None,
        }).collect();
        let new_text: Vec<_> = new_cmd.parts.iter().filter_map(|p| match p {
            gogo_core::CommandPart::Text(t) => Some(t),
            _ => None,
        }).collect();
        if old_var_names.len() == new_var_names.len() && old_text == new_text {
            for (i, name) in new_var_names.iter().enumerate() {
                if let Some(var) = state.edit_variables.get_mut(i) {
                    var.name = name.clone().into();
                }
            }
        } else {
            // Otherwise, rebuild the array using sync_variables logic
            state.edit_variables = AppState::sync_variables(
                &state.edit_variables,
                &old_cmd.parts,
                &new_cmd.parts,
            );
        }
        if let Some(app) = app_weak_command.upgrade() {
            state.sync_to_ui(&app);
        }
    });

    // Variable name edited: update command string using replace_variable_name
    let app_state_varname = app_state.clone();
    let app_weak_varname = app.as_weak();
    app.on_variable_name_edited(move |index, old_name, new_name| {
        let mut state = app_state_varname.borrow_mut();
        let mut cmd = Command::new(state.edit_command.clone());
        update_variable_name_in_place(&mut state.edit_variables, index as usize, &old_name, &new_name, &mut cmd);
        state.edit_command = cmd.raw.clone();
        if let Some(app) = app_weak_varname.upgrade() {
            state.sync_to_ui(&app);
        }
    });

    app.run().unwrap();
}
