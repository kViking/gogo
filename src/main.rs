use clap::Parser;
use std::process::Command;
use std::collections::HashMap;
use regex;

mod gadget_store;
use crate::gadget_store::{Gadget, GadgetStore, GadgetVariable};

/// Simple program to greet a person
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    /// List all gadgets
    /// -l --list
    #[arg(short, long, default_value_t = false)]
    list: bool,

    /// Delete a gadget
    /// -d --delete <GADGET_NAME>
    #[arg(short, long)]
    delete: bool,

    /// Add a gadget
    /// -a --add <GADGET_NAME> 
    #[arg(short, long)]
    add: bool,

    /// -c --command <COMMAND> 
    /// Required when adding a gadget
    /// Name required for deleting a gadget
    #[arg(short, long)]
    name: Option<String>,
    /// -d --description <DESCRIPTION> 
    /// Required when adding a gadget
    #[arg(short, long)]
    command: Option<String>,
    /// -n --name <GADGET_NAME>
    /// Required when adding a gadget
    #[arg(long)]
    description: Option<String>,

    /// -e --edit [OTHER FLAGS]
    /// Edit a gadget. Requires other flags to specify what to edit.
    #[arg(short, long, default_value_t = false)]
    edit: bool,

    /// positional arguments for variable descriptions
    vars: Vec<String>,
}

fn main() {
    let args: Args = Args::parse();
    let mut store: GadgetStore = GadgetStore::new().unwrap_or(GadgetStore(HashMap::new()));

    if args.edit {
        
    }

    if args.add {
        let vars = args.vars.clone();
        // Example: check for required optionals before using them
        if args.command.as_ref().map(|s| s.is_empty()).unwrap_or(true)
            || args.description.as_ref().map(|s| s.is_empty()).unwrap_or(true)
            || args.name.as_ref().map(|s| s.is_empty()).unwrap_or(true)
        {
            eprintln!("Error: --command, --description, and --name are required and cannot be empty.");
            std::process::exit(1);
        }
        let mut new_gadget: Gadget = Gadget {
            name: args.name.clone().unwrap(),
            command: args.command.clone().unwrap(),
            description: args.description.clone().unwrap(),
            variables: vec![],
        };

        let regexp = regex::Regex::new(r"\{\{(\w+)\}\}").unwrap();
        let user_vars = regexp.captures_iter(args.command.as_deref().unwrap_or(""))
            .map(|cap| cap[1].to_string())
            .collect::<Vec<String>>();
        if vars.len() != user_vars.len() {
            eprintln!("Error: Number of provided descriptions does not match the number of variables in the command. If calling on the command line, ensure the descriptions are wrapped in double quotes.");
            std::process::exit(1);
        }
        for (var, desc) in user_vars.into_iter().zip(vars.into_iter()) {
            let gadget_var = GadgetVariable {
                name: var,
                description: desc,
                default: None,
            };
            new_gadget.variables.push(gadget_var);
        }
        
        store.save_gadget(new_gadget).unwrap_or_else(|err| {
            eprintln!("Error adding gadget: {}", err);
            std::process::exit(1);
        });
    }

    if args.delete {
        if args.name.as_ref().map_or(true, |s| s.is_empty()) {
            eprintln!("Error: --name is required when deleting a gadget.");
            std::process::exit(1);
        }
        store.delete_gadget(args.name.as_deref().unwrap_or("unknown")).unwrap_or_else(|err| {
            eprintln!("Error deleting gadget: {}", err);
            std::process::exit(1);
        });
    }

    if args.list {
        let gadgets: Vec<&Gadget> = store.list_gadgets();
        for gadget in gadgets {
            println!("Name: {}", gadget.name);
            println!("Description: {}", gadget.description);
            println!("Command: {}", gadget.command);
            if !gadget.variables.is_empty() {
                println!("Variables:");
                for var in &gadget.variables {
                    println!("  - Name: {}", var.name);
                    println!("    Description: {}", var.description);
                    if let Some(default) = &var.default {
                        println!("    Default: {}", default);
                    }
                }
            }
            println!();
        }
    }
    
    if !args.list && !args.add && !args.delete {
        let vars = args.vars.clone();
        if args.name.as_ref().map_or(true, |s| s.is_empty()) && vars.is_empty(){
            eprintln!("Error: Gadget name is required.");
            std::process::exit(1);
        }
        let gadget_opt: Option<&Gadget> = store.get_gadget(args.name.as_ref().unwrap());
        let gadget: &Gadget = match gadget_opt {
            Some(g) => g,
            None => {
                eprintln!("Error: Gadget '{}' not found.", args.name.as_deref().unwrap_or("unknown"));
                std::process::exit(1);
            }
        };

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
    }
}