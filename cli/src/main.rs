mod commands;

use clap::Parser;
use gogo_core::{GadgetStore};
use std::collections::HashMap;

/// GoGoGadget: a simple program to store and run command snippets.
///
/// Examples:
///   gogo-cli --add mygadget --command 'echo hi {{who}}' --description 'Say hi' 'Who to greet'
///   gogo-cli --edit mygadget --name sayhello
///   gogo-cli --list
///   gogo-cli --delete mygadget
///   gogo-cli sayhi --who Jennifer
#[derive(Parser, Debug)]
#[command(
    version,
    about = "GoGoGadget: a simple program to store and run command snippets.",
    long_about = None,
    after_help = "Examples:\n  gogo-cli --add mygadget --command 'echo hi {{who}}' --description 'Say hi' 'Who to greet'\n  gogo-cli --edit mygadget --name sayhello\n  gogo-cli --list\n  gogo-cli --delete mygadget\n  gogo-cli sayhi --who Jennifer"
)]
struct Args {
    /// List all gadgets
    #[arg(short, long, default_value_t = false)]
    list: bool,
    /// Add a gadget by name (requires --command and --description)
    #[arg(short, long, value_name = "GADGET_NAME", help = "Add a gadget by name. Requires --command and --description.")]
    add: Option<String>,
    /// Edit a gadget or a variable in a gadget.
    #[arg(long, value_name = "GADGETNAME")]
    edit: Option<String>,
    /// Edit a variable in a gadget by name.
    #[arg(long, value_name = "OLDNAME")]
    varname: Option<String>,
    /// Delete a gadget by name
    #[arg(short, long, value_name = "GADGET_NAME", help = "Delete a gadget by name. Example: --delete mygadget")]
    delete: Option<String>,
    /// View information about a specific gadget
    #[arg(short, long, value_name = "GADGET_NAME")]
    info: Option<String>,
    /// Set the new name for the gadget or variable (when used with --varname)
    #[arg(long)]
    name: Option<String>,
    /// Set the new description for the gadget or variable (when used with --varname)
    #[arg(long)]
    description: Option<String>,
    /// Set the new command for the gadget (not for variables)
    #[arg(long)]
    command: Option<String>,
    /// Variable descriptions for --add (one for each variable in the command, in order)
    #[arg(help = "Variable descriptions for --add (in order of appearance in the command)")]
    vars: Vec<String>,
    /// Positional arguments for running gadgets: first is gadget name, rest are variable flags
    #[arg(hide = true)]
    args: Vec<String>,
}

fn extract_variable_names(command: &str) -> Vec<String> {
    let re = regex::Regex::new(r"\{\{(\w+)\}\}").unwrap();
    re.captures_iter(command)
        .map(|cap| cap[1].to_string())
        .collect()
}

fn main() {
    let args: Args = Args::parse();
    let mut store: GadgetStore = GadgetStore::new().unwrap_or(GadgetStore(std::collections::HashMap::new()));

    if let Some(info_name) = &args.info {
        commands::handle_info(&store, info_name);
        return;
    }
    if let Some(edit_name) = &args.edit {
        commands::handle_edit(
            edit_name,
            args.varname.as_deref(),
            args.name.as_deref(),
            args.description.as_deref(),
            args.command.as_deref(),
            &mut store,
        );
        return;
    }
    if let Some(add_name) = &args.add {
        if let (Some(command), Some(description)) = (args.command.as_deref(), args.description.as_deref()) {
            let var_names = extract_variable_names(command);
            if !var_names.is_empty() && args.vars.len() != var_names.len() {
                eprintln!("Error: Number of variable descriptions does not match number of variables in the command.\nVariables: {:?}\nDescriptions: {:?}", var_names, args.vars);
                std::process::exit(1);
            }
            commands::handle_add(add_name, command, description, &args.vars, &mut store);
        } else {
            eprintln!("--command and --description are required for --add");
            std::process::exit(1);
        }
        return;
    }
    if let Some(delete_name) = &args.delete {
        commands::handle_delete(delete_name, &mut store);
        return;
    }
    if args.list {
        commands::handle_list(&store);
        return;
    }
    // If positional args are provided, treat as a gadget invocation
    if !args.args.is_empty() {
        let gadget_name = &args.args[0];
        let mut var_map: HashMap<String, String> = HashMap::new();
        let mut iter = args.args[1..].iter();
        while let Some(arg) = iter.next() {
            if arg.starts_with("--") {
                let key = arg.trim_start_matches("--").to_string();
                if let Some(val) = iter.next() {
                    var_map.insert(key, val.to_string());
                }
            }
        }
        commands::handle_run_with_vars(gadget_name, var_map, &store);
        return;
    }
    eprintln!("No command specified. Use --help for usage.");
    std::process::exit(1);
}
