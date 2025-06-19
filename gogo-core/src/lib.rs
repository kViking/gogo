use regex::Regex;
use serde::{Serialize, Deserialize};
use std::fs::read_to_string;
use std::collections::HashMap;
use std::error::Error;
use std::path::Path;

#[derive(Serialize, Deserialize, Debug, Clone, PartialEq, Eq)]
pub enum CommandPart {
    Text(String),
    Variable(String),
}

#[derive(Serialize, Deserialize, Debug, Clone, PartialEq, Eq)]
pub struct Command {
    pub raw: String,
    #[serde(skip)]
    pub parts: Vec<CommandPart>,
}

impl Command {
    pub fn new(raw: String) -> Self {
        let parts = Self::parse_parts(&raw);
        Self { raw, parts }
    }

    pub fn parse_parts(raw: &str) -> Vec<CommandPart> {
        let re = Regex::new(r"\{\{(\w+)\}\}").unwrap();
        let mut parts = Vec::new();
        let mut last = 0;
        for cap in re.captures_iter(raw) {
            let m = cap.get(0).unwrap();
            if m.start() > last {
                parts.push(CommandPart::Text(raw[last..m.start()].to_string()));
            }
            parts.push(CommandPart::Variable(cap[1].to_string()));
            last = m.end();
        }
        if last < raw.len() {
            parts.push(CommandPart::Text(raw[last..].to_string()));
        }
        parts
    }

    /// Extracts variable names from the command string, e.g. "echo {{foo}}" -> ["foo"]
    pub fn extract_variables(&self) -> Vec<String> {
        self.parts.iter().filter_map(|p| {
            if let CommandPart::Variable(v) = p { Some(v.clone()) } else { None }
        }).collect()
    }

    /// Replace a variable name in the command string (all occurrences)
    pub fn replace_variable_name(&mut self, old: &str, new: &str) {
        self.raw = self.raw.replace(&format!("{{{{{}}}}}", old), &format!("{{{{{}}}}}", new));
        self.parts = Self::parse_parts(&self.raw);
    }

    /// Build the final command by injecting variable values
    pub fn build_final_command(&self, values: &HashMap<String, String>) -> String {
        let mut result = String::new();
        for part in &self.parts {
            match part {
                CommandPart::Text(t) => result.push_str(t),
                CommandPart::Variable(v) => result.push_str(values.get(v).map(|s| s.as_str()).unwrap_or("")),
            }
        }
        result
    }

    /// Returns the command string with variable slots as {{name}} (template preview)
    pub fn current(&self) -> String {
        let mut result = String::new();
        for part in &self.parts {
            match part {
                CommandPart::Text(t) => result.push_str(t),
                CommandPart::Variable(v) => {
                    result.push_str(&format!("{{{{{}}}}}", v));
                }
            }
        }
        result
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct GadgetVariable {
    pub name: String,
    pub description: String,
    pub default: Option<String>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Gadget {
    pub name: String,
    pub command: Command,
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
        println!("Command: {}", self.command.raw);
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
    Command::new(command.to_string()).extract_variables()
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

    pub fn from_path(path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let data: String = std::fs::read_to_string(path)?;
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
        let detected_vars = gadget.command.extract_variables();
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
