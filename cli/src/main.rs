mod commands;

use clap::{Parser, Subcommand};
use gogo_core::{Gadget, GadgetVariable, Command, GadgetStore};
use std::collections::HashMap;

#[derive(Parser, Debug)]
#[command(
    name = "gogo-cli",
    about = "GoGoGadget: a simple program to store and run command snippets.",
    version,
    long_about = None,
    after_help = "Examples:\n  gogo-cli add mygadget --command 'echo hi {{who}}' --description 'Say hi' 'Who to greet'\n  gogo-cli edit mygadget --new-name sayhello\n  gogo-cli list\n  gogo-cli delete mygadget\n  gogo-cli greet --who Jennifer",
    arg_required_else_help = true,
    disable_help_subcommand = true
)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Option<Commands>,
    /// If no subcommand is given, treat the first arg as a gadget name
    #[arg(hide = true, trailing_var_arg = true)]
    pub run_args: Vec<String>,
}

#[derive(Subcommand, Debug)]
pub enum Commands {
    /// Add a new gadget
    Add {
        name: String,
        #[arg(long)]
        command: String,
        #[arg(long)]
        description: Option<String>,
        /// Variable descriptions, in order
        #[arg()]
        vars: Vec<String>,
    },
    /// Edit an existing gadget
    Edit {
        name: String,
        #[arg(long)]
        new_name: Option<String>,
        #[arg(long)]
        command: Option<String>,
        #[arg(long)]
        description: Option<String>,
    },
    /// Delete a gadget
    Delete {
        name: String,
    },
    /// List all gadgets
    List,
    /// Show information about a specific gadget
    Info {
        name: String,
    },
}

fn main() {
    let cli = Cli::parse();
    let mut store: GadgetStore = GadgetStore::new().unwrap_or(GadgetStore(std::collections::HashMap::new()));

    match &cli.command {
        Some(Commands::Add { name, command, description, vars }) => {
            let var_names = extract_variable_names(command);
            if !var_names.is_empty() && vars.len() != var_names.len() {
                eprintln!("Error: Number of variable descriptions does not match number of variables in the command.\nVariables: {:?}\nDescriptions: {:?}", var_names, vars);
                std::process::exit(1);
            }
            commands::handle_add(name, command, description.as_deref().unwrap_or(""), vars, &mut store);
        }
        Some(Commands::Edit { name, new_name, command, description }) => {
            commands::handle_edit(
                name,
                None,
                new_name.as_deref(),
                description.as_deref(),
                command.as_deref(),
                &mut store,
            );
        }
        Some(Commands::Delete { name }) => {
            commands::handle_delete(name, &mut store);
        }
        Some(Commands::List) => {
            commands::handle_list(&store);
        }
        Some(Commands::Info { name }) => {
            commands::handle_info(&store, name);
        }
        None => {
            // Default: treat as run
            if !cli.run_args.is_empty() {
                let name = &cli.run_args[0];
                let mut var_map: HashMap<String, String> = HashMap::new();
                let mut iter = cli.run_args[1..].iter();
                while let Some(arg) = iter.next() {
                    if arg.starts_with("--") {
                        let key = arg.trim_start_matches("--").to_string();
                        if let Some(val) = iter.next() {
                            var_map.insert(key, val.to_string());
                        }
                    }
                }
                commands::handle_run(name, var_map, &store);
            } else {
                eprintln!("No command specified. Use --help for usage.");
                std::process::exit(1);
            }
        }
    }
}

fn extract_variable_names(command: &str) -> Vec<String> {
    let re = regex::Regex::new(r"\{\{(\w+)\}\}").unwrap();
    re.captures_iter(command)
        .map(|cap| cap[1].to_string())
        .collect()
}
