use regex::Regex;
use serde::{Serialize, Deserialize};
use std::fs::read_to_string;
use std::collections::HashMap;
use std::error::Error;
use std::path::Path;

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct GadgetVariable {
    pub name: String,
    pub description: String,
    pub default: Option<String>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Gadget {
    pub name: String,
    pub command: String,
    pub description: String,
    pub variables: Vec<GadgetVariable>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct GadgetStore(pub HashMap<String, Gadget>);

fn file_path() -> Result<String, Box<dyn Error>> {
    if std::env::consts::OS == "linux" {
        let home_dir: String = std::env::var("HOME")?;
        return Ok(format!("{}/.config/GoGoGadget/gadgets.json", home_dir));
    }
    let local_app_data: String = std::env::var("LOCALAPPDATA")?;
    Ok(format!("{}/GoGoGadget/gadgets.json", local_app_data))
}

impl Gadget {
    pub fn pretty_print(&self) {
        println!("Name: {}", self.name);
        println!("Description: {}", self.description);
        println!("Command: {}", self.command);
        if !self.variables.is_empty() {
            println!("Variables:");
            for var in &self.variables {
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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum GadgetAddError {
    NoVariablesButDescriptions,
    VariableCountMismatch { detected: Vec<String>, provided: Vec<String> },
    InvalidVariableName { name: String },
}

impl std::fmt::Display for GadgetAddError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            GadgetAddError::NoVariablesButDescriptions => write!(f, "No variables found in command, but variable descriptions were provided."),
            GadgetAddError::VariableCountMismatch { detected, provided } => write!(f, "Number of variable descriptions does not match number of variables in the command. Variables: {:?} Descriptions: {:?}", detected, provided),
            GadgetAddError::InvalidVariableName { name } => write!(f, "Variable name '{}' contains invalid characters. Only alphanumeric characters, underscores, and dashes are allowed.", name),
        }
    }
}

impl std::error::Error for GadgetAddError {}

/// Extracts variable names from a command string, e.g. "echo {{foo}}" -> ["foo"]
pub fn extract_variables_from_command(command: &str) -> Vec<String> {
    let re = Regex::new(r"\{\{(\w+)\}\}").unwrap();
    re.captures_iter(command)
        .map(|cap| cap[1].to_string())
        .collect()
}

impl GadgetStore {
    pub fn new() -> Result<Self, Box<dyn Error>> {
        let gadget_file_path: String = file_path()?;
        let parent_dir = Path::new(&gadget_file_path)
            .parent()
            .map(|p| p.exists())
            .unwrap_or(false);
        if !parent_dir {
            return Err("Gadget config directory does not exist".into());
        }
        let data: String = read_to_string(&gadget_file_path)?;
        let parsed: GadgetStore = serde_json::from_str(&data)?;
        Ok(parsed)
    }

    pub fn get_gadget(&self, name: &str) -> Option<&Gadget> {
        self.0.get(name)
    }

    pub fn list_gadgets(&self) -> Vec<&Gadget> {
        self.0.values().collect()
    }

    pub fn save_gadget(&mut self, gadget: &Gadget) -> Result<(), Box<dyn std::error::Error>> {
        let gadget_file_path = file_path()?;
        self.0.insert(gadget.name.clone(), gadget.clone());
        let serialized = serde_json::to_string_pretty(&self.0)?;
        std::fs::write(&gadget_file_path, serialized)?;
        Ok(())
    }

    pub fn delete_gadget(&mut self, name: &str) -> Result<(), Box<dyn std::error::Error>> {
        let gadget_file_path = file_path()?;
        self.0.remove(name);
        let serialized = serde_json::to_string_pretty(&self.0)?;
        std::fs::write(&gadget_file_path, serialized)?;
        Ok(())
    }

    pub fn add(&mut self, mut gadget: Gadget) -> Result<(), GadgetAddError> {
        let detected_vars = extract_variables_from_command(&gadget.command);
        if detected_vars.is_empty() && !gadget.variables.is_empty() {
            return Err(GadgetAddError::NoVariablesButDescriptions);
        }
        if !detected_vars.is_empty() && detected_vars.len() != gadget.variables.len() {
            return Err(GadgetAddError::VariableCountMismatch {
                detected: detected_vars.clone(),
                provided: gadget.variables.iter().map(|v| v.name.clone()).collect(),
            });
        }
        for var in &gadget.variables {
            if !Regex::new(r"^[\w-]+$").unwrap().is_match(&var.name) {
                return Err(GadgetAddError::InvalidVariableName { name: var.name.clone() });
            }
        }
        let mut desc_map = std::collections::HashMap::new();
        for v in &gadget.variables {
            desc_map.insert(v.name.clone(), (v.description.clone(), v.default.clone()));
        }
        let mut new_vars = Vec::new();
        for var_name in detected_vars {
            let (description, default) = desc_map.remove(&var_name)
                .unwrap_or((String::new(), None));
            new_vars.push(GadgetVariable {
                name: var_name,
                description,
                default,
            });
        }
        gadget.variables = new_vars;
        self.0.insert(gadget.name.clone(), gadget);
        Ok(())
    }
}
