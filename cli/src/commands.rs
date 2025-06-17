// Handles CLI command logic for gogo

use gogo_core::{Gadget, GadgetStore, GadgetVariable};
use std::process::Command;
use regex;

fn extract_variable_names(command: &str) -> Vec<String> {
    let re = regex::Regex::new(r"\{\{(\w+)\}\}").unwrap();
    re.captures_iter(command)
        .map(|cap| cap[1].to_string())
        .collect()
}

pub fn handle_info(store: &GadgetStore, info_name: &str) {
    if let Some(gadget) = store.get_gadget(info_name) {
        gadget.pretty_print();
    } else {
        eprintln!("Error: Gadget '{}' not found.", info_name);
        std::process::exit(1);
    }
}

pub fn handle_edit(
    edit_name: &str,
    varname: Option<&str>,
    new_name: Option<&str>,
    new_desc: Option<&str>,
    new_command: Option<&str>,
    store: &mut GadgetStore,
) {
    if let Some(mut gadget) = store.get_gadget(edit_name).cloned() {
        let mut changed = false;
        if let Some(varname) = varname {
            if let Some(var) = gadget.variables.iter_mut().find(|v| v.name == varname) {
                if let Some(new_name) = new_name {
                    var.name = new_name.to_string();
                    changed = true;
                }
                if let Some(new_desc) = new_desc {
                    var.description = new_desc.to_string();
                    changed = true;
                }
            } else {
                eprintln!("Variable '{}' not found in gadget '{}'.", varname, edit_name);
                std::process::exit(1);
            }
        } else {
            let mut renamed = false;
            if let Some(new_name) = new_name {
                if new_name != &gadget.name {
                    gadget.name = new_name.to_string();
                    renamed = true;
                    changed = true;
                }
            }
            if let Some(new_command) = new_command {
                gadget.command = new_command.to_string();
                changed = true;
            }
            if let Some(new_desc) = new_desc {
                gadget.description = new_desc.to_string();
                changed = true;
            }
            if renamed {
                store.delete_gadget(edit_name).ok();
            }
        }
        if changed {
            store.save_gadget(&gadget).unwrap_or_else(|err| {
                eprintln!("Error editing gadget: {}", err);
                std::process::exit(1);
            });
        } else {
            eprintln!("No changes provided to edit gadget or variable.");
        }
    } else {
        eprintln!("Error: Gadget '{}' not found.", edit_name);
        std::process::exit(1);
    }
}

pub fn handle_add(
    add_name: &str,
    command: &str,
    description: &str,
    vars: &[String],
    store: &mut GadgetStore,
) {
    let var_names = extract_variable_names(command);
    if !var_names.is_empty() && vars.len() != var_names.len() {
        eprintln!("Error: Number of variable descriptions does not match number of variables in the command.\nVariables: {:?}\nDescriptions: {:?}", var_names, vars);
        std::process::exit(1);
    }
    let mut new_gadget: Gadget = Gadget {
        name: add_name.to_string(),
        command: command.to_string(),
        description: description.to_string(),
        variables: vec![],
    };
    for (var, desc) in var_names.into_iter().zip(vars.iter()) {
        let gadget_var = GadgetVariable {
            name: var,
            description: desc.clone(),
            default: None,
        };
        new_gadget.variables.push(gadget_var);
    }
    store.save_gadget(&new_gadget).unwrap_or_else(|err| {
        eprintln!("Error adding gadget: {}", err);
        std::process::exit(1);
    });
}

pub fn handle_delete(delete_name: &str, store: &mut GadgetStore) {
    store.delete_gadget(delete_name).unwrap_or_else(|err| {
        eprintln!("Error deleting gadget: {}", err);
        std::process::exit(1);
    });
}

pub fn handle_list(store: &GadgetStore) {
    for gadget in store.list_gadgets() {
        gadget.pretty_print();
    }
}

pub fn handle_run(name: &str, vars: &[String], store: &GadgetStore) {
    if let Some(gadget) = store.get_gadget(name) {
        let mut final_command = gadget.command.clone();
        for var in &gadget.variables {
            println!("Enter value for {} ({}): ", var.name, var.description);
            let mut input = String::new();
            std::io::stdin().read_line(&mut input).expect("Failed to read line");
            let input = input.trim();
            let value = if input.is_empty() {
                match &var.default {
                    Some(def) => def.clone(),
                    None => {
                        eprintln!("Error: No value provided for variable '{}' and no default is set.", var.name);
                        std::process::exit(1);
                    }
                }
            } else {
                input.to_string()
            };
            final_command = final_command.replace(&format!("{{{{{}}}}}", var.name), &value);
        }
        let status = if cfg!(target_os = "windows") {
            Command::new("cmd")
                .args(&["/C", &final_command])
                .status()
                .expect("failed to execute process")
        } else {
            Command::new("sh")
                .arg("-c")
                .arg(&final_command)
                .status()
                .expect("failed to execute process")
        };
        if !status.success() {
            eprintln!("Command executed with failing error code");
            std::process::exit(1);
        }
    } else {
        eprintln!("Error: Gadget '{}' not found.", name);
        std::process::exit(1);
    }
}

pub fn handle_run_with_vars(gadget_name: &str, var_map: std::collections::HashMap<String, String>, store: &GadgetStore) {
    if let Some(gadget) = store.get_gadget(gadget_name) {
        let mut final_command = gadget.command.clone();
        for var in &gadget.variables {
            let value = var_map.get(&var.name).cloned().or_else(|| {
                // Prompt if not provided
                println!("Enter value for {} ({}): ", var.name, var.description);
                let mut input = String::new();
                std::io::stdin().read_line(&mut input).expect("Failed to read line");
                let input = input.trim();
                if input.is_empty() {
                    var.default.clone()
                } else {
                    Some(input.to_string())
                }
            });
            let value = match value {
                Some(v) => v,
                None => {
                    eprintln!("Error: No value provided for variable '{}' and no default is set.", var.name);
                    std::process::exit(1);
                }
            };
            final_command = final_command.replace(&format!("{{{{{}}}}}", var.name), &value);
        }
        let status = if cfg!(target_os = "windows") {
            Command::new("cmd")
                .args(&["/C", &final_command])
                .status()
                .expect("failed to execute process")
        } else {
            Command::new("sh")
                .arg("-c")
                .arg(&final_command)
                .status()
                .expect("failed to execute process")
        };
        if !status.success() {
            eprintln!("Command executed with failing error code");
            std::process::exit(1);
        }
    } else {
        eprintln!("Error: Gadget '{}' not found.", gadget_name);
        std::process::exit(1);
    }
}
