slint::include_modules!();

mod gadget_convert {
    use slint::{ModelRc, VecModel, Model};
    use super::{Gadget as SlintGadget, GadgetVariable as SlintGadgetVariable};
    use super::{CoreGadget, CoreGadgetVariable};

    pub fn to_slint_gadget(core: &CoreGadget) -> SlintGadget {
        println!("[to_slint_gadget] name: {}, command: {}", core.name, core.command);
        SlintGadget {
            name: core.name.clone().into(),
            command: core.command.clone().into(),
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

    pub fn from_slint_gadget(slint_gadget: &SlintGadget) -> CoreGadget {
        CoreGadget {
            name: slint_gadget.name.to_string(),
            command: slint_gadget.command.to_string(),
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
use gogo_core::{Gadget as CoreGadget, GadgetVariable as CoreGadgetVariable, GadgetStore};
use gogo_core::GadgetAddError;
use slint::Model;

fn main() {
    use slint::{ModelRc, VecModel};
    use gadget_convert::{to_slint_gadget, from_slint_gadget};

    // Store dummy gadgets in a GadgetStore
    let mut store = GadgetStore([
        ("filecount".into(), CoreGadget {
            name: "filecount".into(),
            command: "ls {{dir}} | wc -l".into(),
            description: "Count files in a directory.".into(),
            variables: vec![CoreGadgetVariable {
                name: "dir".into(),
                description: "Directory to count files in".into(),
                default: Some(".".into()),
            }],
        }),
        ("greet".into(), CoreGadget {
            name: "greet".into(),
            command: "echo Hello, {{who}}!".into(),
            description: "Greet someone by name.".into(),
            variables: vec![CoreGadgetVariable {
                name: "who".into(),
                description: "Who to greet".into(),
                default: Some("World".into()),
            }],
        }),
        ("sayhi".into(), CoreGadget {
            name: "sayhi".into(),
            command: "echo Hi there!".into(),
            description: "Say hi with no variables.".into(),
            variables: vec![],
        }),
        ("date".into(), CoreGadget {
            name: "date".into(),
            command: "date".into(),
            description: "Show the current date and time.".into(),
            variables: vec![],
        }),
    ].into_iter().collect());
    let store = Rc::new(RefCell::new(store));

    let app = AppWindow::new().unwrap();

    // Helper to reload gadgets from dummy GadgetStore, returns Vec of SlintGadget
    let reload_gadgets = {
        let store = store.clone();
        let app_weak = app.as_weak();
        move || {
            let gadgets: Vec<_> = store.borrow().list_gadgets().iter().map(|g| to_slint_gadget(g)).collect();
            println!("[reload_gadgets] All gadget commands:");
            for g in &gadgets {
                println!("  {}: {}", g.name, g.command);
            }
            if let Some(app) = app_weak.upgrade() {
                app.set_gadgets(ModelRc::new(VecModel::from(gadgets.clone())));
            }
            gadgets
        }
    };

    reload_gadgets();

    // Save callback: update dummy GadgetStore and reload, then select by name
    let store_save = store.clone();
    let reload_gadgets_save = reload_gadgets.clone();
    let app_weak = app.as_weak();
    app.on_save_gadget(move |
        name: slint::SharedString,
        description: slint::SharedString,
        command: slint::SharedString,
        variables: slint::ModelRc<EditableVariable>
    | {
        use slint::{ModelRc, VecModel, SharedString};
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
            command: command.to_string(),
            variables: backend_vars,
        };
        let edited_name = edited.name.clone();
        let add_result = store_save.borrow_mut().add(edited);
        if let Some(app) = app_weak.upgrade() {
            // Clear all error flags by default
            let mut new_edit_vars = variables_vec.clone();
            let mut error_message = SharedString::default();
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
                app.set_edit_variables(ModelRc::new(VecModel::from(new_edit_vars.clone())));
                app.set_error_message(error_message);
                return;
            } else {
                // Success: clear errors
                for var in new_edit_vars.iter_mut() {
                    var.name_error = false;
                    var.desc_error = false;
                    var.default_error = false;
                }
                app.set_error_message(SharedString::default());
                app.set_edit_variables(ModelRc::new(VecModel::from(new_edit_vars.clone())));
            }
            let gadgets = reload_gadgets_save();
            // Find index by name
            let idx = gadgets.iter().position(|g| g.name == edited_name).map(|i| i as i32).unwrap_or(-1);
            app.set_selected_index(-1); // Force rebind
            app.set_selected_index(idx);
        }
    });

    let app_weak = app.as_weak();
    // Handle request_gadget_details from Slint
    app.on_request_gadget_details(move |name| {
        use slint::{ModelRc, VecModel};
        let gadget_opt = {
            let store_ref = store.borrow();
            store_ref.get_gadget(&name).cloned()
        };
        let app = app_weak.upgrade().unwrap();
        if let Some(gadget) = gadget_opt {
            let edit_vars = gadget.variables.iter().map(|v| EditableVariable {
                name: v.name.clone().into(),
                description: v.description.clone().into(),
                default: v.default.clone().unwrap_or_default().into(),
                name_error: false,
                desc_error: false,
                default_error: false,
            }).collect::<Vec<_>>();
            app.set_edit_name(gadget.name.clone().into());
            app.set_edit_description(gadget.description.clone().into());
            app.set_edit_command(gadget.command.clone().into());
            app.set_edit_variables(ModelRc::new(VecModel::from(edit_vars)));
        } else {
            app.set_edit_name("".into());
            app.set_edit_description("".into());
            app.set_edit_command("".into());
            app.set_edit_variables(ModelRc::new(VecModel::from(Vec::new())));
        }
    });

    app.run().unwrap();
}
